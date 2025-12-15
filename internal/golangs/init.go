package golangs

import (
	"encoding/json"
	"fmt"
	"gocli/config"
	"os"
	"path/filepath"
)

const GoBinDir = "bin"

func InitGoProject(projectName string) error {
	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %v", err)
	}
	var binFullDirPath = filepath.Join(currentDir, GoBinDir)
	err = os.MkdirAll(binFullDirPath, 0755)
	if err != nil {
		return err
	}
	cfg := &config.Config{
		ProjectName: projectName,
		BuildPath:   binFullDirPath,
		DotNetHosts: map[string]config.DotNetHost{},
		Proxy: config.Proxy{
			Socks5: "socks5://127.0.0.1:1080",
		},
	}
	configPath := filepath.Join(currentDir, config.ConfigName)
	data, err := json.MarshalIndent(cfg, "", "  ")
	if err != nil {
		return fmt.Errorf("序列化配置失败: %v", err)
	}
	err = os.WriteFile(configPath, data, 0644)
	if err != nil {
		return fmt.Errorf("写入配置文件失败: %v", err)
	}
	return nil
}
