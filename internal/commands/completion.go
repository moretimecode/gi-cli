package commands

import (
	"os"

	"github.com/spf13/cobra"
)

var CompletionCmd = &cobra.Command{
	Use:   "completion [powershell|bash|zsh]",
	Short: "生成自动补全脚本",
	Long: `生成 shell 自动补全脚本

安装方法 (PowerShell):
  tools completion powershell > $HOME\tools-completion.ps1
  
  然后在 PowerShell 配置文件中添加:
  . $HOME\tools-completion.ps1
  
  重新加载: . $PROFILE`,
	Args: cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		switch args[0] {
		case "powershell":
			cmd.Root().GenPowerShellCompletion(os.Stdout)
		case "bash":
			cmd.Root().GenBashCompletion(os.Stdout)
		case "zsh":
			cmd.Root().GenZshCompletion(os.Stdout)
		}
	},
}
