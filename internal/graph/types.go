package graph

import (
	"time"
)

// Node represents a symbol or file in the graph
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

// ScanResult contains data extracted from a single file
type ScanResult struct {
	Nodes []Node
	Edges []Edge
	Err   error
}

// FileScanner is the interface each language driver must implement
type FileScanner interface {
	Scan(path string) ScanResult
	SupportedExtensions() []string
}
