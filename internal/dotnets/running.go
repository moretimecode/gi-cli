package dotnets

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"gocli/config"
)

const DotNetDefaultProjectFile = "Program.cs"

func RunDotNetProject(hostType string) error {
	cfg, err := config.LoadDotnetConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}

	var projectPath string
	var found bool
	for key, value := range cfg.Config.DotNetHosts {
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

	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("运行失败: %v", err)
	}

	return nil
}

func PublishDotNetProject(hostType string) error {
	cfg, err := config.LoadDotnetConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}

	var projectPath string
	var publishPath string
	var found bool
	for key, value := range cfg.Config.DotNetHosts {
		if strings.EqualFold(key, hostType) {
			projectPath = value.CsProjectPath
			publishPath = value.PublishPath
			found = true
			break
		}
	}

	if !found {
		return fmt.Errorf("未找到指定的 .NET 启动项: %s", hostType)
	}

	fmt.Printf("项目路径: %s\n发布根目录: %s\n", projectPath, publishPath)

	for _, platform := range cfg.Config.TargetPlatforms {
		fmt.Printf("为平台 %s/%s 还原依赖...\n", platform.DotNetOS(), platform.DotNetArch())
		restoreCmd := exec.Command("dotnet", "restore", projectPath, "-r", string(platform))
		restoreCmd.Stdout = os.Stdout
		restoreCmd.Stderr = os.Stderr
		err = restoreCmd.Run()
		if err != nil {
			return fmt.Errorf("还原依赖失败 (%s): %v", platform, err)
		}

		fmt.Printf("为平台 %s/%s 发布中...\n", platform.DotNetOS(), platform.DotNetArch())
		targetPath := filepath.Join(publishPath, platform.DotNetOS())
		cmd := exec.Command("dotnet", "publish", projectPath,
			"-c", "Release",
			"-r", string(platform),
			"--self-contained", "false",
			"--no-restore",
			"-o", targetPath,
			"-p:PublishSingleFile=true",
			"-p:DebugType=None",
			"-p:EnableCompressionInSingleFile=true",
			"-p:PublishTrimmed=true")

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("发布失败: %v", err)
		}
	}

	if len(cfg.Config.TargetPlatforms) == 0 {
		cmd := exec.Command("dotnet", "publish", projectPath,
			"-c", "Release",
			"--no-restore",
			"-o", publishPath,
			"--self-contained", "false")

		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("发布失败: %v", err)
		}
	}

	return nil
}
func BuildDotNetDockerImage(hostType string) error {
	cfg, err := config.LoadDotnetConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %w", err)
	}
	var dockerfilePath string
	var found bool
	for key, value := range cfg.Config.DotNetHosts {
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
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("构建 Docker 镜像失败: %v", err)
	}
	return nil
}
