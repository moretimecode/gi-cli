package commands

import (
	"gocli/config"

	"github.com/spf13/cobra"
)

var (
	hostType string
)

func RegisterHostTypeFlag(cmd *cobra.Command) {
	cmd.Flags().StringVarP(&hostType, "hosttype", "t", "", "启动项名称")
	cmd.RegisterFlagCompletionFunc("hosttype", hostTypeCompletion)
}

func hostTypeCompletion(cmd *cobra.Command, args []string, toComplete string) ([]string, cobra.ShellCompDirective) {
	cfg, _ := config.LoadDotnetConfig()

	hosts := make([]string, 0, len(cfg.Config.DotNetHosts))
	for name := range cfg.Config.DotNetHosts {
		hosts = append(hosts, name)
	}

	return hosts, cobra.ShellCompDirectiveNoFileComp
}

func GetHostType() string {
	return hostType
}
