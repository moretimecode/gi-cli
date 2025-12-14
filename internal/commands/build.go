package commands

import (
	"gocli/internal/dotnets"
	"gocli/internal/golangs"

	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "build",
	Short: "编译项目",
	Run: func(cmd *cobra.Command, args []string) {
		projectInfo := DetectProjectType()

		switch projectInfo.Type {
		case ProjectTypeGo:
			error := golangs.BuildGoProject(projectInfo.MainFilePaths[0])
			if error != nil {
				panic(error)
			}
		case ProjectTypeDotNet:
			error := dotnets.BuildDotNetProject()
			if error != nil {
				panic(error)
			}
		default:
			panic("未检测到 Go 或 .NET 项目")
		}
	},
}
