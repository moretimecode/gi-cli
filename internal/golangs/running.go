package golangs

import (
	"fmt"
	"gocli/config"
	"gocli/internal/pathhelp"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const GoDefaultMainFile = "main.go"

func BuildGoProject(mainFilePath string) error {
	_, err := build(mainFilePath)
	return err
}

func BuildGoDockerImage(mainFilePath string) error {
	cfg := config.GetConfigInstance()
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %v", err)
	}

	dockerfilePath, err := pathhelp.FindFileByFullNameInDir(currentDir, "Dockerfile")
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
	exePath, err := build(mainFilePath)
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

func build(mainFilePath string) (string, error) {
	var config = config.GetConfigInstance()
	exeName := config.ProjectName + ".exe"
	outputPath := filepath.Join(config.BuildPath, exeName)
	err := os.MkdirAll(config.BuildPath, 0755)
	if err != nil {
		return "", fmt.Errorf("创建目录失败: %v", err)
	}

	cmd := exec.Command("go", "build", "-o", outputPath, mainFilePath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("编译失败: %v\n%s", err, string(output))
	}

	fmt.Printf("编译成功: %s\n", outputPath)
	return outputPath, nil
}
