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

func FindFileByExtInDir(ext string, dirPath string) (string, error) {
	if !strings.HasPrefix(ext, ".") {
		ext = "." + ext
	}

	var csprojPath string
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return "", err
	}

	for _, entry := range entries {
		if !entry.IsDir() && filepath.Ext(entry.Name()) == ext {
			csprojPath = filepath.Join(dirPath, entry.Name())
			absPath, err := filepath.Abs(csprojPath)
			if err == nil {
				csprojPath = absPath
			}
			break
		}
	}

	if csprojPath == "" {
		return "", fmt.Errorf("目录 %s 中未找到 %s 文件", dirPath, ext)
	}

	return csprojPath, nil
}
