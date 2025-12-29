package gits

import (
	"fmt"
	"gocli/internal/utils"
	"sort"
	"strings"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// ============ 排行榜相关 ============

// ContributorRank 贡献者排行榜项
type ContributorRank struct {
	AuthorName   string
	TotalCommits int
	TotalAdded   int
	TotalDeleted int
	TotalChanged int
	Rank         int
}

func GetTopContributors(authors []AuthorMetrics, top int) []ContributorRank {
	sorted := make([]AuthorMetrics, len(authors))
	copy(sorted, authors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalCommits > sorted[j].TotalCommits
	})

	if top > 0 && top < len(sorted) {
		sorted = sorted[:top]
	}

	return utils.Select(sorted, func(author AuthorMetrics) ContributorRank {
		return ContributorRank{
			AuthorName:   author.Name,
			TotalCommits: author.TotalCommits,
			TotalAdded:   author.TotalAdded,
			TotalDeleted: author.TotalDeleted,
			TotalChanged: author.TotalAdded + author.TotalDeleted,
			Rank:         0,
		}
	})
}

// GetTopCodeWriters 获取代码编写排行榜（按新增行数）
func GetTopCodeWriters(authors []AuthorMetrics, top int) []ContributorRank {
	sorted := make([]AuthorMetrics, len(authors))
	copy(sorted, authors)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalAdded > sorted[j].TotalAdded
	})

	if top > 0 && top < len(sorted) {
		sorted = sorted[:top]
	}

	ranks := utils.Select(sorted, func(author AuthorMetrics) ContributorRank {
		return ContributorRank{
			AuthorName:   author.Name,
			TotalCommits: author.TotalCommits,
			TotalAdded:   author.TotalAdded,
			TotalDeleted: author.TotalDeleted,
			TotalChanged: author.TotalAdded + author.TotalDeleted,
		}
	})

	// 添加排名
	for i := range ranks {
		ranks[i].Rank = i + 1
	}

	return ranks
}

// ============ 时间分析相关 ============

// TimeSlot 时间段
type TimeSlot struct {
	Hour    int
	Commits int
}

// GetCommitsByHour 按小时统计提交分布
func GetCommitsByHour(commits []GitCommitMetrics) []TimeSlot {
	hourMap := make(map[int]int)

	for _, commit := range commits {
		hour := commit.ToLocalTime().Hour()
		hourMap[hour]++
	}

	slots := make([]TimeSlot, 0, 24)
	for hour := 0; hour < 24; hour++ {
		slots = append(slots, TimeSlot{
			Hour:    hour,
			Commits: hourMap[hour],
		})
	}

	return slots
}

// NightOwlStats 夜猫子统计
type NightOwlStats struct {
	AuthorName   string
	NightCommits int     // 22:00-06:00 的提交
	DayCommits   int     // 06:00-22:00 的提交
	NightRatio   float64 // 夜间提交占比
}

// FindNightOwls 找出夜猫子程序员（只要有 22:00-次日06:00 的提交就算）
func FindNightOwls(commits []GitCommitMetrics, threshold float64) []NightOwlStats {
	// 按作者分组统计
	authorMap := make(map[string]*NightOwlStats)

	for _, commit := range commits {
		hour := commit.ToLocalTime().Hour()

		if _, exists := authorMap[commit.Author]; !exists {
			authorMap[commit.Author] = &NightOwlStats{
				AuthorName: commit.Author,
			}
		}

		if hour >= 22 || hour < 6 {
			authorMap[commit.Author].NightCommits++
		} else {
			authorMap[commit.Author].DayCommits++
		}
	}

	// 计算占比，只要有夜间提交就加入结果
	results := make([]NightOwlStats, 0)
	for _, stats := range authorMap {
		total := stats.NightCommits + stats.DayCommits
		if total > 0 {
			stats.NightRatio = float64(stats.NightCommits) / float64(total)
			// 只要有夜间提交就算夜猫子
			if stats.NightCommits > 0 {
				results = append(results, *stats)
			}
		}
	}

	// 按夜间提交数排序（夜间提交越多越靠前）
	sort.Slice(results, func(i, j int) bool {
		return results[i].NightCommits > results[j].NightCommits
	})

	return results
}

