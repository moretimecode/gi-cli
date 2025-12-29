package gits

import (
	"fmt"
	"gocli/internal/utils"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
	"github.com/go-git/go-git/v5/plumbing/storer"
)

// 约定：统计范围为全仓库（遍历所有 refs 可达提交），且排除 merge commit。
type goGitRepository struct{}

func (r *goGitRepository) LoadCommits(repoPath string) (commits []GitCommitMetrics, err error) {
	// 1. 打开仓库
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("打开仓库失败: %w", err)
	}

	refs, err := repo.References()
	if err != nil {
		return nil, fmt.Errorf("获取引用失败: %w", err)
	}
	defer refs.Close()

	seen := make(map[string]bool, 4096)
	err = refs.ForEach(func(ref *plumbing.Reference) error {
		name := ref.Name().String()
		if !(strings.HasPrefix(name, "refs/heads/") || strings.HasPrefix(name, "refs/remotes/") || strings.HasPrefix(name, "refs/tags/")) {
			return nil
		}

		commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			return nil
		}
		defer commitIter.Close()

		_ = commitIter.ForEach(func(c *object.Commit) error {
			id := c.Hash.String()
			if seen[id] {
				return storer.ErrStop
			}
			seen[id] = true

			if c.NumParents() > 1 {
				return nil
			}

			commitMetrics := GitCommitMetrics{
				CommitID:       id,
				ShortHash:      id[:7],
				Author:         c.Author.Name,
				Email:          c.Author.Email,
				Timestamp:      c.Author.When.Unix(),
				Message:        c.Message,
				ParentCommitID: "",
			}
			if len(c.ParentHashes) > 0 {
				commitMetrics.ParentCommitID = c.ParentHashes[0].String()
			}
			commits = append(commits, commitMetrics)
			return nil
		})

		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("遍历引用失败: %w", err)
	}

	return commits, nil
}

func (r *goGitRepository) LoadCommitStats(repoPath string, commits []GitCommitMetrics) ([]CommitStatsMetrics, error) {
	total := len(commits)
	workerCount := runtime.NumCPU() * 2

	taskChan := make(chan GitCommitMetrics, total)
	resultChan := make(chan CommitStatsMetrics, total)
	errChan := make(chan error, total)
	doneChan := make(chan bool)

	var wg sync.WaitGroup
	for i := range workerCount {
		wg.Add(1)
		go worker(i, repoPath, taskChan, resultChan, errChan, &wg)
	}

	go func() {
		for _, commit := range commits {
			taskChan <- commit
		}
		close(taskChan)
	}()

	go func() {
		wg.Wait()
		close(resultChan)
		close(errChan)
		doneChan <- true
	}()

	var stats []CommitStatsMetrics
	var errors []string

	fmt.Printf("\r⏳ 正在分析提交统计: 0/%d (0%%)", total)

	for {
		select {
		case stat, ok := <-resultChan:
			if ok {
				stats = append(stats, stat)
				// 更新进度条
				processed := len(stats)
				percentage := float64(processed) / float64(total) * 100
				fmt.Printf("\r⏳ 正在分析提交统计: %d/%d (%.1f%%)", processed, total, percentage)
			}
		case err, ok := <-errChan:
			if ok {
				errors = append(errors, err.Error())
			}
		case <-doneChan:
			fmt.Printf("\r✅ 提交统计分析完成: %d/%d (100%%)  \n", total, total)
			if len(errors) > 0 {
				return stats, fmt.Errorf("处理提交统计信息时发生错误: %v", errors)
			}
			return stats, nil
		}
	}
}

func LoadAuthors(commits []GitCommitMetrics, commitStats []CommitStatsMetrics) []AuthorMetrics {
	statsByCommitID := make(map[string]CommitStatsMetrics, len(commitStats))
	for _, stat := range commitStats {
		statsByCommitID[stat.CommitID] = stat
	}

	authorStatsMap := make(map[string]*AuthorMetrics)

	for _, commit := range commits {
		author, exists := authorStatsMap[commit.Email]
		stat, hasStat := statsByCommitID[commit.CommitID]

		if !exists {
			author = &AuthorMetrics{
				Email:        commit.Email,
				Name:         commit.Author,
				FirstSeen:    commit.Timestamp,
				LastSeen:     commit.Timestamp,
				TotalCommits: 1,
			}
			if hasStat {
				author.TotalAdded = stat.LinesAdded
				author.TotalDeleted = stat.LinesDeleted
			}
			authorStatsMap[commit.Email] = author
			continue
		}

		if author.Name != commit.Author && commit.Author != "" {
			author.Name = commit.Author
		}

		if commit.Timestamp < author.FirstSeen {
			author.FirstSeen = commit.Timestamp
		}

		if commit.Timestamp > author.LastSeen {
			author.LastSeen = commit.Timestamp
		}

		author.TotalCommits++
		if hasStat {
			author.TotalAdded += stat.LinesAdded
			author.TotalDeleted += stat.LinesDeleted
		}
	}

	authorStatsSlice := make([]AuthorMetrics, 0, len(authorStatsMap))
	for _, author := range authorStatsMap {
		authorStatsSlice = append(authorStatsSlice, *author)
	}

	return authorStatsSlice
}

