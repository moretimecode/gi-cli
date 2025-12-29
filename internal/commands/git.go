package commands

import (
	"fmt"
	"gocli/internal/gits"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var (
	// gitEngine 存储用户选择的仓储引擎
	gitEngine string
)

var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git 工作区管理和分析工具",
}

var gitRefreshCmd = &cobra.Command{
	Use:   "refresh",
	Short: "刷新工作区：清理已删除文件并显示状态",
	Run: func(cmd *cobra.Command, args []string) {
		if err := gits.Refresh(); err != nil {
			cmd.PrintErrf("刷新失败: %v\n", err)
		}
	},
}

// ============ Git 分析命令 ============

// git rank - 贡献者排行榜
var gitRankCmd = &cobra.Command{
	Use:   "rank [path]",
	Short: "显示贡献者排行榜",
	Long:  "分析 Git 仓库，显示按提交数或代码量排名的贡献者排行榜",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		// 获取绝对路径
		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		rankType, _ := cmd.Flags().GetString("type")
		top, _ := cmd.Flags().GetInt("top")

		fmt.Printf("📊 正在分析仓库: %s\n\n", absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		// 1. 加载提交数据（全仓库、排除 merge）
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		// 2. 加载提交统计（用于作者的新增/删除行统计）
		commitStats, err := repo.LoadCommitStats(absPath, commits)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交统计失败: %v\n", err)
			return
		}

		// 3. 聚合作者数据
		authors := gits.LoadAuthors(commits, commitStats)

		// 4. 生成排行榜
		var ranks []gits.ContributorRank
		if rankType == "code" {
			ranks = gits.GetTopCodeWriters(authors, top)
			fmt.Printf("🏆 代码编写排行榜 (Top %d)\n", top)
			fmt.Println(strings.Repeat("=", 80))
			fmt.Printf("%s %s %s %s %s %s\n",
				padString("排名", 4),
				padString("作者", 20),
				padString("提交数", 10),
				padString("新增行", 10),
				padString("删除行", 10),
				padString("总改动", 10))
			fmt.Println(strings.Repeat("-", 80))
		} else {
			ranks = gits.GetTopContributors(authors, top)
			fmt.Printf("🏆 贡献者排行榜 (Top %d)\n", top)
			fmt.Println(strings.Repeat("=", 80))
			fmt.Printf("%s %s %s %s %s %s\n",
				padString("排名", 4),
				padString("作者", 20),
				padString("提交数", 10),
				padString("新增行", 10),
				padString("删除行", 10),
				padString("总改动", 10))
			fmt.Println(strings.Repeat("-", 80))
		}

		for i, rank := range ranks {
			fmt.Printf("%s %s %s %s %s %s\n",
				padString(fmt.Sprintf("%d", i+1), 4),
				padString(truncateString(rank.AuthorName, 20), 20),
				padString(fmt.Sprintf("%d", rank.TotalCommits), 10),
				padString(fmt.Sprintf("%d", rank.TotalAdded), 10),
				padString(fmt.Sprintf("%d", rank.TotalDeleted), 10),
				padString(fmt.Sprintf("%d", rank.TotalChanged), 10),
			)
		}
	},
}

// git night-owl - 夜猫子程序员分析
var gitNightOwlCmd = &cobra.Command{
	Use:   "night-owl [path]",
	Short: "找出夜猫子程序员",
	Long:  "分析哪些开发者喜欢在深夜 (22:00-06:00) 提交代码",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		threshold, _ := cmd.Flags().GetFloat64("threshold")

		fmt.Printf("🦉 正在分析夜间提交模式: %s\n\n", absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		nightOwls := gits.FindNightOwls(commits, threshold)

		if len(nightOwls) == 0 {
			fmt.Printf("😴 未发现夜猫子程序员（无深夜提交记录）\n")
			return
		}

		fmt.Printf("🦉 发现 %d 位夜猫子程序员 (有深夜 22:00-06:00 提交记录)\n", len(nightOwls))
		fmt.Println(strings.Repeat("=", 70))
		fmt.Printf("%s %s %s %s\n",
			padString("作者", 20),
			padString("夜间提交", 10),
			padString("白天提交", 10),
			padString("夜间占比", 10))
		fmt.Println(strings.Repeat("-", 70))

		for _, owl := range nightOwls {
			fmt.Printf("%s %s %s %s\n",
				padString(truncateString(owl.AuthorName, 20), 20),
				padString(fmt.Sprintf("%d", owl.NightCommits), 10),
				padString(fmt.Sprintf("%d", owl.DayCommits), 10),
				padString(fmt.Sprintf("%.1f%%", owl.NightRatio*100), 10),
			)
		}
	},
}

