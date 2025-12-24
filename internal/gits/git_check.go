package gits

import (
	"fmt"
	"os/exec"
	"strings"
)

func Refresh() error {
	cmd := exec.Command("git", "ls-files", "--deleted")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("获取已删除文件列表失败: %v", err)
	}

	deletedFiles := strings.TrimSpace(string(output))

	if deletedFiles != "" {
		files := strings.Split(deletedFiles, "\n")
		fmt.Printf("正在清理 %d 个已删除的文件...\n", len(files))

		args := append([]string{"rm"}, files...)
		rmCmd := exec.Command("git", args...)
		if err := rmCmd.Run(); err != nil {
			return fmt.Errorf("清理已删除文件失败: %v", err)
		}
		fmt.Println("✓ 已删除文件清理完成")
	} else {
		fmt.Println("✓ 没有需要清理的已删除文件")
	}

	fmt.Println("\n工作区状态:")
	fmt.Println("─────────────────────────────────────")
	statusCmd := exec.Command("git", "st")
	statusCmd.Stdout = nil
	statusCmd.Stderr = nil
	output, err = statusCmd.Output()
	if err != nil {
		return fmt.Errorf("获取工作区状态失败: %v", err)
	}
	fmt.Print(string(output))

	return nil
}
