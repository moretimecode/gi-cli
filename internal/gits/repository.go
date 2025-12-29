package gits

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

// RepositoryEngine 定义仓储引擎类型
type RepositoryEngine string

const (
	// EngineAuto 自动选择：优先 native git，回退到 go-git
	EngineAuto RepositoryEngine = "auto"
	// EngineNative 强制使用原生 git 命令
	EngineNative RepositoryEngine = "native"
	// EngineGoGit 强制使用 go-git 库
	EngineGoGit RepositoryEngine = "go-git"
)

// GitRepository 定义“仓储层”接口：负责从 Git 仓库中加载基础数据结构。
//
// 约定：
// - 统计范围：全仓库（all refs 可达提交）
// - merge commit：不计入（no-merges）
//
// 备注：当前仅对 hotfile/stability/collab 相关链路使用。
// 其他命令仍可直接调用现有 LoadXxx 函数。
type GitRepository interface {
	LoadCommits(repoPath string) ([]GitCommitMetrics, error)
	LoadCommitStats(repoPath string, commits []GitCommitMetrics) ([]CommitStatsMetrics, error)
	LoadFileMetrics(repoPath string, commits []GitCommitMetrics) ([]FileMetrics, error)
}

// engineStrategy 定义引擎创建策略
type engineStrategy struct {
	condition func(engine RepositoryEngine, repoPath string) bool
	creator   func() GitRepository
	logMsg    string
}

// NewRepository 根据指定引擎创建仓储实例
func NewRepository(engine RepositoryEngine, repoPath string) GitRepository {
	// 定义引擎选择策略（按优先级排序）
	strategies := []engineStrategy{
		// 策略1: 明确指定使用 go-git
		{
			condition: func(e RepositoryEngine, _ string) bool {
				return e == EngineGoGit
			},
			creator: func() GitRepository {
				return &goGitRepository{}
			},
			logMsg: "🔧 使用 go-git 引擎",
		},
		// 策略2: 明确指定使用 native（且可用）
		{
			condition: func(e RepositoryEngine, path string) bool {
				return e == EngineNative && isNativeGitAvailable(path)
			},
			creator: func() GitRepository {
				return &nativeGitRepository{}
			},
			logMsg: "🔧 使用 Native Git 引擎",
		},
		// 策略3: 自动选择且 native 可用
		{
			condition: func(e RepositoryEngine, path string) bool {
				return e == EngineAuto && isNativeGitAvailable(path)
			},
			creator: func() GitRepository {
				return &nativeGitRepository{}
			},
			logMsg: "🔧 自动选择: Native Git 引擎",
		},
	}

	// 遍历策略，找到第一个匹配的
	for _, strategy := range strategies {
		if strategy.condition(engine, repoPath) {
			fmt.Println(strategy.logMsg)
			return strategy.creator()
		}
	}

	// 默认 fallback：使用 go-git
	if engine == EngineNative {
		fmt.Println("⚠️  Native Git 不可用，回退到 go-git")
	} else {
		fmt.Println("🔧 自动选择: go-git 引擎")
	}
	return &goGitRepository{}
}

func isNativeGitAvailable(repoPath string) bool {
	if _, err := exec.LookPath("git"); err != nil {
		return false
	}

	cmd := exec.Command("git", "-C", repoPath, "rev-parse", "--is-inside-work-tree")
	cmd.Env = os.Environ()
	out, err := cmd.Output()
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(out)) == "true"
}
