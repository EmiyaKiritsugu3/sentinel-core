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
	"sync"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type GoScanner struct {
	db *sqlite.DB
}

func NewGoScanner(db *sqlite.DB) *GoScanner {
	return &GoScanner{db: db}
}

type scanResult struct {
	nodes []nodeData
	edges []edgeData
	err   error
}

type nodeData struct {
	id, name, nType, path, hash string
	start, end                  int
}

type edgeData struct {
	from, to, rel string
}

// ScanProject varre o diretório em paralelo e indexa seus símbolos
func (s *GoScanner) ScanProject(root string) error {
	fset := token.NewFileSet()
	filesChan := make(chan string)
	resultsChan := make(chan scanResult)
	var wg sync.WaitGroup

	// 1. Inicia o Worker Pool (Concorrência de CPU)
	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filesChan {
				res := s.scanFile(fset, path)
				resultsChan <- res
			}
		}()
	}

	// 2. Inicia o Coletor (Escrita no DB serializada para segurança)
	var scanErr error
	done := make(chan bool)
	go func() {
		for res := range resultsChan {
			if res.err != nil {
				scanErr = res.err
				continue
			}
			for _, n := range res.nodes {
				s.upsertNode(n.id, n.name, n.nType, n.path, n.start, n.end, n.hash)
			}
			for _, e := range res.edges {
				s.createEdge(e.from, e.to, e.rel)
			}
		}
		done <- true
	}()

	// 3. Varre os arquivos (File System Walk)
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || filepath.Ext(path) != ".go" || isIgnored(path) {
			return nil
		}
		filesChan <- path
		return nil
	})

	close(filesChan)
	wg.Wait()
	close(resultsChan)
	<-done

	if err != nil {
		return fmt.Errorf("walk failed: %w", err)
	}
	return scanErr
}

func (s *GoScanner) scanFile(fset *token.FileSet, path string) scanResult {
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return scanResult{err: fmt.Errorf("could not parse file %s: %w", path, err)}
	}

	hash, _ := calculateHash(path)
	fileID := "file:" + path

	res := scanResult{}
	res.nodes = append(res.nodes, nodeData{id: fileID, name: filepath.Base(path), nType: "file", path: path, hash: hash})

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			funcID := fmt.Sprintf("func:%s:%s", path, x.Name.Name)
			start := fset.Position(x.Pos()).Line
			end := fset.Position(x.End()).Line
			res.nodes = append(res.nodes, nodeData{id: funcID, name: x.Name.Name, nType: "function", path: path, start: start, end: end})
			res.edges = append(res.edges, edgeData{from: fileID, to: funcID, rel: "contains"})

		case *ast.TypeSpec:
			if _, ok := x.Type.(*ast.StructType); ok {
				structID := fmt.Sprintf("struct:%s:%s", path, x.Name.Name)
				start := fset.Position(x.Pos()).Line
				end := fset.Position(x.End()).Line
				res.nodes = append(res.nodes, nodeData{id: structID, name: x.Name.Name, nType: "struct", path: path, start: start, end: end})
				res.edges = append(res.edges, edgeData{from: fileID, to: structID, rel: "contains"})
			}
		}
		return true
	})

	return res
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
