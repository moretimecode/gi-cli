package gits

import (
	"bufio"
	"bytes"
	"fmt"
	"os/exec"
	"path"
	"regexp"
	"strconv"
	"strings"
)

// nativeGitRepository 使用系统 git 命令实现（通常更快）。
// 约定：全仓库（--all）且排除 merge（--no-merges）。
type nativeGitRepository struct{}

func (r *nativeGitRepository) LoadCommits(repoPath string) ([]GitCommitMetrics, error) {
	// 使用记录分隔符(0x1e)与字段分隔符(0x1f)来增强解析健壮性。
	// 字段：H, h, an, ae, at, P, s
	format := "%H%x1f%h%x1f%an%x1f%ae%x1f%at%x1f%P%x1f%s%x1e"
	cmd := exec.Command(
		"git",
		"-C", repoPath,
		"log",
		"--all",
		"--no-merges",
		"--pretty=format:"+format,
	)

	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("执行 git log 失败: %w", err)
	}

	records := bytes.Split(out, []byte{0x1e})
	commits := make([]GitCommitMetrics, 0, len(records))

	for _, rec := range records {
		if len(rec) == 0 {
			continue
		}
		fields := bytes.Split(rec, []byte{0x1f})
		if len(fields) < 7 {
			continue
		}

		commitID := string(fields[0])
		shortHash := string(fields[1])
		author := string(fields[2])
		email := string(fields[3])
		tsStr := string(fields[4])
		parentsStr := string(fields[5])
		message := string(fields[6])

		// 清理空白字符
		commitID = strings.TrimSpace(commitID)
		shortHash = strings.TrimSpace(shortHash)
		author = strings.TrimSpace(author)
		email = strings.TrimSpace(email)

		ts, _ := strconv.ParseInt(strings.TrimSpace(tsStr), 10, 64)
		parentCommitID := ""
		if p := strings.Fields(parentsStr); len(p) > 0 {
			parentCommitID = p[0]
		}

		commits = append(commits, GitCommitMetrics{
			CommitID:       commitID,
			ShortHash:      shortHash,
			Author:         author,
			Email:          email,
			Timestamp:      ts,
			Message:        message,
			ParentCommitID: parentCommitID,
		})
	}

	return commits, nil
}

func (r *nativeGitRepository) LoadCommitStats(repoPath string, commits []GitCommitMetrics) ([]CommitStatsMetrics, error) {
	cmd := exec.Command(
		"git",
		"-C", repoPath,
		"log",
		"--all",
		"--no-merges",
		"--pretty=format:COMMIT:%H",
		"--numstat",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("获取 git 输出失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("获取 git 错误输出失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("执行 git 失败: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	var current CommitStatsMetrics
	hasCurrent := false
	stats := make([]CommitStatsMetrics, 0, 1024)

	flush := func() {
		if !hasCurrent {
			return
		}
		current.LinesModified = current.LinesAdded + current.LinesDeleted
		stats = append(stats, current)
		hasCurrent = false
		current = CommitStatsMetrics{}
	}

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "COMMIT:") {
			flush()
			current = CommitStatsMetrics{
				CommitID: strings.TrimSpace(strings.TrimPrefix(line, "COMMIT:")),
				IsMerge:  false,
			}
			hasCurrent = current.CommitID != ""
			continue
		}
		if !hasCurrent || line == "" {
			continue
		}
		parts := strings.Split(line, "\t")
		if !isNumStatLine(parts) {
			continue
		}

		added := parseNumStatInt(parts[0])
		deleted := parseNumStatInt(parts[1])

		current.FilesChanged++
		current.LinesAdded += added
		current.LinesDeleted += deleted

		if added > 0 && deleted == 0 {
			current.FileAddCount++
		} else if deleted > 0 && added == 0 {
			current.FileDelCount++
		}
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("读取 git 输出失败: %w", err)
	}

	flush()

	errBytes, _ := ioReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		msg := strings.TrimSpace(string(errBytes))
		if msg != "" {
			return nil, fmt.Errorf("git 执行失败: %s", msg)
		}
		return nil, fmt.Errorf("git 执行失败: %w", err)
	}

	return stats, nil
}

