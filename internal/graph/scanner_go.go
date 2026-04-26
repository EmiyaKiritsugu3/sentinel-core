package graph

import (
	"database/sql"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sync"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
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

	// 1. Inicia o Worker Pool
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

	// 2. Inicia o Coletor (Escrita no DB serializada e Atômica)
	var scanErr error
	done := make(chan bool)
	go func() {
		for res := range resultsChan {
			if res.err != nil {
				scanErr = res.err
				continue
			}
			if len(res.nodes) == 0 {
				continue
			}

			tx, err := s.db.Conn.Begin()
			if err != nil {
				scanErr = fmt.Errorf("scan: could not start transaction: %w", err)
				continue
			}

			filePath := res.nodes[0].path
			_, err = tx.Exec("DELETE FROM nodes WHERE file_path = ? AND type != 'file'", filePath)
			if err != nil {
				tx.Rollback()
				scanErr = fmt.Errorf("scan: failed to prune old symbols in %s: %w", filePath, err)
				continue
			}

			for _, n := range res.nodes {
				err = s.upsertNodeTx(tx, n.id, n.name, n.nType, n.path, n.start, n.end, n.hash)
				if err != nil {
					break
				}
			}
			if err == nil {
				for _, e := range res.edges {
					err = s.createEdgeTx(tx, e.from, e.to, e.rel)
					if err != nil {
						break
					}
				}
			}

			if err != nil {
				tx.Rollback()
				scanErr = fmt.Errorf("scan: transaction failed for %s: %w", filePath, err)
			} else {
				tx.Commit()
			}
		}
		done <- true
	}()

	// 3. Varre os arquivos
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
		return fmt.Errorf("scan: walk failed: %w", err)
	}
	return scanErr
}

func (s *GoScanner) scanFile(fset *token.FileSet, path string) scanResult {
	hash, err := utils.CalculateHash(path)
	if err != nil {
		return scanResult{err: fmt.Errorf("scan: hash failed for %s: %w", path, err)}
	}

	var existingHash string
	fileID := "file:" + path
	err = s.db.Conn.QueryRow("SELECT hash FROM nodes WHERE id = ?", fileID).Scan(&existingHash)
	if err == nil && existingHash == hash {
		return scanResult{} 
	}

	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		return scanResult{err: fmt.Errorf("scan: parse error in %s: %w", path, err)}
	}

	res := scanResult{}
	res.nodes = append(res.nodes, nodeData{id: fileID, name: filepath.Base(path), nType: "file", path: path, hash: hash})

	ast.Inspect(f, func(n ast.Node) bool {
		switch x := n.(type) {
		case *ast.FuncDecl:
			// Corrigindo Bug de Colisão de ID: Adicionando o Receiver (se existir)
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

func (s *GoScanner) upsertNodeTx(tx *sql.Tx, id, name, nType, path string, start, end int, hash string) error {
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
	_, err := tx.Exec(query, id, name, nType, path, start, end, hash)
	if err != nil {
		return fmt.Errorf("db: upsert failed for %s: %w", id, err)
	}
	return nil
}

func (s *GoScanner) createEdgeTx(tx *sql.Tx, from, to, rel string) error {
	query := `INSERT OR IGNORE INTO edges (from_node_id, to_node_id, relation_type) VALUES (?, ?, ?)`
	_, err := tx.Exec(query, from, to, rel)
	if err != nil {
		return fmt.Errorf("db: edge creation failed: %w", err)
	}
	return nil
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
