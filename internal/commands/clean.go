package commands

import (
	"os"
	"os/exec"
	"runtime"

	"github.com/spf13/cobra"
)

var CleanCmd = &cobra.Command{
	Use:   "cls",
	Short: "清理终端屏幕",
	Run: func(cmd *cobra.Command, args []string) {
		var command *exec.Cmd
		if runtime.GOOS == "windows" {
			command = exec.Command("cmd", "/c", "cls")
		} else {
			command = exec.Command("clear")
		}
		command.Stdout = os.Stdout
		command.Stdin = os.Stdin
		command.Stderr = os.Stderr
		_ = command.Run()
	},
}
