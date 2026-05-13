// Package utils provides shared utility functions for the project.
package utils

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"
)

// IgnoreFilter gerencia as regras de exclusão baseadas no .gitignore
type IgnoreFilter struct {
	patterns []string
}

// NewIgnoreFilter carrega os padrões de um diretório raiz
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

	// 2. Padrões do .gitignore
	for _, p := range f.patterns {
		pattern := strings.ToLower(strings.TrimPrefix(p, "./"))

		// Match exato de arquivo ou pasta
		if filepath.Base(cleanPath) == strings.Trim(pattern, "/") {
			return true
		}

		// Match de diretório (prefixo ou conteúdo)
		if strings.Contains(cleanPath, "/"+strings.Trim(pattern, "/")+"/") ||
			strings.HasPrefix(cleanPath, strings.TrimSuffix(pattern, "/")+"/") {
			return true
		}

		// Match de sufixo (extensões como *.log ou caminhos específicos)
		if strings.HasSuffix(cleanPath, pattern) && strings.Contains(pattern, ".") {
			return true
		}
	}

	// 3. Arquivos ocultos por padrão
	base := filepath.Base(path)
	if strings.HasPrefix(base, ".") && base != "." {
		return true
	}

	return false
}