type branchAnalysisResult struct {
	branch    BranchMetrics
	commitIDs []string
	err       error
}

func LoadBranches(repoPath string) ([]BranchMetrics, []BranchCommitMetrics, error) {
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, nil, fmt.Errorf("打开仓库失败: %w", err)
	}

	// 获取主分支引用
	mainRef, err := getMainBranch(repo)
	if err != nil {
		return nil, nil, fmt.Errorf("获取主分支失败: %w", err)
	}

	// 构建主分支提交集合（用于判断分支是否已合并）
	fmt.Printf("⏳ 正在分析主分支 [%s] 的提交历史...\n", mainRef.Name().Short())
	mainCommits := make(map[string]bool)
	mainIter, _ := repo.Log(&git.LogOptions{From: mainRef.Hash()})
	mainIter.ForEach(func(c *object.Commit) error {
		mainCommits[c.Hash.String()] = true
		return nil
	})
	mainIter.Close()
	fmt.Printf("✅ 主分支包含 %d 个提交\n\n", len(mainCommits))

	// 收集所有分支引用
	branchIter, err := repo.Branches()
	if err != nil {
		return nil, nil, fmt.Errorf("获取分支列表失败: %w", err)
	}
	defer branchIter.Close()

	var branchRefs []*plumbing.Reference
	err = branchIter.ForEach(func(ref *plumbing.Reference) error {
		branchRefs = append(branchRefs, ref)
		return nil
	})
	if err != nil {
		return nil, nil, fmt.Errorf("收集分支引用失败: %w", err)
	}

	total := len(branchRefs)
	fmt.Printf("⏳ 正在并发分析 %d 个分支...\n", total)

	// 创建任务和结果通道
	taskChan := make(chan *plumbing.Reference, total)
	resultChan := make(chan branchAnalysisResult, total)

	// 启动多个 worker（5个并发，避免打开太多文件）
	var wg sync.WaitGroup
	workerCount := 5
	for range workerCount {
		wg.Add(1)
		go analyzeBranchWorker(repo, mainRef, mainCommits, taskChan, resultChan, &wg)
	}

	// 发送任务
	go func() {
		for _, ref := range branchRefs {
			taskChan <- ref
		}
		close(taskChan)
	}()

	// 等待所有 worker 完成
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// 收集结果并显示进度
	branches := []BranchMetrics{}
	branchCommits := []BranchCommitMetrics{}
	var errors []string
	processed := 0

	for result := range resultChan {
		processed++
		percentage := float64(processed) / float64(total) * 100
		fmt.Printf("\r⏳ 分析进度: %d/%d (%.1f%%)", processed, total, percentage)

		if result.err != nil {
			errors = append(errors, result.err.Error())
			continue
		}

		branches = append(branches, result.branch)

		// 直接生成 BranchCommitMetrics
		for _, commitID := range result.commitIDs {
			branchCommits = append(branchCommits, BranchCommitMetrics{
				BranchName: result.branch.BranchName,
				CommitID:   commitID,
			})
		}
	}

	fmt.Printf("\r✅ 分支分析完成: %d/%d (100%%)  \n", total, total)

	if len(errors) > 0 {
		return branches, branchCommits, fmt.Errorf("部分分支分析失败: %v", errors)
	}

	fmt.Printf("✅ 共分析 %d 个分支，%d 条分支-提交关系\n", len(branches), len(branchCommits))
	return branches, branchCommits, nil
}

// fileAnalysisResult 文件分析结果（内部使用）
type fileAnalysisResult struct {
	metrics []FileMetrics
	err     error
}

