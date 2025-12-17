package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

const ConfigName = "toolsconfig.json"

type Config struct {
	ProjectName string                `json:"ProjectName"`
	BuildPath   string                `json:"BuildPath"`
	DotNetHosts map[string]DotNetHost `json:"DotNetHosts"`
}

type DotNetHost struct {
	CsProjectPath  string `json:"CsProjectPath"`
	DockerfilePath string `json:"DockerfilePath"`
}

var (
	instance *Config
	once     sync.Once
)

func GetConfigInstance() *Config {
	once.Do(func() {
		instance = &Config{} // 先初始化

		currentDir, err := os.Getwd()
		if err != nil {
			panic(err)
		}
		configPath := filepath.Join(currentDir, ConfigName)

		if _, err := os.Stat(configPath); os.IsNotExist(err) {
			panic(fmt.Errorf("配置文件不存在: %s", configPath))
		}

		data, err := os.ReadFile(configPath)
		if err != nil {
			panic(fmt.Errorf("读取配置文件失败: %v", err))
		}

		if err := json.Unmarshal(data, instance); err != nil {
			panic(fmt.Errorf("解析配置文件失败: %v", err))
		}
	})
	return instance
}
