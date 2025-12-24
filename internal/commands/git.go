package commands

import (
	"gocli/internal/gits"

	"github.com/spf13/cobra"
)

var GitCmd = &cobra.Command{
	Use:   "git",
	Short: "Git 工作区管理工具",
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

func init() {
	GitCmd.AddCommand(gitRefreshCmd)
}
