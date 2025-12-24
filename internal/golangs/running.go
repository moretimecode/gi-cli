package golangs

import (
	"fmt"
	"gocli/config"
	"gocli/internal/utils"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const GoDefaultMainFile = "main.go"

func PublishGoProject(mainFilePath string) error {
	_, err := publish(mainFilePath)
	return err
}

func BuildGoDockerImage(mainFilePath string) error {
	cfg, err := config.LoadGolangConfig()
	if err != nil {
		return fmt.Errorf("加载配置失败: %v", err)
	}
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %v", err)
	}

	dockerfilePath, err := utils.FindFileByFullNameInDir(currentDir, "Dockerfile")
	if err != nil {
		return fmt.Errorf("未找到 Dockerfile: %v", err)
	}

	imageName := strings.ToLower(fmt.Sprintf("%s-image:%d", cfg.ProjectName, time.Now().Unix()))
	fmt.Printf("正在构建 Docker 镜像: %s\n", imageName)

	cmd := exec.Command("docker", "build", "-f", dockerfilePath, "--no-cache", "--rm", "-t", imageName, ".")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Printf("正在为 GO 项目 (%s) 构建 Docker 镜像: %s\n", cfg.ProjectName, dockerfilePath)
	err = cmd.Run()
	if err != nil {
		return fmt.Errorf("构建 Docker 镜像失败: %v", err)
	}
	return nil
}

func RunGoProject(mainFilePath string) error {
	exePath, err := publish(mainFilePath)
	if err != nil {
		return err
	}

	fmt.Printf("正在运行: %s\n", exePath)

	cmd := exec.Command(exePath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	err = cmd.Run()
	return err
}

func publish(mainFilePath string) (string, error) {
	cfg, err := config.LoadGolangConfig()
	if err != nil {
		return "", fmt.Errorf("加载配置失败: %v", err)
	}

	currentOs := runtime.GOOS
	var resultPath string

	for _, platform := range cfg.Config.TargetPlatforms {
		fmt.Printf("为平台 %s/%s 编译中...\n", platform.GOOS(), platform.GOARCH())

		targetPath := cfg.Config.PublishPath + "\\" + platform.GOOS()
		executableExt := cfg.ProjectName
		if platform.GOOS() == "windows" {
			executableExt += ".exe"
		}

		err = os.MkdirAll(targetPath, 0755)
		if err != nil {
			return "", fmt.Errorf("创建目录失败: %v", err)
		}

		outputPath := filepath.Join(targetPath, executableExt)
		cmd := exec.Command("go", "build", "-o", outputPath, mainFilePath)
		cmd.Env = append(os.Environ(),
			"GOOS="+platform.GOOS(),
			"GOARCH="+platform.GOARCH(),
		)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return "", fmt.Errorf("编译失败: %v\n%s", err, string(output))
		}
		if currentOs == platform.GOOS() {
			resultPath = outputPath
		}
	}
	return resultPath, nil
}
