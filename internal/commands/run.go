package commands

import (
	"gocli/config"
	"gocli/internal/dotnets"
	"gocli/internal/golangs"

	"github.com/spf13/cobra"
)

var hostType string

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
			error := dotnets.RunDotNetProject(hostType)
			if error != nil {
				panic(error)
			}
		default:
			panic("未检测到 Go 或 .NET 项目")
		}
	},
}

func init() {
	RunCmd.Flags().StringVarP(&hostType, "hosttype", "t", "", "启动项名称")
	RunCmd.RegisterFlagCompletionFunc("hosttype", func(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
		cfg := config.GetConfigInstance()

		hosts := make([]string, 0, len(cfg.DotNetHosts))
		for name := range cfg.DotNetHosts {
			hosts = append(hosts, name)
		}

		return hosts, cobra.ShellCompDirectiveNoFileComp
	})
}
