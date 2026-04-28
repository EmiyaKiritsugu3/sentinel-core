package graph

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

// TreeSitterScanner agora atua como um scanner resiliente (Pure Go)
// enquanto o ambiente CGO/Tree-sitter não é estabilizado.
type TreeSitterScanner struct {
	componentRegex *regexp.Regexp
	interfaceRegex *regexp.Regexp
}

func NewTreeSitterScanner() *TreeSitterScanner {
	return &TreeSitterScanner{
		// Identifica export function Component() ou const Component = () =>
		componentRegex: regexp.MustCompile(`(?:export\s+)?(?:function|const)\s+([A-Z][a-zA-Z0-9_]+)`),
		// Identifica interface IName
		interfaceRegex: regexp.MustCompile(`interface\s+([A-Za-z][a-zA-Z0-9_]+)`),
	}
}

func (s *TreeSitterScanner) SupportedExtensions() []string {
	return []string{".tsx", ".ts"}
}

func (s *TreeSitterScanner) Scan(path string) ScanResult {
	file, err := os.Open(path)
	if err != nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to open %s: %w", path, err)}
	}
	defer file.Close()

	res := ScanResult{}
	fileID := "file:" + path
	res.Nodes = append(res.Nodes, Node{
		ID:       fileID,
		Name:     filepath.Base(path),
		Type:     "file",
		FilePath: path,
	})

	scanner := bufio.NewScanner(file)
	lineNum := 0
	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Procura Componentes
		if matches := s.componentRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			symbolID := fmt.Sprintf("component:%s:%s", path, name)
			res.Nodes = append(res.Nodes, Node{
				ID:        symbolID,
				Name:      name,
				Type:      "component",
				FilePath:  path,
				StartLine: lineNum,
			})
			res.Edges = append(res.Edges, Edge{From: fileID, To: symbolID, Rel: "contains"})
		}

		// Procura Interfaces
		if matches := s.interfaceRegex.FindStringSubmatch(line); len(matches) > 1 {
			name := matches[1]
			symbolID := fmt.Sprintf("interface:%s:%s", path, name)
			res.Nodes = append(res.Nodes, Node{
				ID:        symbolID,
				Name:      name,
				Type:      "interface",
				FilePath:  path,
				StartLine: lineNum,
			})
			res.Edges = append(res.Edges, Edge{From: fileID, To: symbolID, Rel: "contains"})
		}
	}

	return res
}
