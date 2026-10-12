package utils

import (
	"io/fs"
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

	// Parent rules precede child rules; each .gitignore is scoped to its directory.
	_ = filepath.WalkDir(root, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return nil
		}
		if !entry.IsDir() {
			return nil
		}
		if entry.Name() == ".git" {
			return filepath.SkipDir
		}
		data, err := os.ReadFile(filepath.Join(path, ".gitignore"))
		if err != nil {
			return nil
		}
		rel, err := filepath.Rel(root, path)
		if err != nil {
			return nil
		}
		var domain []string
		if rel != "." {
			domain = strings.Split(filepath.ToSlash(rel), "/")
		}
		patterns = append(patterns, parsePatternLines(strings.Split(string(data), "\n"), domain)...)
		return nil
	})

	for _, rule := range customIgnores {
		patterns = append(patterns, gitignore.ParsePattern(rule, nil))
	}

	matcher := gitignore.NewMatcher(patterns)

	return &pathIgnoreMatcher{
		root:    root,
		matcher: matcher,
	}
}

// NewPathPatternMatcher matches only the supplied rules, without loading .gitignore.
func NewPathPatternMatcher(root string, rules []string) PathIgnoreMatcher {
	return &pathIgnoreMatcher{
		root:    root,
		matcher: gitignore.NewMatcher(parsePatternLines(rules, nil)),
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

	// Git does not allow a child exception to re-include an ignored parent directory.
	for i := 1; i < len(segments); i++ {
		if m.matcher.Match(segments[:i], true) {
			return true
		}
	}
	return m.matcher.Match(segments, isDir)
}

func parsePatternLines(lines []string, domain []string) []gitignore.Pattern {
	var patterns []gitignore.Pattern

	for _, line := range lines {
		line = strings.TrimSuffix(line, "\r")

		if strings.TrimSpace(line) == "" {
			continue
		}

		if strings.HasPrefix(line, "#") {
			continue
		}

		patterns = append(patterns, gitignore.ParsePattern(line, domain))
	}

	return patterns
}