// ============ 活跃度分析 ============

// ActivityPeriod 活跃时期
type ActivityPeriod struct {
	Year    int
	Month   int
	Commits int
	Authors []string
}

// GetMonthlyActivity 按月统计活跃度
func GetMonthlyActivity(commits []GitCommitMetrics) []ActivityPeriod {
	type YearMonth struct {
		Year  int
		Month int
	}

	periodMap := make(map[YearMonth]map[string]bool)

	for _, commit := range commits {
		t := commit.ToLocalTime()
		ym := YearMonth{Year: t.Year(), Month: int(t.Month())}

		if _, exists := periodMap[ym]; !exists {
			periodMap[ym] = make(map[string]bool)
		}
		periodMap[ym][commit.Author] = true
	}

	// 转换为切片
	results := make([]ActivityPeriod, 0, len(periodMap))
	for ym, authors := range periodMap {
		authorList := make([]string, 0, len(authors))
		for author := range authors {
			authorList = append(authorList, author)
		}

		results = append(results, ActivityPeriod{
			Year:    ym.Year,
			Month:   ym.Month,
			Commits: len(authors),
			Authors: authorList,
		})
	}

	// 按时间排序
	sort.Slice(results, func(i, j int) bool {
		if results[i].Year != results[j].Year {
			return results[i].Year < results[j].Year
		}
		return results[i].Month < results[j].Month
	})

	return results
}

// AuthorCommitPattern 作者提交模式
type AuthorCommitPattern struct {
	AuthorName      string
	TotalCommits    int
	HourlyCommits   [24]int          // 24小时提交分布
	WeekdayCommits  [7]int           // 周一到周日提交分布
	MonthlyActivity []ActivityPeriod // 按月活跃度
}

// GetAuthorCommitPattern 获取指定作者的提交模式
func GetAuthorCommitPattern(commits []GitCommitMetrics, authorName string) *AuthorCommitPattern {
	pattern := &AuthorCommitPattern{
		AuthorName: authorName,
	}

	// 按月统计
	type YearMonth struct {
		Year  int
		Month int
	}
	monthMap := make(map[YearMonth]int)

	for _, commit := range commits {
		// 支持模糊匹配
		if !strings.Contains(commit.Author, authorName) && !strings.Contains(commit.Email, authorName) {
			continue
		}

		pattern.TotalCommits++
		t := commit.ToLocalTime()

		// 小时分布
		pattern.HourlyCommits[t.Hour()]++

		// 星期分布 (0=Sunday, 1=Monday, ...)
		pattern.WeekdayCommits[t.Weekday()]++

		// 月度分布
		ym := YearMonth{Year: t.Year(), Month: int(t.Month())}
		monthMap[ym]++
	}

	// 转换月度数据
	for ym, count := range monthMap {
		pattern.MonthlyActivity = append(pattern.MonthlyActivity, ActivityPeriod{
			Year:    ym.Year,
			Month:   ym.Month,
			Commits: count,
			Authors: []string{authorName},
		})
	}

	// 按时间排序
	sort.Slice(pattern.MonthlyActivity, func(i, j int) bool {
		if pattern.MonthlyActivity[i].Year != pattern.MonthlyActivity[j].Year {
			return pattern.MonthlyActivity[i].Year < pattern.MonthlyActivity[j].Year
		}
		return pattern.MonthlyActivity[i].Month < pattern.MonthlyActivity[j].Month
	})

	return pattern
}

// ============ 文件热点分析 ============

// HotFile 热点文件
type HotFile struct {
	FilePath      string
	ChangeCount   int
	TotalModified int
	Contributors  []string
}

