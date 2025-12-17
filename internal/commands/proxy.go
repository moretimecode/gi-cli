package commands

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

var ProxyCmd = &cobra.Command{
	Use:   "proxy",
	Short: "代理管理命令",
}

// proxy set - 显示设置代理的命令
var proxySetCmd = &cobra.Command{
	Use:   "set",
	Short: "显示设置代理的 PowerShell 命令",
	Run: func(cmd *cobra.Command, args []string) {
		socks5 := os.Getenv("socks5")

		if socks5 == "" {
			fmt.Println("错误: 未设置环境变量 socks5")
			fmt.Println("请先设置环境变量，例如: export socks5=\"127.0.0.1:1080\"")
			return
		}

		fmt.Println("复制并执行以下命令来设置代理:")
		fmt.Println()
		fmt.Printf("$env:http_proxy = \"%s\"\n", socks5)
		fmt.Printf("$env:https_proxy = \"%s\"\n", socks5)
		fmt.Println()
		fmt.Println("取消代理:")
		fmt.Println("$env:http_proxy = \"\"; $env:https_proxy = \"\"")
	},
}

// proxy run - 在代理环境下运行命令
var proxyRunCmd = &cobra.Command{
	Use:   "run [command]",
	Short: "在代理环境下运行命令",
	Args:  cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		socks5 := os.Getenv("socks5")

		if socks5 == "" {
			fmt.Println("错误: 未设置环境变量 socks5")
			return
		}

		// 创建命令
		command := exec.Command(args[0], args[1:]...)

		// 设置环境变量
		command.Env = os.Environ()
		command.Env = append(command.Env, fmt.Sprintf("http_proxy=%s", socks5))
		command.Env = append(command.Env, fmt.Sprintf("https_proxy=%s", socks5))

		// 连接输出
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin

		fmt.Printf("使用代理 %s 运行: %v\n", socks5, args)

		// 执行命令
		err := command.Run()
		if err != nil {
			fmt.Printf("命令执行失败: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	ProxyCmd.AddCommand(proxySetCmd)
	ProxyCmd.AddCommand(proxyRunCmd)
}
