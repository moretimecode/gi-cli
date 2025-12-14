package golangs

import (
	"fmt"
	"os"
	"os/exec"
)

func CreateGoProject(projectName string) error {

	commands := [][]string{
		{"go", "mod", "init", projectName},
	}

	for _, cmdArgs := range commands {
		fmt.Printf("执行: %v\n", cmdArgs)

		cmd := exec.Command(cmdArgs[0], cmdArgs[1:]...)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("命令执行失败 %v: %v", cmdArgs, err)
		}
	}
	os.MkdirAll("cmd", 0755)
	os.MkdirAll("internal", 0755)
	os.MkdirAll("pkg", 0755)
	os.MkdirAll("configs", 0755)
	os.MkdirAll("api", 0755)
	err := createMainGoFile(projectName)
	if err != nil {
		return fmt.Errorf("创建 main.go 失败: %v", err)
	}
	return nil
}

func createMainGoFile(projectName string) error {
	mainGoContent := `package main
			import "fmt"

			func main() {
				fmt.Println("Hello from ` + projectName + `!")
			}
			`
	err := os.WriteFile("cmd/main.go", []byte(mainGoContent), 0644)
	if err != nil {
		return fmt.Errorf("创建 main.go 失败: %v", err)
	}
	fmt.Println("✓ 已创建 cmd/main.go 文件")
	return nil
}
