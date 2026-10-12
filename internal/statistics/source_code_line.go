package statistics

import (
	"fmt"
	"gocli/internal/utils"
	"os"
	"path/filepath"
	"runtime"
	"sync"
)

type FileSourceCodeLineStatistic struct {
	FilePath  string
	LineCount int
	Extension string
}

var statisticFileExtension = []string{".go", ".py", ".java", ".js", ".ts", ".cpp", ".c", ".rb", ".php", ".rs", ".sql", ".cs", ".html", ".css", ".swift", ".kt", ".m", ".sh", ".dart", ".scala", ".json", ".xml", ".yml", ".yaml", ".toml", ".ini", ".jsx", ".tsx", ".vue", ".fs", ".fsx", ".fsi", ".r", ".pl", ".lua", ".hs", ".erl", ".ex", ".exs", ".clj", ".cljs", ".groovy", ".bat", ".ps1", ".vb"}

var excludeDirs = []string{
	"node_modules/",
	"bin/",
	"vendor/",
	".git/",
	".svn/",
	".hg/",
	".idea/",
	".vscode/",
	"dist/",
	"build/",
	"out/",
	"publish/",
	".vs/",
	"_pycache_/",
}

// FileFilter is one stage in the pipeline. Include keeps matches; otherwise matches are removed.
type FileFilter struct {
	Include  bool
	Patterns []string
}

func Summary(path string, filters []FileFilter) error {
	if path == "" {
		var err error
		path, err = os.Getwd()
		if err != nil {
			return err
		}
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	filepaths, err := utils.FindFilesByExtsInDir(&statisticFileExtension, absPath)
	if err != nil {
		return err
	}
	filepaths = filterSourceFiles(absPath, filepaths, filters)

	resultCh := make(chan FileSourceCodeLineStatistic, len(filepaths))
	cpuCount := runtime.NumCPU()

	var wg sync.WaitGroup
	for cpu := range cpuCount {
		start := cpu * len(filepaths) / cpuCount
		end := (cpu + 1) * len(filepaths) / cpuCount
		if start >= end {
			continue
		}

		wg.Add(1)
		go func(paths []string) {
			defer wg.Done()
			statisticFilesCountLines(paths, resultCh)
		}(filepaths[start:end])
	}

	go func() {
		wg.Wait()
		close(resultCh)
	}()

	totalLines := 0
	statistics := make([]FileSourceCodeLineStatistic, 0, len(filepaths))
	for stat := range resultCh {
		totalLines += stat.LineCount
		statistics = append(statistics, stat)
	}

	printSourceCodeLineStatistics(totalLines, statistics)
	return nil
}

func filterSourceFiles(root string, paths []string, filters []FileFilter) []string {
	// Default exclusions are a separate first stage and cannot be undone by later stages.
	matchers := []utils.PathIgnoreMatcher{utils.NewPathIgnoreMatcher(root, excludeDirs)}
	for _, filter := range filters {
		matchers = append(matchers, utils.NewPathPatternMatcher(root, filter.Patterns))
	}
	result := make([]string, 0, len(paths))
	for _, path := range paths {
		if matchers[0].Match(path) {
			continue
		}
		keep := true
		for i, filter := range filters {
			if matchers[i+1].Match(path) != filter.Include {
				keep = false
				break
			}
		}
		if keep {
			result = append(result, path)
		}
	}
	return result
}

func printSourceCodeLineStatistics(totalLines int, statistics []FileSourceCodeLineStatistic) {
	if len(statistics) == 0 {
		fmt.Println("源代码行数统计结果:\n总行数: 0\n没有可统计的源代码文件")
		return
	}
	extensionStatisticsMap := make(map[string]int)
	for _, stat := range statistics {
		extensionStatisticsMap[stat.Extension] += stat.LineCount
	}
	byLineCount := func(stat FileSourceCodeLineStatistic) int { return stat.LineCount }
	query := utils.From(statistics)
	maxLineCountFile, _ := query.MaxBy(byLineCount)
	minLineCountFile, _ := query.MinBy(byLineCount)

	fmt.Printf("源代码行数统计结果:\n")
	fmt.Printf("总行数: %d\n", totalLines)
	fmt.Printf("各文件类型行数分布:\n")
	for ext, lineCount := range extensionStatisticsMap {
		fmt.Printf("  %s: %d 行\n", ext, lineCount)
	}
	fmt.Printf("行数最多的文件: %s (%d 行)\n", maxLineCountFile.FilePath, maxLineCountFile.LineCount)
	fmt.Printf("行数最少的文件: %s (%d 行)\n", minLineCountFile.FilePath, minLineCountFile.LineCount)
}

func statisticFilesCountLines(paths []string, fileSourceCodeLineStatistics chan FileSourceCodeLineStatistic) {
	for _, filePath := range paths {
		count, err := utils.CountLinesFast(filePath)
		if err != nil {
			continue
		}
		ext := filepath.Ext(filePath)
		fileSourceCodeLineStatistics <- FileSourceCodeLineStatistic{
			FilePath:  filePath,
			LineCount: count,
			Extension: ext,
		}
	}
}