// GetHotFiles 获取最频繁修改的文件
func GetHotFiles(fileMetrics []FileMetrics, repoPath string, top int) []HotFile {
	// 创建 gitignore 匹配器
	matcher := utils.NewPathIgnoreMatcher(repoPath, []string{})

	// 打开仓库并获取当前 HEAD 的树
	repo, err := git.PlainOpen(repoPath)
	if err != nil {
		// 如果无法打开仓库，就不过滤已删除文件
		return getHotFilesWithoutCheck(fileMetrics, matcher, top)
	}

	head, err := repo.Head()
	if err != nil {
		return getHotFilesWithoutCheck(fileMetrics, matcher, top)
	}

	commit, err := repo.CommitObject(head.Hash())
	if err != nil {
		return getHotFilesWithoutCheck(fileMetrics, matcher, top)
	}

	tree, err := commit.Tree()
	if err != nil {
		return getHotFilesWithoutCheck(fileMetrics, matcher, top)
	}

	fileMap := make(map[string]*HotFile)

	for _, fm := range fileMetrics {
		// 使用 gitignore 规则过滤文件
		absPath := repoPath + "/" + fm.FilePath
		if matcher.Match(absPath) {
			continue
		}

		// 检查文件是否存在于当前树中
		_, err := tree.File(fm.FilePath)
		if err != nil {
			// 文件已被删除，跳过
			continue
		}

		if _, exists := fileMap[fm.FilePath]; !exists {
			fileMap[fm.FilePath] = &HotFile{
				FilePath:     fm.FilePath,
				Contributors: make([]string, 0),
			}
		}

		fileMap[fm.FilePath].ChangeCount++
		fileMap[fm.FilePath].TotalModified += fm.LinesModified
	}

	// 转换为结果
	results := make([]HotFile, 0, len(fileMap))
	for _, hotFile := range fileMap {
		results = append(results, *hotFile)
	}

	// 按修改次数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].ChangeCount > results[j].ChangeCount
	})

	if top > 0 && top < len(results) {
		results = results[:top]
	}

	return results
}

// getHotFilesWithoutCheck 不检查文件是否存在的后备逻辑
func getHotFilesWithoutCheck(fileMetrics []FileMetrics, matcher utils.PathIgnoreMatcher, top int) []HotFile {
	fileMap := make(map[string]*HotFile)

	for _, fm := range fileMetrics {
		if _, exists := fileMap[fm.FilePath]; !exists {
			fileMap[fm.FilePath] = &HotFile{
				FilePath:     fm.FilePath,
				Contributors: make([]string, 0),
			}
		}

		fileMap[fm.FilePath].ChangeCount++
		fileMap[fm.FilePath].TotalModified += fm.LinesModified
	}

	// 转换为结果
	results := make([]HotFile, 0, len(fileMap))
	for _, hotFile := range fileMap {
		results = append(results, *hotFile)
	}

	// 按修改次数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].ChangeCount > results[j].ChangeCount
	})

	if top > 0 && top < len(results) {
		results = results[:top]
	}

	return results
}

// GetAuthorHotFiles 获取特定作者修改频率最高的文件
func GetAuthorHotFiles(fileMetrics []FileMetrics, commits []GitCommitMetrics, authorName string, top int) []HotFile {
	// 建立 CommitID -> Author 映射
	commitAuthorMap := make(map[string]string)
	for _, c := range commits {
		commitAuthorMap[c.CommitID] = c.Author
	}

	// 筛选出该作者的文件变更记录
	authorFileMetrics := utils.Where(fileMetrics, func(fm FileMetrics) bool {
		author, ok := commitAuthorMap[fm.CommitID]
		if !ok {
			return false
		}
		// 支持模糊匹配
		return strings.Contains(author, authorName)
	})

	if len(authorFileMetrics) == 0 {
		return []HotFile{}
	}

	// 统计每个文件的修改次数
	fileMap := make(map[string]*HotFile)
	for _, fm := range authorFileMetrics {
		if _, exists := fileMap[fm.FilePath]; !exists {
			fileMap[fm.FilePath] = &HotFile{
				FilePath:     fm.FilePath,
				Contributors: []string{authorName},
			}
		}
		fileMap[fm.FilePath].ChangeCount++
		fileMap[fm.FilePath].TotalModified += fm.LinesModified
	}

	// 转换为结果
	results := make([]HotFile, 0, len(fileMap))
	for _, hotFile := range fileMap {
		results = append(results, *hotFile)
	}

	// 按修改次数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].ChangeCount > results[j].ChangeCount
	})

	if top > 0 && top < len(results) {
		results = results[:top]
	}

	return results
}

