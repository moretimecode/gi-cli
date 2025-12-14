package golangs

import (
	"fmt"
	"gocli/config"
	"os"
	"os/exec"
	"path/filepath"
)

const GoDefaultMainFile = "main.go"

func BuildGoProject(mainFilePath string) error {
	_, err := build(mainFilePath)
	return err
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
