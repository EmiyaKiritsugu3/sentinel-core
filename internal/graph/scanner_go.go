package graph

import (
	"crypto/sha256"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"path/filepath"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type GoScanner struct {
	db *sqlite.DB
}

func NewGoScanner(db *sqlite.DB) *GoScanner {
	return &GoScanner{db: db}
}

// ScanProject varre o diretório em busca de arquivos Go e indexa seus símbolos
func (s *GoScanner) ScanProject(root string) error {
	fset := token.NewFileSet()

	return filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Ignora diretórios e arquivos que não são .go
		if info.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}

		// Ignora pastas de build/vendor
		if isIgnored(path) {
			return nil
		}

		return s.scanFile(fset, path)
	})
}

func (s *GoScanner) scanFile(fset *token.FileSet, path string) error {
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return fmt.Errorf("could not parse file %s: %w", path, err)
	}

	hash, _ := calculateHash(path)
	fileID := "file:" + path

	// 1. Indexa o Arquivo (Node)
	err = s.upsertNode(fileID, filepath.Base(path), "file", path, 0, 0, hash)
	if err != nil {
		return err
	}

	// 2. Indexa Símbolos Granulares (Functions, Structs)
	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcID := fmt.Sprintf("func:%s:%s", path, x.Name.Name)
			start := fset.Position(x.Pos()).Line
			end := fset.Position(x.End()).Line
			s.upsertNode(funcID, x.Name.Name, "function", path, start, end, "")
			s.createEdge(fileID, funcID, "contains")

		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.StructType); ok {
				structID := fmt.Sprintf("struct:%s:%s", path, x.Name.Name)
				start := fset.Position(x.Pos()).Line
				end := fset.Position(x.End()).Line
				s.upsertNode(structID, x.Name.Name, "struct", path, start, end, "")
				s.createEdge(fileID, structID, "contains")
			}
		}
		return true
	})

	return nil
}

func (s *GoScanner) upsertNode(id, name, nType, path string, start, end int, hash string) error {
	query := `
	INSERT INTO nodes (id, name, type, file_path, start_line, end_line, hash)
	VALUES (?, ?, ?, ?, ?, ?, ?)
	ON CONFLICT(id) DO UPDATE SET
		name=excluded.name,
		type=excluded.type,
		start_line=excluded.start_line,
		end_line=excluded.end_line,
		hash=excluded.hash,
		last_indexed=CURRENT_TIMESTAMP
	`
	_, err := s.db.Conn.Exec(query, id, name, nType, path, start, end, hash)
	return err
}

func (s *GoScanner) createEdge(from, to, rel string) error {
	query := `INSERT OR IGNORE INTO edges (from_node_id, to_node_id, relation_type) VALUES (?, ?, ?)`
	_, err := s.db.Conn.Exec(query, from, to, rel)
	return err
}

func isIgnored(path string) bool {
	ignored := []string{"vendor", "node_modules", ".git", "legacy"}
	for _, i := range ignored {
		if filepath.Base(path) == i || filepath.Base(filepath.Dir(path)) == i {
			return true
		}
	}
	return false
}

func calculateHash(path string) (string, error) {
	f, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer f.Close()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
