package commands

import (
	"gocli/internal/statistics"

	"github.com/spf13/cobra"
)

var (
	projectPath string
	filters     []string
)

var SummaryCmd = &cobra.Command{
	Use:   "code-summary",
	Short: "统计源代码行数",
	Run: func(cmd *cobra.Command, args []string) {
		if err := statistics.Summary(projectPath, filters); err != nil {
			cmd.Printf("统计源代码行数失败: %v\n", err)
			return
		}
	},
}

func init() {
	SummaryCmd.Flags().StringVarP(&projectPath, "path", "p", "", "要统计的项目路径，默认为当前目录")
	SummaryCmd.Flags().StringSliceVarP(&filters, "filters", "f", []string{}, "要排除的文件夹或文件类型过滤器,格式和gitignore一样")
}
