package commands

import (
	"fmt"
	"os"
	"os/exec"

	"gocli/config"

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
		cfg := config.GetConfigInstance()

		if cfg.Proxy.Socks5 == "" {
			fmt.Println("错误: toolsconfig.json 中未配置 Proxy.Socks5")
			fmt.Println("请先在配置文件中添加代理地址，例如: \"127.0.0.1:1080\"")
			return
		}

		fmt.Println("复制并执行以下命令来设置代理:")
		fmt.Println()
		fmt.Printf("$env:http_proxy = \"%s\"\n", cfg.Proxy.Socks5)
		fmt.Printf("$env:https_proxy = \"%s\"\n", cfg.Proxy.Socks5)
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
		cfg := config.GetConfigInstance()

		if cfg.Proxy.Socks5 == "" {
			fmt.Println("错误: toolsconfig.json 中未配置 Proxy.Socks5")
			return
		}

		// 创建命令
		command := exec.Command(args[0], args[1:]...)

		// 设置环境变量
		command.Env = os.Environ()
		command.Env = append(command.Env, fmt.Sprintf("http_proxy=%s", cfg.Proxy.Socks5))
		command.Env = append(command.Env, fmt.Sprintf("https_proxy=%s", cfg.Proxy.Socks5))

		// 连接输出
		command.Stdout = os.Stdout
		command.Stderr = os.Stderr
		command.Stdin = os.Stdin

		fmt.Printf("使用代理 %s 运行: %v\n", cfg.Proxy.Socks5, args)

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
