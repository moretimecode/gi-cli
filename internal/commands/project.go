package commands

import (
	"fmt"
	"gocli/internal/dotnets"
	"gocli/internal/golangs"

	"github.com/spf13/cobra"
)

type ProjectParams struct {
	Name string
}

var projectParams = &ProjectParams{}

var ProjectCmd = &cobra.Command{
	Use:   "project",
	Short: "项目相关命令",
}

var projectCreateCmd = &cobra.Command{
	Use:   "create <go|dotnet> <项目名>",
	Short: "创建新项目",
	Args:  cobra.ExactArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		projectTypeStr := args[0]
		projectName := args[1]

		// 解析项目类型
		projectType, ok := ParseProjectType(projectTypeStr)
		if !ok {
			fmt.Printf("错误：不支持的项目类型 '%s'，请使用 'go' 或 'dotnet'\n", projectTypeStr)
			return
		}

		fmt.Printf("正在创建 %s 项目: %s\n", projectTypeStr, projectName)

		switch projectType {
		case ProjectTypeGo:
			creatErr := golangs.CreateGoProject(projectName)
			if creatErr != nil {
				fmt.Printf("创建 Go 项目失败: %v\n", creatErr)
				return
			}
			initErr := golangs.InitGoProject(projectName)
			if initErr != nil {
				fmt.Printf("初始化 Go 项目失败: %v\n", initErr)
				return
			}
		case ProjectTypeDotNet:
			creatErr := dotnets.CreateDotNetProject(projectName)
			if creatErr != nil {
				fmt.Printf("创建 .NET 项目失败: %v\n", creatErr)
				return
			}
			projectInfo := DetectProjectType()
			initErr := dotnets.InitDotNetProject(projectInfo.MainFilePaths)
			if initErr != nil {
				fmt.Printf("初始化 .NET 项目失败: %v\n", initErr)
				return
			}
		default:
			fmt.Printf("错误：未实现的项目类型\n")
		}

		InstallCompletion(cmd.Root())
		fmt.Println()
		fmt.Printf("✓ 项目 %s 创建完成\n", projectName)

	},
}

var projectInitCmd = &cobra.Command{
	Use:   "init [项目名]",
	Short: "初始化项目",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		projectName := args[0]
		fmt.Printf("正在初始化项目: %s\n", projectName)
		projectInfo := DetectProjectType()
		switch projectInfo.Type {
		case ProjectTypeGo:
			err := golangs.InitGoProject(projectName)
			if err != nil {
				fmt.Printf("初始化 Go 项目失败: %v\n", err)
				return
			}
		case ProjectTypeDotNet:
			err := dotnets.InitDotNetProject(projectInfo.MainFilePaths)
			if err != nil {
				fmt.Printf("初始化 .NET 项目失败: %v\n", err)
				return
			}
		default:
			panic("未检测到 Go 或 .NET 项目")
		}

		// 自动安装补全（幂等操作）
		InstallCompletion(cmd.Root())

		fmt.Println()
		fmt.Printf("✓ 项目 %s 初始化完成\n", projectName)
	},
}

func init() {
	projectCreateCmd.Flags().StringVarP(&projectParams.Name, "name", "n", "", "项目名称")

	ProjectCmd.AddCommand(projectCreateCmd)
	ProjectCmd.AddCommand(projectInitCmd)
}
