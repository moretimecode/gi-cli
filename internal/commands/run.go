package commands

import (
	"gocli/internal/dotnets"
	"gocli/internal/golangs"

	"github.com/spf13/cobra"
)

var RunCmd = &cobra.Command{
	Use:   "run",
	Short: "编译并且运行项目",

	Run: func(cmd *cobra.Command, args []string) {
		projectInfo := DetectProjectType()

		switch projectInfo.Type {
		case ProjectTypeGo:
			if err := golangs.RunGoProject(projectInfo.MainFilePaths[0]); err != nil {
				cmd.Printf("运行 Go 项目失败: %v\n", err)
				return
			}
		case ProjectTypeDotNet:
			if err := dotnets.RunDotNetProject(GetHostType()); err != nil {
				cmd.Printf("运行 .NET 项目失败: %v\n", err)
				return
			}
		default:
			cmd.Println("错误: 未检测到 Go 或 .NET 项目")
		}
	},
}

func init() {
	RegisterHostTypeFlag(RunCmd)
}
