package main

import (
	"fmt"

	"gocli/internal/commands"

	"github.com/spf13/cobra"
)

func main() {
	// 创建根命令
	var rootCmd = &cobra.Command{
		Use:   "tools",
		Short: "这是我的 CLI 工具",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("Hello from Cobra!")
		},
	}
	rootCmd.AddCommand(commands.ProjectCmd)
	rootCmd.AddCommand(commands.RunCmd)
	rootCmd.AddCommand(commands.BuildCmd)
	rootCmd.AddCommand(commands.BuildImageCmd)
	rootCmd.AddCommand(commands.ProxyCmd)
	rootCmd.AddCommand(commands.CompletionCmd)
	rootCmd.Execute()
}
