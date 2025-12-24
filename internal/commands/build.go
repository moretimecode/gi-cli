package commands

import (
	"gocli/internal/dotnets"
	"gocli/internal/golangs"

	"github.com/spf13/cobra"
)

var BuildCmd = &cobra.Command{
	Use:   "publish",
	Short: "编译发布项目",
	Run: func(cmd *cobra.Command, args []string) {
		projectInfo := DetectProjectType()

		switch projectInfo.Type {
		case ProjectTypeGo:
			error := golangs.PublishGoProject(projectInfo.MainFilePaths[0])
			if error != nil {
				panic(error)
			}
		case ProjectTypeDotNet:
			error := dotnets.PublishDotNetProject(GetHostType())
			if error != nil {
				panic(error)
			}
		default:
			panic("未检测到 Go 或 .NET 项目")
		}
	},
}

var BuildImageCmd = &cobra.Command{
	Use:   "build-image",
	Short: "为项目构建 Docker 镜像",
	Run: func(cmd *cobra.Command, args []string) {
		projectInfo := DetectProjectType()
		switch projectInfo.Type {
		case ProjectTypeGo:
			error := golangs.BuildGoDockerImage(projectInfo.MainFilePaths[0])
			if error != nil {
				panic(error)
			}
		case ProjectTypeDotNet:
			error := dotnets.BuildDotNetDockerImage(GetHostType())
			if error != nil {
				panic(error)
			}
		}
	},
}

func init() {
	RegisterHostTypeFlag(BuildCmd)
	RegisterHostTypeFlag(BuildImageCmd)
}
