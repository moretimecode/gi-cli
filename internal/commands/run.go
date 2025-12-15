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
			error := golangs.RunGoProject(projectInfo.MainFilePaths[0])
			if error != nil {
				panic(error)
			}
		case ProjectTypeDotNet:
			error := dotnets.RunDotNetProject(GetHostType())
			if error != nil {
				panic(error)
			}
		default:
			panic("未检测到 Go 或 .NET 项目")
		}
	},
}

func init() {
	RegisterHostTypeFlag(RunCmd)
}
