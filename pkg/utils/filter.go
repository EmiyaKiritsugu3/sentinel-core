// Package utils provides shared utility functions for the project.
package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreFilter manages exclusion rules based on .gitignore
type IgnoreFilter struct {
	patterns []string
}

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
	}
}

// IsIgnored verifica se um caminho deve ser ignorado
func (f *IgnoreFilter) IsIgnored(path string) bool {
	cleanPath := strings.ToLower(filepath.ToSlash(path))

	// 1. Hardcoded Sovereign Protections (Sempre ignorados)
	internalIgnores := []string{".git", ".sentinel", "vendor", "dist", "build"}
	for _, part := range internalIgnores {
		if strings.Contains(cleanPath, "/"+part+"/") ||
			strings.HasPrefix(cleanPath, part+"/") ||
			filepath.Base(cleanPath) == part {
			return true
		}
	}

	// 2. .gitignore patterns
	for _, p := range f.patterns {
		pattern := strings.ToLower(strings.TrimPrefix(p, "./"))

		// Exact file or folder match
		if filepath.Base(cleanPath) == strings.Trim(pattern, "/") {
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

	// 3. Hidden files by default
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}

	return false
}
