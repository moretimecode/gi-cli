package commands

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
)

// InstallCompletion 安装自动补全（幂等操作，可以多次调用）
func InstallCompletion(rootCmd *cobra.Command) error {
	home := os.Getenv("USERPROFILE")
	if home == "" {
		return nil
	}

	scriptPath := filepath.Join(home, "tools-completion.ps1")

	f, err := os.Create(scriptPath)
	if err != nil {
		return nil
	}
	rootCmd.GenPowerShellCompletion(f)
	f.Close()

	profilePath := filepath.Join(home, "Documents", "WindowsPowerShell", "Microsoft.PowerShell_profile.ps1")

	if _, err := os.Stat(profilePath); os.IsNotExist(err) {
		os.MkdirAll(filepath.Dir(profilePath), 0755)
	}

	content, err := os.ReadFile(profilePath)
	if err != nil {
		content = []byte{}
	}

	loadCommand := fmt.Sprintf(". %s", scriptPath)

	if strings.Contains(string(content), "tools-completion.ps1") {
		return nil
	}

	newContent := string(content)
	if len(newContent) > 0 && !strings.HasSuffix(newContent, "\n") {
		newContent += "\n"
	}
	newContent += "\n# Tools CLI 自动补全\n"
	newContent += loadCommand + "\n"

	err = os.WriteFile(profilePath, []byte(newContent), 0644)
	if err != nil {
		return nil
	}

	return nil
}