func (r *goGitRepository) LoadFileMetrics(repoPath string, commits []GitCommitMetrics) ([]FileMetrics, error) {
	_, err := git.PlainOpen(repoPath)
	if err != nil {
		return nil, fmt.Errorf("打开仓库失败: %w", err)
	}

	total := len(commits)
	workerCount := runtime.NumCPU() * 2

	fmt.Printf("⏳ 正在分析 %d 个提交的文件变更...\n", total)

	taskChan := make(chan GitCommitMetrics, total)
	resultChan := make(chan fileAnalysisResult, total)
	doneChan := make(chan bool)

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go analyzeCommitFilesWorker(repoPath, taskChan, resultChan, &wg)
	}

	// 发送任务
	go func() {
		for _, commit := range commits {
			taskChan <- commit
		}
		close(taskChan)
	}()

	// 等待所有 Worker 完成
	go func() {
		wg.Wait()
		close(resultChan)
		doneChan <- true
	}()

	var allFileMetrics []FileMetrics
	var errors []string
	processed := 0

	// 显示初始进度
	fmt.Printf("\r⏳ 分析进度: 0/%d (0%%)", total)

	for {
		select {
		case result, ok := <-resultChan:
			if ok {
				processed++
				percentage := float64(processed) / float64(total) * 100
				fmt.Printf("\r⏳ 分析进度: %d/%d (%.1f%%)", processed, total, percentage)

				if result.err != nil {
					errors = append(errors, result.err.Error())
				} else {
					allFileMetrics = append(allFileMetrics, result.metrics...)
				}
			}
		case <-doneChan:
			fmt.Printf("\r✅ 文件分析完成: %d/%d (100%%)  \n", total, total)
			if len(errors) > 0 {
				fmt.Printf("⚠️  警告: %d 个提交分析失败，已跳过\n", len(errors))
				// 显示前3个错误详情
				for i := 0; i < len(errors) && i < 3; i++ {
					fmt.Printf("  - %s\n", errors[i])
				}
				if len(errors) > 3 {
					fmt.Printf("  ... 还有 %d 个错误未显示\n", len(errors)-3)
				}
			}
			fmt.Printf("✅ 共分析 %d 个文件变更\n", len(allFileMetrics))
			return allFileMetrics, nil
		}
	}
}

// analyzeCommitFilesWorker 分析单个提交的文件变更（并发安全）

func analyzeCommitFilesWorker(
	repoPath string,
	tasks <-chan GitCommitMetrics,
	results chan<- fileAnalysisResult,
	wg *sync.WaitGroup) {

	defer wg.Done()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		results <- fileAnalysisResult{err: fmt.Errorf("打开仓库失败: %w", err)}
		return
	}

	for commitMeta := range tasks {
		commit, err := repo.CommitObject(plumbing.NewHash(commitMeta.CommitID))
		if err != nil {
			// 提交对象不存在，记录错误
			results <- fileAnalysisResult{err: fmt.Errorf("提交 %s: 获取对象失败", commitMeta.ShortHash)}
			continue
		}

		// 跳过初始提交
		if commit.NumParents() == 0 {
			results <- fileAnalysisResult{metrics: []FileMetrics{}, err: nil}
			continue
		}

		parent, err := commit.Parent(0)
		if err != nil {
			// 父提交不存在，记录错误
			results <- fileAnalysisResult{err: fmt.Errorf("提交 %s: 获取父提交失败", commitMeta.ShortHash)}
			continue
		}

		patch, err := parent.Patch(commit)
		if err != nil {
			// 无法生成补丁，静默跳过（这通常是正常的，比如只有二进制文件变更）
			results <- fileAnalysisResult{metrics: []FileMetrics{}, err: nil}
			continue
		}

		// 获取整个提交的文件统计
		commitStats, err := commit.Stats()
		if err != nil {
			// 无法获取统计，静默跳过
			results <- fileAnalysisResult{metrics: []FileMetrics{}, err: nil}
			continue
		}

		// 建立文件路径到统计的映射
		fileStatsMap := make(map[string]object.FileStat)
		for _, fs := range commitStats {
			fileStatsMap[fs.Name] = fs
		}

		var fileMetrics []FileMetrics

		for _, filePatch := range patch.FilePatches() {
			from, to := filePatch.Files()
			var filePath, fileName, changeType string

			if to != nil {
				filePath = to.Path()
				fileName = filepath.Base(filePath)
			} else if from != nil {
				filePath = from.Path()
				fileName = filepath.Base(filePath)
			}

			// 判断变更类型
			if from == nil && to != nil {
				changeType = "added"
			} else if from != nil && to == nil {
				changeType = "deleted"
			} else if from != nil && to != nil {
				if from.Path() != to.Path() {
					changeType = "renamed"
				} else {
					changeType = "modified"
				}
			}

			// 从 map 中查找统计信息
			linesModified := 0
			if stat, ok := fileStatsMap[filePath]; ok {
				linesModified = stat.Addition + stat.Deletion
			}

			// 获取提交后的总行数
			totalLines := 0
			if to != nil && changeType != "deleted" {
				tree, err := commit.Tree()
				if err == nil {
					file, err := tree.File(to.Path())
					if err == nil {
						reader, err := file.Reader()
						if err == nil {
							totalLines, _ = utils.CountLinesFromReader(reader)
							reader.Close()
						}
					}
				}
			}

			fileMetric := FileMetrics{
				CommitID:      commitMeta.CommitID,
				FileName:      fileName,
				FilePath:      filePath,
				LinesModified: linesModified,
				TotalLines:    totalLines,
				ChangeType:    changeType,
			}

			fileMetrics = append(fileMetrics, fileMetric)
		}

		results <- fileAnalysisResult{metrics: fileMetrics, err: nil}
	}
}

