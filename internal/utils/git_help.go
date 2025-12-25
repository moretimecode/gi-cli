package utils

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/go-git/go-git/v5/plumbing/format/gitignore"
)

type PathIgnoreMatcher interface {
	Match(absPath string) bool
}

type pathIgnoreMatcher struct {
	root    string
	matcher gitignore.Matcher
}

func NewPathIgnoreMatcher(root string, customIgnores []string) PathIgnoreMatcher {
	var patterns []gitignore.Pattern

	isExit, gitIgnorePaths := FindFileInCurrentAndSubDirs(".gitignore")
	if isExit {
		for _, gitignorePath := range gitIgnorePaths {
			if data, err := os.ReadFile(gitignorePath); err == nil {
				ps := parsePatternLines(strings.Split(string(data), "\n"))
				patterns = append(patterns, ps...)
			}
		}
	}

	for _, rule := range customIgnores {
		patterns = append(patterns, gitignore.ParsePattern(rule, nil))
	}

	matcher := gitignore.NewMatcher(patterns)

	return &pathIgnoreMatcher{
		root:    root,
		matcher: matcher,
	}
}

func (m *pathIgnoreMatcher) Match(absPath string) bool {
	rel, err := filepath.Rel(m.root, absPath)
	if err != nil {
		return false
	}

	rel = filepath.ToSlash(rel)

	segments := strings.Split(rel, "/")

	info, err := os.Stat(absPath)
	isDir := err == nil && info.IsDir()

	return m.matcher.Match(segments, isDir)
}

func parsePatternLines(lines []string) []gitignore.Pattern {
	var patterns []gitignore.Pattern

	for _, line := range lines {
		line = strings.TrimSpace(line)

		if line == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, gitignore.ParsePattern(line, nil))
	}

	return patterns
}
