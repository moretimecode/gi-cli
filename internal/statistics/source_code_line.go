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

func Summary(path string, filterExtensions []string) error {
	if path == "" {
		path, _ = os.Getwd()
	}

	absPath, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	filepaths, err := utils.FindFilesByExtsInDir(&statisticFileExtension, absPath)
	if err != nil {
		return err
	}
	if len(filterExtensions) > 0 {
		excludeDirs = append(excludeDirs, filterExtensions...)
	}

	matcher := utils.NewPathIgnoreMatcher(absPath, excludeDirs)

	filteredFilepaths := make([]string, 0, len(filepaths))
	for _, fp := range filepaths {
		if matcher.Match(fp) {
			continue
		}
		filteredFilepaths = append(filteredFilepaths, fp)
	}

	filepaths = filteredFilepaths

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

func printSourceCodeLineStatistics(totalLines int, statistics []FileSourceCodeLineStatistic) {
	extensionStatisticsMap := make(map[string]int)
	for _, stat := range statistics {
		extensionStatisticsMap[stat.Extension] += stat.LineCount
	}
	maxLineCountFile := utils.MaxBy(statistics, func(a, b FileSourceCodeLineStatistic) bool {
		return a.LineCount < b.LineCount
	})

	minLineCountFile := utils.MinBy(statistics, func(a, b FileSourceCodeLineStatistic) bool {
		return a.LineCount < b.LineCount
	})

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
