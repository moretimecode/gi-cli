package dotnets

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"gocli/config"
)

const DotNetDefaultProjectFile = "Program.cs"

func RunDotNetProject(hostType string) error {
	cfg := config.GetConfigInstance()

	var projectPath string
	var found bool
	for key, value := range cfg.DotNetHosts {
		if strings.EqualFold(key, hostType) {
			projectPath = value.CsProjectPath
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

func BuildDotNetDockerImage(hostType string) error {
	cfg := config.GetConfigInstance()
	var dockerfilePath string
	var found bool
	for key, value := range cfg.DotNetHosts {
		if strings.EqualFold(key, hostType) {
			dockerfilePath = value.DockerfilePath
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到指定的 .NET 启动项: %s", hostType)
	}

	if dockerfilePath == "" {
		return fmt.Errorf("未配置 Dockerfile 路径，请在配置文件中设置 DotNetHosts[%s].DockerfilePath", hostType)
	}

	if _, err := os.Stat(dockerfilePath); os.IsNotExist(err) {
		return fmt.Errorf("dockerfile 不存在: %s", dockerfilePath)
	}

	imageName := strings.ToLower(fmt.Sprintf("%s-%s-image:%d", cfg.ProjectName, hostType, time.Now().Unix()))
	cmd := exec.Command("docker", "build", "-f", dockerfilePath, "--no-cache", "--rm", "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("正在为 .NET 项目 (%s) 构建 Docker 镜像: %s\n", hostType, dockerfilePath)
	err := cmd.Run()
	if err != nil {
		return fmt.Errorf("构建 Docker 镜像失败: %v", err)
	}
	return nil
}
