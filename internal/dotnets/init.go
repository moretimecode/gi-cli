package dotnets

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gocli/config"
	"gocli/internal/pathhelp"
)

func InitDotNetProject(mainFilePaths []string) error {
	if len(mainFilePaths) == 0 {
		return fmt.Errorf("未找到 .NET 项目文件")
	}

	currentDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("获取当前目录失败: %v", err)
	}

	projectMaps := make(map[string]config.DotNetHost)

	for _, path := range mainFilePaths {
		projectDir := filepath.Dir(path)
		if projectDir != "." {
			fmt.Printf("初始化 .NET 项目目录: %s\n", projectDir)
			csprojPath, err := pathhelp.FindFileByExtInDir(".csproj", projectDir)
			if err != nil {
				return fmt.Errorf("查找 .csproj 文件失败: %v", err)
			}
			if csprojPath != "" {
				fileName := filepath.Base(csprojPath)
				nameWithoutExt := strings.TrimSuffix(fileName, filepath.Ext(fileName))
				dockerfilePath, _ := pathhelp.FindFileByFullNameInDir("Dockerfile", projectDir)

				projectMaps[nameWithoutExt] = config.DotNetHost{
					CsProjectPath:  csprojPath,
					DockerfilePath: dockerfilePath,
				}
			}
			continue
		}
	}
	cfg := &config.Config{
		ProjectName: filepath.Base(currentDir),
		DotNetHosts: projectMaps,
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
