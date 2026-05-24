// Package utils provides shared utility functions for the project.
package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

type parsedPattern struct {
	baseMatch   string
	dirContains string
	dirPrefix   string
	suffix      string
	hasSuffix   bool
}

// IgnoreFilter manages exclusion rules based on .gitignore
type IgnoreFilter struct {
	patterns []string
	parsed   []parsedPattern
}

var internalIgnoresPaths = []string{"/.git/", "/.sentinel/", "/vendor/", "/dist/", "/build/"}
var internalIgnoresPrefixes = []string{".git/", ".sentinel/", "vendor/", "dist/", "build/"}
var internalIgnoresBases = []string{".git", ".sentinel", "vendor", "dist", "build"}

// NewIgnoreFilter loads patterns from a root directory
func NewIgnoreFilter(root string) *IgnoreFilter {
	f := &IgnoreFilter{}
	f.loadGitignore(root)
	return f
}

func (f *IgnoreFilter) loadGitignore(root string) {
	path := filepath.Join(root, ".gitignore")
	file, err := os.Open(path) //nolint:gosec // .gitignore path from trusted root
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		f.patterns = append(f.patterns, line)

		pattern := strings.ToLower(strings.TrimPrefix(line, "./"))
		parsed := parsedPattern{
			baseMatch:   strings.Trim(pattern, "/"),
			dirContains: "/" + strings.Trim(pattern, "/") + "/",
			dirPrefix:   strings.TrimSuffix(pattern, "/") + "/",
		}
		if strings.Contains(pattern, ".") {
			parsed.suffix = pattern
			parsed.hasSuffix = true
		}
		f.parsed = append(f.parsed, parsed)
	}
}

// IsIgnored verifica se um caminho deve ser ignorado
func (f *IgnoreFilter) IsIgnored(path string) bool {
	cleanPath := strings.ToLower(filepath.ToSlash(path))
	base := filepath.Base(cleanPath)

	// 1. Hardcoded Sovereign Protections (Sempre ignorados)
	for _, part := range internalIgnoresBases {
		if base == part {
			return true
		}
	}
	for _, part := range internalIgnoresPrefixes {
		if strings.HasPrefix(cleanPath, part) {
			return true
		}
	}
	for _, part := range internalIgnoresPaths {
		if strings.Contains(cleanPath, part) {
			return true
		}
	}

	// 2. .gitignore patterns
	for _, p := range f.parsed {
		// Exact file or folder match
		if base == p.baseMatch {
			return true
		}

		// Directory match (prefix or content)
		if strings.Contains(cleanPath, p.dirContains) || strings.HasPrefix(cleanPath, p.dirPrefix) {
			return true
		}

		// Suffix match (extensions like *.log or specific paths)
		if p.hasSuffix && strings.HasSuffix(cleanPath, p.suffix) {
			return true
		}
	}

	// Fallback to unparsed patterns if any (e.g. for manually constructed IgnoreFilter)
	if len(f.patterns) > 0 && len(f.parsed) == 0 {
		for _, p := range f.patterns {
			pattern := strings.ToLower(strings.TrimPrefix(p, "./"))
			// Exact file or folder match
			if base == strings.Trim(pattern, "/") {
				return true
			}
			// Directory match (prefix or content)
			if strings.Contains(cleanPath, "/"+strings.Trim(pattern, "/")+"/") ||
				strings.HasPrefix(cleanPath, strings.TrimSuffix(pattern, "/")+"/") {
				return true
			}
			// Suffix match (extensions like *.log or specific paths)
			if strings.HasSuffix(cleanPath, pattern) && strings.Contains(pattern, ".") {
				return true
			}
		}
	}

	// 3. Hidden files by default
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}

	return false
}
