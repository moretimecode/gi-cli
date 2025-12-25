package main

import (
	"fmt"
	"os"

	"gocli/internal/commands"

	"github.com/spf13/cobra"
)

func main() {
	var rootCmd = &cobra.Command{
		Use:   "tools",
		Short: "这是 CLI 工具",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("欢迎使用 CLI 工具！使用 -h 查看可用命令。")
		},
	}
	rootCmd.AddCommand(commands.ProjectCreateCmd)
	rootCmd.AddCommand(commands.ProjectInitCmd)
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.BuildCmd)
	rootCmd.AddCommand(commands.BuildImageCmd)
	rootCmd.AddCommand(commands.ProxyCmd)
	rootCmd.AddCommand(commands.CompletionCmd)
	rootCmd.AddCommand(commands.HttpCmd)
	rootCmd.AddCommand(commands.GitCmd)
	rootCmd.AddCommand(commands.SummaryCmd)
	if err := rootCmd.Execute(); err != nil {
		fmt.Printf("命令执行失败: %v\n", err)
		os.Exit(1)
	}
}
