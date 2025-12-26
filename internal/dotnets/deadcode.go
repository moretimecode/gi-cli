package dotnets

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
)

// analyzerBinary 在平台特定的文件中定义：
// - deadcode_windows.go
// - deadcode_linux.go
// - deadcode_darwin.go

// DotNetDeadCodeDetector .NET 僵尸代码检测器
type DotNetDeadCodeDetector struct {
	projectPath string
}

// AnalysisResult C# 分析器的输出格式
type AnalysisResult struct {
	TotalDefined  int            `json:"totalDefined"`
	TotalUsed     int            `json:"totalUsed"`
	DeadCodeItems []DeadCodeItem `json:"deadCodeItems"`
}

// DeadCodeItem 僵尸代码项
type DeadCodeItem struct {
	Symbol string `json:"symbol"`
}

// NewDotNetDeadCodeDetector 创建 .NET 检测器
func NewDotNetDeadCodeDetector(projectPath string) *DotNetDeadCodeDetector {
	return &DotNetDeadCodeDetector{
		projectPath: projectPath,
	}
}

// Analyze 执行分析
func (d *DotNetDeadCodeDetector) Analyze() ([]string, error) {
	// 获取当前可执行文件所在目录
	exePath, err := os.Executable()
	if err != nil {
		return nil, fmt.Errorf("无法获取可执行文件路径: %w", err)
	}
	exeDir := filepath.Dir(exePath)

	// 分析器路径相对于可执行文件
	analyzerPath := filepath.Join(exeDir, "internal", "dotnets", "DeadCodeAnalyzer",
		"bin", "Release", "net10.0", "win-x64", "publish", "DeadCodeAnalyzer.exe")
	
	if _, err := os.Stat(analyzerPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("分析器不存在: %s\n请先运行 build_dotnet.bat 编译分析器", analyzerPath)
	}

	// 执行分析器
	cmd := exec.Command(analyzerPath, d.projectPath)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("执行分析器失败: %w\n输出: %s", err, output)
	}

	// 解析 JSON 结果
	var result AnalysisResult
	if err := json.Unmarshal(output, &result); err != nil {
		return nil, fmt.Errorf("解析结果失败: %w\n输出: %s", err, output)
	}

	// 5. 转换为字符串列表
	var results []string
	for _, item := range result.DeadCodeItems {
		results = append(results, item.Symbol)
	}

	return results, nil
}

// getAnalyzerBinary 获取当前平台的分析器二进制
func (d *DotNetDeadCodeDetector) getAnalyzerBinary() ([]byte, error) {
	// analyzerBinary 由平台特定的文件提供
	// 编译时会根据目标平台自动选择对应的文件
	return analyzerBinary, nil
}

// getAnalyzerFileName 获取分析器文件名
func (d *DotNetDeadCodeDetector) getAnalyzerFileName() string {
	if runtime.GOOS == "windows" {
		return "deadcode_analyzer.exe"
	}
	return "deadcode_analyzer"
}