// ============ 代码稳定性分析 ============

// CodeStability 代码稳定性指标
type CodeStability struct {
	FilePath            string
	DaysSinceLastChange int
	ChangeFrequency     float64 // 平均多少天修改一次
	IsStable            bool    // 是否稳定（超过90天未修改）
}

// AnalyzeCodeStability 分析代码稳定性
func AnalyzeCodeStability(fileMetrics []FileMetrics, commits []GitCommitMetrics, repoPath string) []CodeStability {
	now := time.Now()

	// 创建 gitignore 匹配器
	matcher := utils.NewPathIgnoreMatcher(repoPath, []string{})

	// 打开仓库并获取当前 HEAD 的树
	repo, err := git.PlainOpen(repoPath)
	var tree *object.Tree
	if err == nil {
		head, err := repo.Head()
		if err == nil {
			commit, err := repo.CommitObject(head.Hash())
			if err == nil {
				tree, _ = commit.Tree()
			}
		}
	}

	// 建立 CommitID -> Timestamp 的映射
	commitTimeMap := make(map[string]int64)
	for _, c := range commits {
		commitTimeMap[c.CommitID] = c.Timestamp
	}

	// 按文件分组
	fileMap := make(map[string][]int64) // filePath -> []timestamps

	for _, fm := range fileMetrics {
		// 使用 gitignore 规则过滤文件
		absPath := repoPath + "/" + fm.FilePath
		if matcher.Match(absPath) {
			continue
		}

		// 检查文件是否存在于当前树中
		if tree != nil {
			_, err := tree.File(fm.FilePath)
			if err != nil {
				// 文件已被删除，跳过
				continue
			}
		}

		if timestamp, ok := commitTimeMap[fm.CommitID]; ok {
			fileMap[fm.FilePath] = append(fileMap[fm.FilePath], timestamp)
		}
	}

	results := make([]CodeStability, 0, len(fileMap))

	for filePath, timestamps := range fileMap {
		if len(timestamps) == 0 {
			continue
		}

		// 排序找到最后修改时间
		sort.Slice(timestamps, func(i, j int) bool {
			return timestamps[i] > timestamps[j]
		})

		lastChange := time.Unix(timestamps[0], 0)
		firstChange := time.Unix(timestamps[len(timestamps)-1], 0)

		daysSinceLast := int(now.Sub(lastChange).Hours() / 24)
		totalDays := int(lastChange.Sub(firstChange).Hours() / 24)

		frequency := 0.0
		if totalDays > 0 && len(timestamps) > 1 {
			frequency = float64(totalDays) / float64(len(timestamps))
		}

		results = append(results, CodeStability{
			FilePath:            filePath,
			DaysSinceLastChange: daysSinceLast,
			ChangeFrequency:     frequency,
			IsStable:            daysSinceLast > 90,
		})
	}

	// 按最后修改时间排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].DaysSinceLastChange > results[j].DaysSinceLastChange
	})

	return results
}

// ============ 使用 LINQ 风格的高级查询 ============

// GetActiveAuthors 获取活跃的作者（提交数超过阈值）
func GetActiveAuthors(authors []AuthorMetrics, minCommits int) []AuthorMetrics {
	return utils.Where(authors, func(a AuthorMetrics) bool {
		return a.TotalCommits >= minCommits
	})
}

