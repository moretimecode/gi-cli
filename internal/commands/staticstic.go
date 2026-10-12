package commands

import (
	"encoding/csv"
	"fmt"
	"gocli/internal/statistics"
	"strings"

	"github.com/spf13/cobra"
)

var SummaryCmd = newSummaryCmd()

func newSummaryCmd() *cobra.Command {
	var projectPath string
	var filters []statistics.FileFilter
	cmd := &cobra.Command{
		Use:   "code-summary [目录路径]",
		Short: "统计源代码行数",
		Long:  "统计源代码行数，默认遵守 .gitignore。-f 排除匹配项，-s 只保留匹配项，可重复使用，按出现顺序依次过滤上一步的结果。",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(args) > 0 {
				if cmd.Flags().Changed("path") {
					return fmt.Errorf("目录路径不能同时通过位置参数和 --path 指定")
				}
				projectPath = args[0]
			}
			return statistics.Summary(projectPath, filters)
		},
	}
	cmd.Flags().StringVarP(&projectPath, "path", "p", "", "要统计的项目路径，默认为当前目录")
	cmd.Flags().VarP(&summaryFilterValue{filters: &filters}, "filters", "f", "排除匹配的文件或目录，支持 gitignore 规则及逗号分隔；按参数顺序执行")
	cmd.Flags().VarP(&summaryFilterValue{filters: &filters, include: true}, "select", "s", "只保留匹配的文件或目录，支持 gitignore 规则及逗号分隔；按参数顺序执行")
	return cmd
}

// Set is called by pflag in argument order, so both flags append to one pipeline.
type summaryFilterValue struct {
	filters *[]statistics.FileFilter
	include bool
}

func (v *summaryFilterValue) Set(value string) error {
	patterns, err := csv.NewReader(strings.NewReader(value)).Read()
	if err != nil {
		return fmt.Errorf("无效的过滤规则: %w", err)
	}
	*v.filters = append(*v.filters, statistics.FileFilter{Include: v.include, Patterns: patterns})
	return nil
}

func (v *summaryFilterValue) String() string { return "" }
func (v *summaryFilterValue) Type() string   { return "patterns" }