func (r *nativeGitRepository) LoadFileMetrics(repoPath string, commits []GitCommitMetrics) ([]FileMetrics, error) {
	// commits 参数保留用于与 GoGitRepository 对齐；native 实现直接以 --all/--no-merges 为准。
	cmd := exec.Command(
		"git",
		"-C", repoPath,
		"log",
		"--all",
		"--no-merges",
		"-M",
		"--pretty=format:COMMIT:%H",
		"--numstat",
		"--name-status",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("获取 git 输出失败: %w", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, fmt.Errorf("获取 git 错误输出失败: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("执行 git 失败: %w", err)
	}

	scanner := bufio.NewScanner(stdout)
	buf := make([]byte, 0, 64*1024)
	scanner.Buffer(buf, 4*1024*1024)

	var currentCommitID string
	statusByPath := make(map[string]string)
	linesByPath := make(map[string]int)

	flushCommit := func(all *[]FileMetrics) {
		if currentCommitID == "" {
			return
		}
		for p := range linesByPath {
			if _, ok := statusByPath[p]; !ok {
				statusByPath[p] = "modified"
			}
		}
		for p, ct := range statusByPath {
			(*all) = append((*all), FileMetrics{
				CommitID:      currentCommitID,
				FileName:      path.Base(p),
				FilePath:      p,
				LinesModified: linesByPath[p],
				TotalLines:    0, // 原生 git 版本不做“提交后文件总行数”计算（成本很高）
				ChangeType:    ct,
			})
		}
		// reset
		for k := range statusByPath {
			delete(statusByPath, k)
		}
		for k := range linesByPath {
			delete(linesByPath, k)
		}
	}

	all := make([]FileMetrics, 0, 1024)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "COMMIT:") {
			flushCommit(&all)
			currentCommitID = strings.TrimSpace(strings.TrimPrefix(line, "COMMIT:"))
			continue
		}
		if line == "" || currentCommitID == "" {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			continue
		}

		// numstat: <add>\t<del>\t<path>
		if isNumStatLine(parts) {
			added := parseNumStatInt(parts[0])
			deleted := parseNumStatInt(parts[1])
			p := ""
			if len(parts) >= 3 {
				p = normalizeGitRenamePath(parts[2])
			}
			if p != "" {
				linesByPath[p] += added + deleted
			}
			continue
		}

		// name-status: A/M/D/T/Rxxx/Cxxx ...
		code := strings.TrimSpace(parts[0])
		if code == "" {
			continue
		}

		switch {
		case strings.HasPrefix(code, "R") && len(parts) >= 3:
			newPath := normalizeGitRenamePath(parts[2])
			if newPath != "" {
				statusByPath[newPath] = "renamed"
			}
		case strings.HasPrefix(code, "C") && len(parts) >= 3:
			newPath := normalizeGitRenamePath(parts[2])
			if newPath != "" {
				statusByPath[newPath] = "added"
			}
		default:
			p := normalizeGitRenamePath(parts[1])
			if p == "" {
				continue
			}
			switch code {
			case "A":
				statusByPath[p] = "added"
			case "D":
				statusByPath[p] = "deleted"
			case "M", "T", "U":
				statusByPath[p] = "modified"
			default:
				// 其他状态（如类型变化等），按 modified 处理
				statusByPath[p] = "modified"
			}
		}
	}

	if err := scanner.Err(); err != nil {
		_ = cmd.Wait()
		return nil, fmt.Errorf("读取 git 输出失败: %w", err)
	}

	// flush last commit
	flushCommit(&all)

	errBytes, _ := ioReadAll(stderr)
	if err := cmd.Wait(); err != nil {
		msg := strings.TrimSpace(string(errBytes))
		if msg != "" {
			return nil, fmt.Errorf("git 执行失败: %s", msg)
		}
		return nil, fmt.Errorf("git 执行失败: %w", err)
	}

	return all, nil
}

func isNumStatLine(parts []string) bool {
	if len(parts) < 2 {
		return false
	}
	return isNumOrDash(parts[0]) && isNumOrDash(parts[1])
}

func isNumOrDash(s string) bool {
	s = strings.TrimSpace(s)
	if s == "-" {
		return true
	}
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func parseNumStatInt(s string) int {
	s = strings.TrimSpace(s)
	if s == "-" || s == "" {
		return 0
	}
	v, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return v
}

var renameBraceRe = regexp.MustCompile(`\{[^{}]*=>\s*([^{}]*)\}`)

// normalizeGitRenamePath 将 git numstat/name-status 中的“重命名路径表示法”归一化为新路径。
func normalizeGitRenamePath(p string) string {
	p = strings.TrimSpace(p)
	if p == "" {
		return ""
	}

	// 处理带花括号的 rename 语法：a/{old => new}/b -> a/new/b
	for renameBraceRe.MatchString(p) {
		p = renameBraceRe.ReplaceAllString(p, `$1`)
	}

	// 处理无花括号的 rename 语法：old => new
	if idx := strings.LastIndex(p, " => "); idx >= 0 {
		p = strings.TrimSpace(p[idx+4:])
	}

	return p
}

func ioReadAll(r interface{ Read([]byte) (int, error) }) ([]byte, error) {
	var b bytes.Buffer
	buf := make([]byte, 32*1024)
	for {
		n, err := r.Read(buf)
		if n > 0 {
			b.Write(buf[:n])
		}
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			return b.Bytes(), err
		}
	}
	return b.Bytes(), nil
}