// GetRecentCommits 获取最近 N 天的提交
func GetRecentCommits(commits []GitCommitMetrics, days int) []GitCommitMetrics {
	cutoff := time.Now().AddDate(0, 0, -days).Unix()
	return utils.Where(commits, func(c GitCommitMetrics) bool {
		return c.Timestamp >= cutoff
	})
}

// GetFilesByType 按文件扩展名筛选
func GetFilesByType(fileMetrics []FileMetrics, extensions []string) []FileMetrics {
	return utils.Where(fileMetrics, func(fm FileMetrics) bool {
		for _, ext := range extensions {
			if len(fm.FileName) > len(ext) &&
				fm.FileName[len(fm.FileName)-len(ext):] == ext {
				return true
			}
		}
		return false
	})
}

// GetLargestCommits 获取改动最大的提交
func GetLargestCommits(commitStats []CommitStatsMetrics, top int) []CommitStatsMetrics {
	sorted := make([]CommitStatsMetrics, len(commitStats))
	copy(sorted, commitStats)

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].TotalLines > sorted[j].TotalLines
	})

	if top > 0 && top < len(sorted) {
		sorted = sorted[:top]
	}

	return sorted
}

// ============ 协作分析 ============

// CollaborationPair 协作对
type CollaborationPair struct {
	Author1     string
	Author2     string
	SharedFiles int
	Score       float64
}

// FindCollaborations 找出协作密切的开发者对（修改相同文件）
func FindCollaborations(fileMetrics []FileMetrics, commits []GitCommitMetrics) []CollaborationPair {
	// 建立 CommitID -> Author 映射
	commitAuthorMap := make(map[string]string)
	for _, c := range commits {
		commitAuthorMap[c.CommitID] = c.Author
	}

	// 筛选出有作者信息的文件变更记录
	validFileMetrics := utils.Where(fileMetrics, func(fm FileMetrics) bool {
		_, ok := commitAuthorMap[fm.CommitID]
		return ok
	})

	// 建立 FilePath -> Authors 映射
	fileAuthorsMap := make(map[string]map[string]bool)
	for _, fm := range validFileMetrics {
		if _, exists := fileAuthorsMap[fm.FilePath]; !exists {
			fileAuthorsMap[fm.FilePath] = make(map[string]bool)
		}
		fileAuthorsMap[fm.FilePath][commitAuthorMap[fm.CommitID]] = true
	}

	// 将 fileAuthorsMap 转换为结构化数据
	type fileAuthors struct {
		authors map[string]bool
	}

	fileAuthorsList := make([]fileAuthors, 0, len(fileAuthorsMap))
	for _, authors := range fileAuthorsMap {
		fileAuthorsList = append(fileAuthorsList, fileAuthors{authors: authors})
	}

	// 使用 Select 将每个文件的作者 map 转换为作者列表
	authorLists := utils.Select(fileAuthorsList, func(fa fileAuthors) []string {
		result := make([]string, 0, len(fa.authors))
		for author := range fa.authors {
			result = append(result, author)
		}
		return result
	})

	// 筛选出有多个作者的文件
	validAuthorLists := utils.Where(authorLists, func(list []string) bool {
		return len(list) >= 2
	})

	// 定义配对结构
	type authorPair struct {
		key string
	}

	// 使用 SelectMany 将每个作者列表转换为配对列表，然后展平
	allPairs := utils.SelectMany(validAuthorLists, func(authorList []string) []authorPair {
		pairs := make([]authorPair, 0)
		for i := range authorList {
			for j := i + 1; j < len(authorList); j++ {
				a1, a2 := authorList[i], authorList[j]
				if a1 > a2 {
					a1, a2 = a2, a1
				}
				pairs = append(pairs, authorPair{
					key: fmt.Sprintf("%s|%s", a1, a2),
				})
			}
		}
		return pairs
	})

	// 统计每个配对出现的次数
	pairMap := make(map[string]int)
	for _, pair := range allPairs {
		pairMap[pair.key]++
	}

	// 使用 Select 转换为结果（过滤掉无效的键）
	type pairEntry struct {
		key   string
		count int
	}

	pairEntries := make([]pairEntry, 0, len(pairMap))
	for key, count := range pairMap {
		pairEntries = append(pairEntries, pairEntry{key: key, count: count})
	}

	results := utils.Select(utils.Where(pairEntries, func(entry pairEntry) bool {
		return len(strings.Split(entry.key, "|")) == 2
	}), func(entry pairEntry) CollaborationPair {
		parts := strings.Split(entry.key, "|")
		return CollaborationPair{
			Author1:     parts[0],
			Author2:     parts[1],
			SharedFiles: entry.count,
			Score:       float64(entry.count),
		}
	})

	// 按共享文件数排序
	sort.Slice(results, func(i, j int) bool {
		return results[i].SharedFiles > results[j].SharedFiles
	})

	return results
}

