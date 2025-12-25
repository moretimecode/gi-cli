package utils

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"slices"
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

type matchFunc func(filename string) bool

func findFileInDir(dirPath string, matcher matchFunc) (string, error) {
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

func FindFilesByExtsInDir(ext *[]string, dirPath string) ([]string, error) {
	for i, e := range *ext {
		if !strings.HasPrefix(e, ".") {
			(*ext)[i] = "." + e
		}
	}
	return findFilesInDir(
		dirPath,
		func(filename string) bool {
			return slices.Contains(*ext, filepath.Ext(filename))
		},
	)
}

func findFilesInDir(dirPath string, matcher matchFunc) ([]string, error) {
	totalFiles := []string{}
	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && matcher(info.Name()) {
			absPath, err := filepath.Abs(path)
			if err == nil {
				totalFiles = append(totalFiles, absPath)
			}
		}
		return nil
	})
	return totalFiles, err
}

func CountLinesFast(path string) (int, error) {
	file, err := os.Open(path)
	if err != nil {
		return 0, err
	}
	defer file.Close()

	buf := make([]byte, 32*1024)
	lines := 0
	hasData := false
	lastByte := byte(0)

	for {
		n, err := file.Read(buf)
		if n > 0 {
			hasData = true
			lastByte = buf[n-1]
			for _, b := range buf[:n] {
				if b == '\n' {
					lines++
				}
			}
		}
		if err != nil {
			if err == io.EOF {
				break
			}
			return lines, err
		}
	}

	if hasData && lastByte != '\n' {
		lines++
	}

	return lines, nil
}