// analyzeBranchWorker 分析单个分支的 worker（并发安全）
func analyzeBranchWorker(
	repo *git.Repository,
	mainRef *plumbing.Reference,
	mainCommits map[string]bool, // 只读，并发安全
	tasks <-chan *plumbing.Reference,
	results chan<- branchAnalysisResult,
	wg *sync.WaitGroup) {

	defer wg.Done()

	for ref := range tasks {
		branchName := ref.Name().Short()

		// 遍历该分支的所有提交
		allInMain := true
		var commitIDs []string

		commitIter, err := repo.Log(&git.LogOptions{From: ref.Hash()})
		if err != nil {
			results <- branchAnalysisResult{
				err: fmt.Errorf("分支 %s: 获取提交历史失败: %w", branchName, err),
			}
			continue
		}

		commitIter.ForEach(func(c *object.Commit) error {
			commitID := c.Hash.String()
			commitIDs = append(commitIDs, commitID)

			// 检查是否在主分支中
			if !mainCommits[commitID] {
				allInMain = false
			}
			return nil
		})
		commitIter.Close()

		// 判断是否已合并
		isMerged := false
		mergedInto := ""
		if allInMain && branchName != mainRef.Name().Short() {
			isMerged = true
			mergedInto = mainRef.Name().Short()
		}

		// 构建分支信息
		branch := BranchMetrics{
			BranchName:   branchName,
			IsDefault:    branchName == mainRef.Name().Short(),
			LastCommitID: ref.Hash().String(),
			IsMerged:     isMerged,
			MergedAt:     0, // 跳过
			MergedInto:   mergedInto,
		}

		results <- branchAnalysisResult{
			branch:    branch,
			commitIDs: commitIDs,
			err:       nil,
		}
	}
}

func worker(id int, repoPath string, tasks <-chan GitCommitMetrics,
	results chan<- CommitStatsMetrics, errors chan<- error, wg *sync.WaitGroup) {
	defer wg.Done()

	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		errors <- fmt.Errorf("Worker %d: 打开仓库失败: %v", id, err)
		// 无法打开仓库，直接退出该 worker
		return
	}

	for commitMeta := range tasks {
		stat, err := processSingleCommit(repo, commitMeta)
		if err != nil {
			errors <- fmt.Errorf("Worker %d: %v", id, err)
			continue
		}
		results <- stat
	}
}

func processSingleCommit(repo *git.Repository, commitMeta GitCommitMetrics) (CommitStatsMetrics, error) {
	commit, err := repo.CommitObject(plumbing.NewHash(commitMeta.CommitID))
	if err != nil {
		return CommitStatsMetrics{}, fmt.Errorf("获取提交对象失败: %w", err)
	}

	fileStats, err := commit.Stats()
	if err != nil {
		return CommitStatsMetrics{}, fmt.Errorf("获取提交统计信息失败: %w", err)
	}

	stat := CommitStatsMetrics{
		CommitID: commitMeta.CommitID,
		IsMerge:  commit.NumParents() > 1,
	}

	for _, fs := range fileStats {
		stat.FilesChanged++
		stat.LinesAdded += fs.Addition
		stat.LinesDeleted += fs.Deletion
		stat.LinesModified += fs.Addition + fs.Deletion

		if fs.Addition > 0 && fs.Deletion == 0 {
			stat.FileAddCount++
		} else if fs.Deletion > 0 && fs.Addition == 0 {
			stat.FileDelCount++
		}
	}
	return stat, nil
}

func getMainBranch(repo *git.Repository) (*plumbing.Reference, error) {
	// 按顺序尝试常见的主分支名称
	tryNames := []string{
		"refs/heads/main",
		"refs/heads/master",
		"refs/remotes/origin/main",
		"refs/remotes/origin/master",
		"origin/main",
		"origin/master",
	}

	for _, name := range tryNames {
		ref, err := repo.Reference(plumbing.ReferenceName(name), true)
		if err == nil && ref != nil {
			return ref, nil
		}
	}

	return nil, fmt.Errorf("未找到主分支")
}

func countCommits(commit *object.Commit) (int, error) {
	count := 0

	commitIter := object.NewCommitIterCTime(commit, nil, nil)
	err := commitIter.ForEach(func(c *object.Commit) error {
		count++
		return nil
	})

	return count, err
}

func isSameBranch(branch1, branch2 *plumbing.Reference) bool {
	return branch1.Name() == branch2.Name()
}