// AuthorFileDetail 作者文件详情
type AuthorFileDetail struct {
	FilePath      string
	AuthorCommits map[string]int // 每个作者的提交次数
	TotalCommits  int            // 该文件总提交次数
}

// FindCollaborationDetail 查找多个作者的详细协作信息
func FindCollaborationDetail(fileMetrics []FileMetrics, commits []GitCommitMetrics, authors ...string) []AuthorFileDetail {
	if len(authors) < 2 {
		return []AuthorFileDetail{}
	}

	// 建立 CommitID -> Author 映射
	commitAuthorMap := make(map[string]string)
	for _, c := range commits {
		commitAuthorMap[c.CommitID] = c.Author
	}

	// 筛选出有作者信息的文件变更记录
	validFileMetrics := utils.Where(fileMetrics, func(fm FileMetrics) bool {
		_, ok := commitAuthorMap[fm.CommitID]
		return ok
	})

	// 建立 FilePath -> Author -> CommitCount 映射
	fileAuthorCommits := make(map[string]map[string]int)
	for _, fm := range validFileMetrics {
		author := commitAuthorMap[fm.CommitID]
		if _, exists := fileAuthorCommits[fm.FilePath]; !exists {
			fileAuthorCommits[fm.FilePath] = make(map[string]int)
		}
		fileAuthorCommits[fm.FilePath][author]++
	}

	// 使用 Where 和 Select 筛选并转换结果
	type fileCommitEntry struct {
		filePath      string
		authorCommits map[string]int
	}

	fileEntries := make([]fileCommitEntry, 0, len(fileAuthorCommits))
	for filePath, authorCommits := range fileAuthorCommits {
		fileEntries = append(fileEntries, fileCommitEntry{
			filePath:      filePath,
			authorCommits: authorCommits,
		})
	}

	// 筛选出所有指定作者都参与的文件
	collabFiles := utils.Where(fileEntries, func(entry fileCommitEntry) bool {
		for _, author := range authors {
			if entry.authorCommits[author] == 0 {
				return false
			}
		}
		return true
	})

	// 转换为 AuthorFileDetail
	results := utils.Select(collabFiles, func(entry fileCommitEntry) AuthorFileDetail {
		totalCommits := 0
		for _, count := range entry.authorCommits {
			totalCommits += count
		}

		// 只保留指定作者的提交数
		authorCommits := make(map[string]int)
		for _, author := range authors {
			authorCommits[author] = entry.authorCommits[author]
		}

		return AuthorFileDetail{
			FilePath:      entry.filePath,
			AuthorCommits: authorCommits,
			TotalCommits:  totalCommits,
		}
	})

	// 按总提交次数排序（最热门的协作文件在前）
	sort.Slice(results, func(i, j int) bool {
		return results[i].TotalCommits > results[j].TotalCommits
	})

	return results
}
