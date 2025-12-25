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
			if err := golangs.PublishGoProject(projectInfo.MainFilePaths[0]); err != nil {
				cmd.Printf("发布 Go 项目失败: %v\n", err)
				return
			}
		case ProjectTypeDotNet:
			if err := dotnets.PublishDotNetProject(GetHostType()); err != nil {
				cmd.Printf("发布 .NET 项目失败: %v\n", err)
				return
			}
		default:
			cmd.Println("错误: 未检测到 Go 或 .NET 项目")
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
			if err := golangs.BuildGoDockerImage(projectInfo.MainFilePaths[0]); err != nil {
				cmd.Printf("构建 Go Docker 镜像失败: %v\n", err)
				return
			}
		case ProjectTypeDotNet:
			if err := dotnets.BuildDotNetDockerImage(GetHostType()); err != nil {
				cmd.Printf("构建 .NET Docker 镜像失败: %v\n", err)
				return
			}
		default:
			cmd.Println("错误: 未检测到 Go 或 .NET 项目")
		}
	},
}

func init() {
	RegisterHostTypeFlag(BuildCmd)
	RegisterHostTypeFlag(BuildImageCmd)
}
