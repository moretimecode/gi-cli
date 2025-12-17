package commands

import "github.com/spf13/cobra"

type httpParams struct {
	url    string `json:"url"`
	header string `json:"header"`
	method string `json:"method"`
	params string `json:"params"`
}

var httpCmd = &cobra.Command{
	Use:   "http",
	Short: "HTTP 代理管理命令",
}

func init() {

}