// git author - 查看某个人的提交模式
var gitAuthorCmd = &cobra.Command{
	Use:   "author [path] <author>",
	Short: "查看指定作者的提交模式",
	Long:  "分析指定开发者的提交时间分布（24小时、星期、月度）",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var repoPath, author string

		if len(args) == 1 {
			repoPath = "."
			author = args[0]
		} else {
			repoPath = args[0]
			author = args[1]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		fmt.Printf("👤 正在分析 [%s] 的提交模式: %s\n\n", author, absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		pattern := gits.GetAuthorCommitPattern(commits, author)

		if pattern.TotalCommits == 0 {
			fmt.Printf("❌ 未找到匹配的作者：%s\n", author)
			return
		}

		fmt.Printf("📊 提交统计：共 %d 次提交\n\n", pattern.TotalCommits)

		// 24小时分布
		fmt.Println("⏰ 24小时提交分布")
		fmt.Println(strings.Repeat("=", 60))

		maxHourly := 0
		for _, count := range pattern.HourlyCommits {
			if count > maxHourly {
				maxHourly = count
			}
		}

		for hour, count := range pattern.HourlyCommits {
			if count > 0 {
				bar := strings.Repeat("█", count*40/max(maxHourly, 1))
				fmt.Printf("%02d:00 │ %-40s %d (%.1f%%)\n",
					hour, bar, count,
					float64(count)/float64(pattern.TotalCommits)*100)
			}
		}

		// 星期分布
		fmt.Println("\n📅 星期提交分布")
		fmt.Println(strings.Repeat("=", 60))

		weekdays := []string{"周日", "周一", "周二", "周三", "周四", "周五", "周六"}
		maxWeekday := 0
		for _, count := range pattern.WeekdayCommits {
			if count > maxWeekday {
				maxWeekday = count
			}
		}

		for day, count := range pattern.WeekdayCommits {
			if count > 0 {
				bar := strings.Repeat("█", count*40/max(maxWeekday, 1))
				fmt.Printf("%s │ %-40s %d (%.1f%%)\n",
					weekdays[day], bar, count,
					float64(count)/float64(pattern.TotalCommits)*100)
			}
		}

		// 月度活跃度
		if len(pattern.MonthlyActivity) > 0 {
			fmt.Println("\n📈 月度提交趋势")
			fmt.Println(strings.Repeat("=", 60))
			fmt.Printf("%-10s %10s\n", "时间", "提交数")
			fmt.Println(strings.Repeat("-", 60))

			for _, period := range pattern.MonthlyActivity {
				fmt.Printf("%d-%02d     %10d\n", period.Year, period.Month, period.Commits)
			}
		}
	},
}

// git timeline - 时间线分析
var gitTimelineCmd = &cobra.Command{
	Use:   "timeline [path]",
	Short: "显示提交时间分布",
	Long:  "按小时或月份统计提交分布情况",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		viewType, _ := cmd.Flags().GetString("view")

		fmt.Printf("📅 正在分析提交时间线: %s\n\n", absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		if viewType == "hour" {
			// 按小时统计
			slots := gits.GetCommitsByHour(commits)
			fmt.Println("⏰ 24小时提交分布")
			fmt.Println(strings.Repeat("=", 60))

			maxCommits := 0
			for _, slot := range slots {
				if slot.Commits > maxCommits {
					maxCommits = slot.Commits
				}
			}

			for _, slot := range slots {
				bar := strings.Repeat("█", slot.Commits*40/max(maxCommits, 1))
				fmt.Printf("%02d:00 │ %-40s %d\n", slot.Hour, bar, slot.Commits)
			}
		} else {
			// 按月统计
			periods := gits.GetMonthlyActivity(commits)
			fmt.Println("📊 月度活跃度统计")
			fmt.Println(strings.Repeat("=", 70))
			fmt.Printf("%-10s %10s %10s\n", "时间", "提交数", "参与人数")
			fmt.Println(strings.Repeat("-", 70))

			for _, period := range periods {
				fmt.Printf("%d-%02d     %10d %10d\n",
					period.Year, period.Month,
					period.Commits, len(period.Authors))
			}
		}
	},
}

// git hotfile - 热点文件分析
var gitHotFileCmd = &cobra.Command{
	Use:   "hotfile [path]",
	Short: "显示最频繁修改的文件",
	Long:  "分析哪些文件被修改最频繁，可能是代码热点或不稳定的区域",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		top, _ := cmd.Flags().GetInt("top")

		fmt.Printf("🔥 正在分析热点文件: %s\n\n", absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		fileMetrics, err := repo.LoadFileMetrics(absPath, commits)
		if err != nil {
			cmd.PrintErrf("⚠️  警告: %v\n\n", err)
			// 继续执行，使用已加载的数据
		}

		if len(fileMetrics) == 0 {
			cmd.PrintErrf("❌ 没有可用的文件变更数据\n")
			return
		}

		hotFiles := gits.GetHotFiles(fileMetrics, absPath, top)

		fmt.Printf("🔥 Top %d 热点文件\n", top)
		fmt.Println(strings.Repeat("=", 70))
		fmt.Printf("%s %s %s\n",
			padString("排名", 4),
			padString("文件路径", 55),
			padString("修改次数", 10))
		fmt.Println(strings.Repeat("-", 70))

		for i, file := range hotFiles {
			fmt.Printf("%s %s %s\n",
				padString(fmt.Sprintf("%d", i+1), 4),
				padString(truncatePathSmart(file.FilePath, 55), 55),
				padString(fmt.Sprintf("%d", file.ChangeCount), 10),
			)
		}
	},
}

// git stability - 代码稳定性分析
var gitStabilityCmd = &cobra.Command{
	Use:   "stability [path]",
	Short: "分析代码稳定性",
	Long:  "找出长时间未修改的稳定代码区域",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		minDays, _ := cmd.Flags().GetInt("days")

		fmt.Printf("🛡️  正在分析代码稳定性: %s\n\n", absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		fileMetrics, err := repo.LoadFileMetrics(absPath, commits)
		if err != nil {
			cmd.PrintErrf("⚠️  警告: %v\n\n", err)
		}

		if len(fileMetrics) == 0 {
			cmd.PrintErrf("❌ 没有可用的文件变更数据\n")
			return
		}

		stabilities := gits.AnalyzeCodeStability(fileMetrics, commits, absPath)

		// 筛选稳定文件
		stableFiles := make([]gits.CodeStability, 0)
		for _, s := range stabilities {
			if s.DaysSinceLastChange >= minDays {
				stableFiles = append(stableFiles, s)
			}
		}

		fmt.Printf("🛡️  发现 %d 个稳定文件 (超过 %d 天未修改)\n", len(stableFiles), minDays)
		fmt.Println(strings.Repeat("=", 80))
		fmt.Printf("%s %s %s\n",
			padString("文件路径", 50),
			padString("距上次修改", 15),
			padString("平均修改频率", 15))
		fmt.Println(strings.Repeat("-", 80))

		for i, file := range stableFiles {
			if i >= 20 { // 只显示前20个
				break
			}
			fmt.Printf("%s %s %s\n",
				padString(truncatePathSmart(file.FilePath, 50), 50),
				padString(fmt.Sprintf("%d 天", file.DaysSinceLastChange), 15),
				padString(fmt.Sprintf("%.1f 天/次", file.ChangeFrequency), 15),
			)
		}
	},
}

// git collab - 协作分析
var gitCollabCmd = &cobra.Command{
	Use:   "collab [path]",
	Short: "分析开发者协作关系",
	Long:  "找出协作密切的开发者对（共同修改文件较多）",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		repoPath := "."
		if len(args) > 0 {
			repoPath = args[0]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		top, _ := cmd.Flags().GetInt("top")

		fmt.Printf("🤝 正在分析协作关系: %s\n\n", absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		fileMetrics, err := repo.LoadFileMetrics(absPath, commits)
		if err != nil {
			cmd.PrintErrf("⚠️  警告: %v\n\n", err)
		}

		fmt.Printf("📊 加载了 %d 个提交, %d 个文件变更记录\n\n", len(commits), len(fileMetrics))

		if len(fileMetrics) == 0 {
			cmd.PrintErrf("❌ 没有可用的文件变更数据\n")
			return
		}

		collabs := gits.FindCollaborations(fileMetrics, commits)

		if len(collabs) == 0 {
			fmt.Println("未发现协作关系")
			return
		}

		fmt.Printf("🤝 Top %d 协作关系\n", min(top, len(collabs)))
		fmt.Println(strings.Repeat("=", 70))
		fmt.Printf("%s %s %s\n",
			padString("开发者1", 25),
			padString("开发者2", 25),
			padString("共享文件数", 15))
		fmt.Println(strings.Repeat("-", 70))

		for i, collab := range collabs {
			if i >= top {
				break
			}
			fmt.Printf("%s %s %s\n",
				padString(truncateString(collab.Author1, 25), 25),
				padString(truncateString(collab.Author2, 25), 25),
				padString(fmt.Sprintf("%d", collab.SharedFiles), 15),
			)
		}
	},
}

var gitCollabDetailCmd = &cobra.Command{
	Use:   "collab-detail <author1> <author2> [author3...] [--path <path>]",
	Short: "查看多个作者的详细协作文件",
	Long:  "显示多个指定作者共同修改过的文件列表及修改次数",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		// 获取作者列表（所有参数都是作者）
		authors := args
		repoPath, _ := cmd.Flags().GetString("path")
		if repoPath == "" {
			repoPath = "."
		}

		if len(authors) < 2 {
			cmd.PrintErrf("❌ 至少需要指定 2 个作者\n")
			return
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		fmt.Printf("🤝 正在分析 %s 的协作关系: %s\n\n", strings.Join(authors, ", "), absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		fileMetrics, err := repo.LoadFileMetrics(absPath, commits)
		if err != nil {
			cmd.PrintErrf("⚠️  警告: %v\n\n", err)
		}

		if len(fileMetrics) == 0 {
			cmd.PrintErrf("❌ 没有可用的文件变更数据\n")
			return
		}

		details := gits.FindCollaborationDetail(fileMetrics, commits, authors...)

		if len(details) == 0 {
			fmt.Printf("❌ 未发现 %s 的协作文件\n", strings.Join(authors, ", "))
			fmt.Println("\n提示: 请确保作者名称完全匹配 Git 提交记录中的名称")
			return
		}

		fmt.Printf("📁 发现 %d 个协作文件\n", len(details))

		// 计算表格宽度
		filePathWidth := 50
		authorColWidth := 12
		totalWidth := filePathWidth + len(authors)*authorColWidth + 15

		fmt.Println(strings.Repeat("=", totalWidth))

		// 输出表头
		var header strings.Builder
		header.WriteString(padString("文件路径", filePathWidth))
		for _, author := range authors {
			header.WriteString(" " + padString(truncateString(author, authorColWidth-1), authorColWidth))
		}
		header.WriteString(" " + padString("总提交", 12))
		fmt.Println(header.String())
		fmt.Println(strings.Repeat("-", totalWidth))

		// 输出数据行
		for _, detail := range details {
			row := padString(truncatePathSmart(detail.FilePath, filePathWidth), filePathWidth)
			for _, author := range authors {
				count := detail.AuthorCommits[author]
				row += " " + padString(fmt.Sprintf("%d", count), authorColWidth)
			}
			row += " " + padString(fmt.Sprintf("%d", detail.TotalCommits), 12)
			fmt.Println(row)
		}

		// 输出统计信息
		fmt.Printf("\n📊 协作统计:\n")
		fmt.Printf("  • 共同修改文件数: %d\n", len(details))
		for _, author := range authors {
			total := 0
			for _, detail := range details {
				total += detail.AuthorCommits[author]
			}
			fmt.Printf("  • %s 在这些文件的提交数: %d\n", author, total)
		}
	},
}

var gitAuthorHotfileCmd = &cobra.Command{
	Use:   "author-hotfile [path] <author>",
	Short: "查看指定作者修改最频繁的文件",
	Long:  "分析指定开发者最常修改的文件列表",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		var repoPath, author string

		if len(args) == 1 {
			repoPath = "."
			author = args[0]
		} else {
			repoPath = args[0]
			author = args[1]
		}

		absPath, err := filepath.Abs(repoPath)
		if err != nil {
			cmd.PrintErrf("❌ 获取路径失败: %v\n", err)
			return
		}

		top, _ := cmd.Flags().GetInt("top")

		fmt.Printf("🔥 正在分析 [%s] 的热点文件: %s\n\n", author, absPath)

		repo := gits.NewRepository(gits.RepositoryEngine(gitEngine), absPath)
		commits, err := repo.LoadCommits(absPath)
		if err != nil {
			cmd.PrintErrf("❌ 加载提交失败: %v\n", err)
			return
		}

		fileMetrics, err := repo.LoadFileMetrics(absPath, commits)
		if err != nil {
			cmd.PrintErrf("⚠️  警告: %v\n\n", err)
		}

		if len(fileMetrics) == 0 {
			cmd.PrintErrf("❌ 没有可用的文件变更数据\n")
			return
		}

		hotFiles := gits.GetAuthorHotFiles(fileMetrics, commits, author, top)

		if len(hotFiles) == 0 {
			fmt.Printf("❌ 未找到匹配的作者：%s\n", author)
			return
		}

		fmt.Printf("🔥 [%s] 的 Top %d 热点文件\n", author, len(hotFiles))
		fmt.Println(strings.Repeat("=", 70))
		fmt.Printf("%s %s %s\n",
			padString("排名", 4),
			padString("文件路径", 55),
			padString("修改次数", 10))
		fmt.Println(strings.Repeat("-", 70))

		for i, file := range hotFiles {
			fmt.Printf("%s %s %s\n",
				padString(fmt.Sprintf("%d", i+1), 4),
				padString(truncatePathSmart(file.FilePath, 55), 55),
				padString(fmt.Sprintf("%d", file.ChangeCount), 10),
			)
		}

		// 统计信息
		totalChanges := 0
		for _, file := range hotFiles {
			totalChanges += file.ChangeCount
		}
		fmt.Printf("\n📊 统计:\n")
		fmt.Printf("  • 修改的文件数: %d\n", len(hotFiles))
		fmt.Printf("  • 总修改次数: %d\n", totalChanges)
	},
}

// 辅助函数
func truncateString(s string, maxLen int) string {
	width := displayWidth(s)
	if width <= maxLen {
		return s
	}

	// 截断到合适的长度
	result := ""
	currentWidth := 0
	for _, r := range s {
		charWidth := runeWidth(r)
		if currentWidth+charWidth+3 > maxLen {
			break
		}
		result += string(r)
		currentWidth += charWidth
	}
	return result + "..."
}

// truncatePathSmart 智能截断路径，优先保留文件名
func truncatePathSmart(path string, maxLen int) string {
	width := displayWidth(path)
	if width <= maxLen {
		return path
	}

	// 统一使用 / 作为分隔符（Git 内部使用 /）
	separator := "/"
	normalizedPath := strings.ReplaceAll(path, "\\", "/")

	// 分离路径和文件名
	parts := strings.Split(normalizedPath, separator)
	if len(parts) == 0 {
		return truncateString(path, maxLen)
	}

	fileName := parts[len(parts)-1]
	fileNameWidth := displayWidth(fileName)

	// 如果文件名本身就超长，直接截断
	if fileNameWidth+3 > maxLen {
		return truncateString(fileName, maxLen)
	}

	// 计算可用于路径的宽度
	availableWidth := maxLen - fileNameWidth - 4 // 留 4 个字符给 ".../"

	if availableWidth <= 0 {
		return ".../" + fileName
	}

	// 从前面开始保留尽可能多的路径
	prefix := ""
	currentWidth := 0
	for i := 0; i < len(parts)-1; i++ {
		partWidth := displayWidth(parts[i]) + 1 // +1 for "/"
		if currentWidth+partWidth > availableWidth {
			break
		}
		if prefix != "" {
			prefix += separator
		}
		prefix += parts[i]
		currentWidth += partWidth
	}

	if prefix == "" {
		return ".../" + fileName
	}
	return prefix + "/.../" + fileName
}

// displayWidth 计算字符串的显示宽度（中文=2，英文=1）
func displayWidth(s string) int {
	width := 0
	for _, r := range s {
		width += runeWidth(r)
	}
	return width
}

// runeWidth 计算单个字符的显示宽度
func runeWidth(r rune) int {
	if r < 128 {
		return 1 // ASCII 字符
	}
	// 中文、日文、韩文等宽字符
	if (r >= 0x4E00 && r <= 0x9FFF) || // CJK统一汉字
		(r >= 0x3400 && r <= 0x4DBF) || // CJK扩展A
		(r >= 0x20000 && r <= 0x2A6DF) || // CJK扩展B
		(r >= 0x2A700 && r <= 0x2B73F) || // CJK扩展C
		(r >= 0x2B740 && r <= 0x2B81F) || // CJK扩展D
		(r >= 0x2B820 && r <= 0x2CEAF) || // CJK扩展E
		(r >= 0xF900 && r <= 0xFAFF) || // CJK兼容汉字
		(r >= 0x2F800 && r <= 0x2FA1F) { // CJK兼容补充
		return 2
	}
	// 全角字符
	if r >= 0xFF00 && r <= 0xFFEF {
		return 2
	}
	return 1
}

// padString 补齐字符串到指定显示宽度
func padString(s string, width int) string {
	currentWidth := displayWidth(s)
	if currentWidth >= width {
		return s
	}
	return s + strings.Repeat(" ", width-currentWidth)
}

func init() {
	// 全局标志：所有 git 子命令都继承
	GitCmd.PersistentFlags().StringVarP(&gitEngine, "engine", "e", "auto",
		"指定 Git 仓储引擎 (auto|native|go-git)")

	// 原有命令
	GitCmd.AddCommand(gitRefreshCmd)

	// 分析命令
	GitCmd.AddCommand(gitRankCmd)
	GitCmd.AddCommand(gitNightOwlCmd)
	GitCmd.AddCommand(gitAuthorCmd)
	GitCmd.AddCommand(gitTimelineCmd)
	GitCmd.AddCommand(gitHotFileCmd)
	GitCmd.AddCommand(gitStabilityCmd)
	GitCmd.AddCommand(gitCollabCmd)
	GitCmd.AddCommand(gitCollabDetailCmd)
	GitCmd.AddCommand(gitAuthorHotfileCmd)

	// rank 命令参数
	gitRankCmd.Flags().StringP("type", "t", "commit", "排行榜类型：commit (提交数) 或 code (代码量)")
	gitRankCmd.Flags().IntP("top", "n", 10, "显示前 N 名")

	// night-owl 命令参数
	gitNightOwlCmd.Flags().Float64P("threshold", "t", 0.3, "夜间提交占比阈值 (0.0-1.0)")

	// timeline 命令参数
	gitTimelineCmd.Flags().StringP("view", "v", "hour", "视图类型：hour (小时) 或 month (月份)")

	// hotfile 命令参数
	gitHotFileCmd.Flags().IntP("top", "n", 20, "显示前 N 个文件")

	// stability 命令参数
	gitStabilityCmd.Flags().IntP("days", "d", 90, "最少未修改天数")

	// collab 命令参数
	gitCollabCmd.Flags().IntP("top", "n", 10, "显示前 N 对协作关系")

	// collab-detail 命令参数
	gitCollabDetailCmd.Flags().StringP("path", "p", ".", "仓库路径")

	// author-hotfile 命令参数
	gitAuthorHotfileCmd.Flags().IntP("top", "n", 20, "显示前 N 个文件")
}
