package commands

import (
	"fmt"
	"gocli/internal/dotnets"
	"gocli/internal/golangs"
	"gocli/internal/utils"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

var DeadCodeCmd = &cobra.Command{
	Use:   "deadcode [路径]",
	Short: "检测项目中的僵尸代码（未使用的函数、变量、类型等）",
	Long: `检测项目中未使用的代码，包括：
- 未使用的函数
- 未使用的变量和常量
- 未使用的类型定义

示例：
  tools deadcode           # 检测当前目录
  tools deadcode ./...     # 检测当前目录及所有子目录
  tools deadcode ./cmd     # 检测指定目录`,
	Run: runDeadCodeCheck,
}

func runDeadCodeCheck(cmd *cobra.Command, args []string) {
	// 1. 确定要分析的路径
	pattern := "./..." // 默认当前目录及子目录
	if len(args) > 0 {
		pattern = args[0]
	}

	fmt.Printf("🔍 开始检测僵尸代码: %s\n", pattern)

	var results []string

	// 2. 根据参数判断项目类型
	var projectType ProjectType
	var projectPath string

	// 如果参数是 .sln 或 .csproj 文件，直接判定为 .NET 项目
	if strings.HasSuffix(pattern, ".sln") || strings.HasSuffix(pattern, ".csproj") {
		projectType = ProjectTypeDotNet
		projectPath = pattern
	} else {
		// 处理 Go 风格的路径模式 (如 ./... 或 ./pkg/...)
		searchDir := pattern
		if strings.HasSuffix(pattern, "/...") || strings.HasSuffix(pattern, "\\...") {
			// 去掉 /... 后缀得到实际目录
			searchDir = strings.TrimSuffix(pattern, "/...")
			searchDir = strings.TrimSuffix(searchDir, "\\...")
			if searchDir == "." || searchDir == "" {
				searchDir = "."
			}
		}

		// 检查是否是目录
		fileInfo, err := os.Stat(searchDir)
		if err == nil && fileInfo.IsDir() {
			// 尝试在目录下查找 .sln 或 .csproj 文件
			slnFile, err := utils.FindFileByExtInDir(".sln", searchDir)
			if err == nil {
				projectType = ProjectTypeDotNet
				projectPath = slnFile
				fmt.Printf("📁 找到解决方案文件: %s\n", filepath.Base(slnFile))
			} else {
				// 没有 .sln，尝试查找 .csproj
				csprojFile, err := utils.FindFileByExtInDir(".csproj", searchDir)
				if err == nil {
					projectType = ProjectTypeDotNet
					projectPath = csprojFile
					fmt.Printf("📁 找到项目文件: %s\n", filepath.Base(csprojFile))
				} else {
					// 都没找到，检测是否是 Go 项目
					projectInfo := DetectProjectType()
					projectType = projectInfo.Type
					projectPath = pattern
				}
			}
		} else {
			// 不是目录，使用默认检测
			projectInfo := DetectProjectType()
			projectType = projectInfo.Type
			projectPath = pattern
		}
	}

	// 3. 执行相应的分析
	switch projectType {
	case ProjectTypeGo:
		fmt.Println("📦 检测到 Go 项目，正在加载...")

		detector := golangs.NewDeadCodeDetector()

		if err := detector.LoadProject(pattern); err != nil {
			fmt.Printf("❌ 加载项目失败: %v\n", err)
			return
		}

		fmt.Println("🔬 正在分析代码...")
		if err := detector.Analyze(); err != nil {
			fmt.Printf("❌ 分析失败: %v\n", err)
			return
		}

		results = detector.Report()

	case ProjectTypeDotNet:
		fmt.Println("📦 检测到 .NET 项目，正在加载...")

		fmt.Printf("🔬 正在分析项目: %s\n", projectPath)

		detector := dotnets.NewDotNetDeadCodeDetector(projectPath)

		deadCode, err := detector.Analyze()
		if err != nil {
			fmt.Printf("❌ 分析失败: %v\n", err)
			return
		}

		results = deadCode

	default:
		fmt.Println("❌ 未识别的项目类型，仅支持 Go 和 .NET 项目")
		return
	}

	// 3. 报告结果
	fmt.Println("\n" + strings.Repeat("=", 60))
	if len(results) == 0 {
		fmt.Println("✅ 太棒了！没有发现僵尸代码！")
	} else {
		fmt.Printf("⚠️  发现 %d 个未使用的符号:\n\n", len(results))
		for _, line := range results {
			fmt.Println("  " + line)
		}
		fmt.Printf("\n总计: %d 个僵尸代码\n", len(results))
	}
	fmt.Println(strings.Repeat("=", 60))
}
