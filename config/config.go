package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

const ConfigName = "toolsconfig.json"

type GoPlatform string

type DotNetPlatform string

const (
	GoLinux        GoPlatform     = "linux/amd64"
	GoWindows      GoPlatform     = "windows/amd64"
	GoDarwinAmd64  GoPlatform     = "darwin/amd64" // macOS Intel
	GoDarwinArm64  GoPlatform     = "darwin/arm64" // macOS Apple Silicon (M1/M2)
	DotnetLinux    DotNetPlatform = "linux-x64"
	DotnetWindows  DotNetPlatform = "win-x64"
	DotnetOsxAmd64 DotNetPlatform = "osx-x64"   // macOS Intel
	DotnetOsxArm64 DotNetPlatform = "osx-arm64" // macOS Apple Silicon (M1/M2)
)

type ConfigBase[T any] struct {
	ProjectName string `json:"ProjectName"`
	Config      T      `json:"Config"`
}

type DotnetProjectConfig struct {
	DotNetHosts     map[string]DotNetHost `json:"DotNetHosts"`
	TargetPlatforms []DotNetPlatform      `json:"TargetPlatforms"`
}

type GolangProjectConfig struct {
	PublishPath     string       `json:"PublishPath"`
	TargetPlatforms []GoPlatform `json:"TargetPlatforms"`
}

type DotnetConfig = ConfigBase[DotnetProjectConfig]
type GolangConfig = ConfigBase[GolangProjectConfig]

func (p GoPlatform) GOOS() string {
	parts := strings.Split(string(p), "/")
	return parts[0]
}

func (p GoPlatform) GOARCH() string {
	parts := strings.Split(string(p), "/")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

func (p DotNetPlatform) DotNetOS() string {
	parts := strings.Split(string(p), "-")
	return parts[0]
}

func (p DotNetPlatform) DotNetArch() string {
	parts := strings.Split(string(p), "-")
	if len(parts) > 1 {
		return parts[1]
	}
	return ""
}

type DotNetHost struct {
	CsProjectPath  string `json:"CsProjectPath"`
	DockerfilePath string `json:"DockerfilePath"`
	PublishPath    string `json:"PublishPath"`
}

var (
	golangConfigLoader = sync.OnceValues(loadConfig[GolangProjectConfig])
	dotnetConfigLoader = sync.OnceValues(loadConfig[DotnetProjectConfig])
)

func loadConfig[T any]() (*ConfigBase[T], error) {
	currentDir, err := os.Getwd()
	if err != nil {
		return nil, err
	}
	configPath := filepath.Join(currentDir, ConfigName)

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("配置文件不存在: %s", configPath)
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %v", err)
	}

	var config ConfigBase[T]
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %v", err)
	}

	return &config, nil
}

func LoadGolangConfig() (*GolangConfig, error) {
	return golangConfigLoader()
}

func LoadDotnetConfig() (*DotnetConfig, error) {
	return dotnetConfigLoader()
}
