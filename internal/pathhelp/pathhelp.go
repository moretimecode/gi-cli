package pathhelp

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func FindFileInCurrentAndSubDirs(filename string) (bool, []string) {
	if _, err := os.Stat(filename); err == nil {
		return true, []string{filename}
	}

	var foundPaths []string

	err := filepath.Walk(".", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() && info.Name() == filename {
			foundPaths = append(foundPaths, path)
		}

		return nil
	})

	if len(foundPaths) > 0 {
		fmt.Printf("找到 %d 个文件:\n", len(foundPaths))
		for _, p := range foundPaths {
			fmt.Printf("  - %s\n", p)
		}
		return true, foundPaths
	}

	// 如果有其他错误
	if err != nil {
		fmt.Printf("遍历目录时出错: %v\n", err)
		return false, nil
	}
	return false, nil
}

// MatchFunc 定义文件名匹配函数类型
type MatchFunc func(filename string) bool

// findFileInDir 通用的目录内文件查找函数
func findFileInDir(dirPath string, matcher MatchFunc) (string, error) {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() && matcher(entry.Name()) {
			filePath := filepath.Join(dirPath, entry.Name())
			absPath, err := filepath.Abs(filePath)
			if err == nil {
				filePath = absPath
			}
			return filePath, nil
		}
	}

	return "", fmt.Errorf("未找到匹配的文件")
}

func FindFileByExtInDir(ext string, dirPath string) (string, error) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	return findFileInDir(
		dirPath,
		func(filename string) bool {
			return filepath.Ext(filename) == ext
		},
	)
}

func FindFileByFullNameInDir(filename string, dirPath string) (string, error) {
	return findFileInDir(
		dirPath,
		func(name string) bool {
			return name == filename
		},
	)
}
