package dotnets

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"gocli/config"
)

const DotNetDefaultProjectFile = "Program.cs"

func RunDotNetProject(hostType string) error {
	cfg := config.GetConfigInstance()

	var projectPath string
	var found bool
	for key, value := range cfg.DotNetHosts {
		if strings.EqualFold(key, hostType) {
			projectPath = value
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到指定的 .NET 启动项: %s", hostType)
	}

	cmd := exec.Command("dotnet", "run", "--project", projectPath)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	fmt.Printf("正在运行 .NET 项目 (%s): %s\n", hostType, projectPath)

	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("运行失败: %v", err)
	}

	return nil
}

func BuildDotNetProject() error {
	cmd := exec.Command("dotnet", "build")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("正在构建 .NET 项目 \n")
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("构建失败: %v", err)
	}
	return nil
}
