package graph

import (
	"time"
)

// Node representa um símbolo ou arquivo no grafo
type Node struct {
	ID          string
	Name        string
	Type        string // file, function, struct, interface, component, etc.
	FilePath    string
	StartLine   int
	EndLine     int
	Hash        string
	LastIndexed time.Time
}

// Edge represents a directed relationship between two nodes in the dependency graph.
type Edge struct {
	From string
	To   string
	Rel  string // contains, imports, calls, renders, etc.
}

// ScanResult contém os dados extraídos de um único arquivo
type ScanResult struct {
	Nodes []Node
	Edges []Edge
	Err   error
}

// FileScanner é a interface que cada driver de linguagem deve implementar
type FileScanner interface {
	Scan(path string) ScanResult
	SupportedExtensions() []string
}
