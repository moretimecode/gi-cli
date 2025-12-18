package commands

import (
	"github.com/spf13/cobra"

	"gocli/internal/networks"
)

var (
	url         string
	httpHeaders []string
	httpMethod  string
	params      string
	outputFile  string
)

var HttpCmd = &cobra.Command{
	Use:   "http-req [url]",
	Short: "HTTP 代理管理命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		url = args[0]
		if err := networks.RunHTTPReqCommand(url, httpMethod, outputFile, httpHeaders, params); err != nil {
			cmd.PrintErrln("错误:", err)
		}
	},
}

func init() {
	HttpCmd.Flags().StringSliceVarP(&httpHeaders, "header", "H", []string{}, "添加请求头 (可多次使用)")
	HttpCmd.Flags().StringVarP(&httpMethod, "method", "X", "GET", "HTTP 请求方法")
	HttpCmd.Flags().StringVarP(&params, "params", "d", "", "请求参数")
	HttpCmd.Flags().StringVarP(&outputFile, "output", "o", "", "输出文件路径")
}
