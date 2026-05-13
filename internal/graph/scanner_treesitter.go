package graph

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

// TreeSitterScanner agora utiliza o motor real Tree-sitter com suporte a concorrência segura
// e extração semântica via Queries.
type TreeSitterScanner struct {
	pool     *sync.Pool
	tsQuery  *sitter.Query
	tsxQuery *sitter.Query
}

const semanticQuery = `
(import_statement) @import
(interface_declaration) @interface
(class_declaration) @class
(function_declaration) @function
(variable_declarator) @variable
`

// NewTreeSitterScanner creates a new TreeSitterScanner with prepared queries and a parser pool.
func NewTreeSitterScanner() *TreeSitterScanner {
	tsQ, err := sitter.NewQuery([]byte(semanticQuery), typescript.GetLanguage())
	if err != nil {
		slog.Warn("failed to create TreeSitter query", "error", err)
	}
	tsxQ, err := sitter.NewQuery([]byte(semanticQuery), tsx.GetLanguage())
	if err != nil {
		slog.Warn("failed to create TreeSitter TSX query", "error", err)
	}

	return &TreeSitterScanner{
		pool: &sync.Pool{
			New: func() interface{} {
				return sitter.NewParser()
			},
		},
		tsQuery:  tsQ,
		tsxQuery: tsxQ,
	}
}

// SupportedExtensions returns the file extensions supported by TreeSitterScanner.
func (s *TreeSitterScanner) SupportedExtensions() []string {
	return []string{".tsx", ".ts"}
}

// selectLanguage returns the appropriate Tree-sitter language and query for the
// given file extension.
func (s *TreeSitterScanner) selectLanguage(ext string) (*sitter.Language, *sitter.Query) {
	if ext == ".tsx" {
		return tsx.GetLanguage(), s.tsxQuery
	}
	return typescript.GetLanguage(), s.tsQuery
}

// executeQuery runs a Tree-sitter semantic query against the parsed tree and
// returns the resulting nodes and edges.
func (s *TreeSitterScanner) executeQuery(query *sitter.Query, tree *sitter.Tree, sourceCode []byte, path string) ScanResult {
	res := ScanResult{}
	fileID := "file:" + path
	res.Nodes = append(res.Nodes, Node{
		ID:       fileID,
		Name:     filepath.Base(path),
		Type:     "file",
		FilePath: path,
	})

	cursor := sitter.NewQueryCursor()
	defer cursor.Close()

	cursor.Exec(query, tree.RootNode())

	for {
		match, ok := cursor.NextMatch()
		if !ok {
			break
		}

		for _, capture := range match.Captures {
			captureName := query.CaptureNameForId(capture.Index)
			node := capture.Node

			switch captureName {
			case "import":
				for i := 0; i < int(node.NamedChildCount()); i++ {
					child := node.NamedChild(i)
					if child.Type() == "string" {
						importPath := strings.Trim(child.Content(sourceCode), "'\"")
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
				}
			case "interface", "class", "function", "variable":
				var name string
				for i := 0; i < int(node.NamedChildCount()); i++ {
					child := node.NamedChild(i)
					if child.Type() == "identifier" || child.Type() == "type_identifier" {
						name = child.Content(sourceCode)
						break
					}
				}

				if name != "" {
					s.processSymbol(node, captureName, name, path, fileID, &res)
				}
			}
		}
	}

	return res
}

// Scan parses a file using Tree-sitter and returns scanned nodes and edges.
func (s *TreeSitterScanner) Scan(path string) ScanResult {
	file, err := os.Open(path) //nolint:gosec // path from scanner input
	if err != nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to open %s: %w", path, err)}
	}
	defer func() { _ = file.Close() }()

	sourceCode, err := io.ReadAll(file)
	if err != nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to read %s: %w", path, err)}
	}

	lang, query := s.selectLanguage(filepath.Ext(path))

	parser := s.pool.Get().(*sitter.Parser)
	defer s.pool.Put(parser)

	parser.SetLanguage(lang)
	tree, err := parser.ParseCtx(context.Background(), nil, sourceCode)
	if err != nil || tree == nil || tree.RootNode() == nil {
		return ScanResult{Err: fmt.Errorf("scanner: failed to parse %s: %w", path, err)}
	}
	defer tree.Close()

	if query == nil {
		return ScanResult{Nodes: []Node{{ID: "file:" + path, Name: filepath.Base(path), Type: "file", FilePath: path}}}
	}

	return s.executeQuery(query, tree, sourceCode, path)
}

func (s *TreeSitterScanner) processSymbol(n *sitter.Node, captureName, name, path, fileID string, res *ScanResult) {
	// Determina o tipo real (Heurística de Componente React)
	symbolType := "symbol"
	switch captureName {
	case "interface":
		symbolType = "interface"
	case "class":
		symbolType = "class"
	case "function", "variable":
		if len(name) > 0 && name[0] >= 'A' && name[0] <= 'Z' {
			symbolType = "component"
		} else {
			symbolType = "function"
		}
	}

	// Encontra o nó pai para pegar o range real (ex: interface_declaration inteiro)
	parent := n.Parent()
	if parent == nil {
		parent = n
	}

	symbolID := fmt.Sprintf("%s:%s:%s", symbolType, path, name)
	res.Nodes = append(res.Nodes, Node{
		ID:        symbolID,
		Name:      name,
		Type:      symbolType,
		FilePath:  path,
		StartLine: int(parent.StartPoint().Row) + 1,
		EndLine:   int(parent.EndPoint().Row) + 1,
	})
	res.Edges = append(res.Edges, Edge{From: fileID, To: symbolID, Rel: "contains"})
}
