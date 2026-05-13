package graph

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"strings"
)

// GoScanner scans Go source files using the standard go/ast and go/parser packages.
type GoScanner struct {
	// Go Scanner no longer needs the DB directly
}

// NewGoScanner creates a new GoScanner.
func NewGoScanner() *GoScanner {
	return &GoScanner{}
}

// SupportedExtensions returns the file extensions supported by GoScanner.
func (s *GoScanner) SupportedExtensions() []string {
	return []string{".go"}
}

// Scan parses a Go file and returns scanned nodes and edges.
func (s *GoScanner) Scan(path string) ScanResult {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return ScanResult{Err: fmt.Errorf("scanner: go parse error in %s: %w", path, err)}
	}

	res := ScanResult{}
	fileID := "file:" + path
	res.Nodes = append(res.Nodes, Node{
		ID:       fileID,
		Name:     filepath.Base(path),
		Type:     "file",
		FilePath: path,
	})

	// Extracts imports from the Go file
	for _, imp := range f.Imports {
		importPath := strings.Trim(imp.Path.Value, "\"")
		importID := fmt.Sprintf("import:%s:%s", path, importPath)

		res.Nodes = append(res.Nodes, Node{
			ID:       importID,
			Name:     importPath,
			Type:     "unresolved_import",
			FilePath: path,
		})

		res.Edges = append(res.Edges, Edge{
			From: fileID,
			To:   importID,
			Rel:  "imports",
		})
	}

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			receiver := ""
			if x.Recv != nil && len(x.Recv.List) > 0 {
				if t, ok := x.Recv.List[0].Type.(*ast.StarExpr); ok {
					receiver = fmt.Sprintf("%s.", t.X)
				} else if t, ok := x.Recv.List[0].Type.(*ast.Ident); ok {
					receiver = fmt.Sprintf("%s.", t.Name)
				}
			}

			funcID := fmt.Sprintf("func:%s:%s%s", path, receiver, x.Name.Name)
			start := fset.Position(x.Pos()).Line
			end := fset.Position(x.End()).Line
			res.Nodes = append(res.Nodes, Node{
				ID:        funcID,
				Name:      x.Name.Name,
				Type:      "function",
				FilePath:  path,
				StartLine: start,
				EndLine:   end,
			})
			res.Edges = append(res.Edges, Edge{From: fileID, To: funcID, Rel: "contains"})

		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.StructType); ok {
				structID := fmt.Sprintf("struct:%s:%s", path, x.Name.Name)
				start := fset.Position(x.Pos()).Line
				end := fset.Position(x.End()).Line
				res.Nodes = append(res.Nodes, Node{
					ID:        structID,
					Name:      x.Name.Name,
					Type:      "struct",
					FilePath:  path,
					StartLine: start,
					EndLine:   end,
				})
				res.Edges = append(res.Edges, Edge{From: fileID, To: structID, Rel: "contains"})
			}
		}
		return true
	})

	return res
}
