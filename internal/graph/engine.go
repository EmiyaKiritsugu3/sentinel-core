package graph

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
)

type Engine struct {
	db       *sqlite.DB
	scanners map[string]FileScanner
	filter   *utils.IgnoreFilter
}

func NewEngine(db *sqlite.DB) *Engine {
	return &Engine{
		db:       db,
		scanners: make(map[string]FileScanner),
	}
}

func (e *Engine) RegisterScanner(s FileScanner) {
	for _, ext := range s.SupportedExtensions() {
		e.scanners[ext] = s
	}
}

// ScanProject varre o diretório e coordena os scanners registrados
func (e *Engine) ScanProject(root string) error {
	// Inicializa o filtro soberano baseado no .gitignore
	e.filter = utils.NewIgnoreFilter(root)

	filesChan := make(chan string)
	resultsChan := make(chan ScanResult)
	var wg sync.WaitGroup

	// 1. Inicia o Worker Pool
	numWorkers := 8
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for path := range filesChan {
				ext := filepath.Ext(path)
				scanner, ok := e.scanners[ext]
				if !ok {
					continue
				}

				// Verificação de Hash Incremental movida para o Engine para ser global
				res := e.scanFileWithIncrementalCheck(scanner, path)
				resultsChan <- res
			}
		}()
	}

	// 2. Coletor de Resultados (Escrita no DB serializada)
	var scanErr error
	done := make(chan bool)
	go func() {
		for res := range resultsChan {
			if res.Err != nil {
				scanErr = res.Err
				continue
			}
			if len(res.Nodes) == 0 {
				continue
			}

			err := e.persistResult(res)
			if err != nil {
				scanErr = err
			}
		}
		done <- true
	}()

	// 3. File Walker utilizando o novo filtro dinâmico
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || e.filter.IsIgnored(path) {
			return nil
		}
		ext := filepath.Ext(path)
		if _, ok := e.scanners[ext]; ok {
			filesChan <- path
		}
		return nil
	})

	close(filesChan)
	wg.Wait()
	close(resultsChan)
	<-done

	if err != nil {
		return fmt.Errorf("engine: walk failed: %w", err)
	}

	if scanErr != nil {
		return scanErr
	}

	// 4. Linker Phase (S11: Dependency Linker)
	return e.LinkDependencies()
}

func (e *Engine) scanFileWithIncrementalCheck(scanner FileScanner, path string) ScanResult {
	hash, err := utils.CalculateHash(path)
	if err != nil {
		return ScanResult{Err: fmt.Errorf("engine: hash failed for %s: %w", path, err)}
	}

	var existingHash string
	fileID := "file:" + path
	err = e.db.Conn.QueryRow("SELECT hash FROM nodes WHERE id = ?", fileID).Scan(&existingHash)
	if err == nil && existingHash == hash {
		return ScanResult{}
	}

	res := scanner.Scan(path)
	if res.Err == nil && len(res.Nodes) > 0 {
		// Garante que o nó do arquivo tenha o hash atualizado
		for i := range res.Nodes {
			if res.Nodes[i].ID == fileID {
				res.Nodes[i].Hash = hash
				break
			}
		}
	}
	return res
}

func (e *Engine) persistResult(res ScanResult) error {
	tx, err := e.db.Conn.Begin()
	if err != nil {
		return fmt.Errorf("engine: could not start transaction: %w", err)
	}

	filePath := res.Nodes[0].FilePath
	_, err = tx.Exec("DELETE FROM nodes WHERE file_path = ? AND type != 'file'", filePath)
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("engine: failed to prune symbols in %s: %w", filePath, err)
	}

	for _, n := range res.Nodes {
		err = e.upsertNodeTx(tx, n)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	for _, ed := range res.Edges {
		err = e.createEdgeTx(tx, ed)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	return tx.Commit()
}

func (e *Engine) upsertNodeTx(tx *sql.Tx, n Node) error {
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
	_, err := tx.Exec(query, n.ID, n.Name, n.Type, n.FilePath, n.StartLine, n.EndLine, n.Hash)
	if err != nil {
		return fmt.Errorf("engine: upsert failed for %s: %w", n.ID, err)
	}
	return nil
}

func (e *Engine) createEdgeTx(tx *sql.Tx, ed Edge) error {
	query := `INSERT OR IGNORE INTO edges (from_node_id, to_node_id, relation_type) VALUES (?, ?, ?)`
	_, err := tx.Exec(query, ed.From, ed.To, ed.Rel)
	if err != nil {
		return fmt.Errorf("engine: edge creation failed: %w", err)
	}
	return nil
}
