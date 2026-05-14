package graph

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
)

// Engine manages the scanning pipeline and coordinates scanners, observers, and persistence.
type Engine struct {
	db         *sqlite.DB
	scanners   map[string]FileScanner
	filter     *utils.IgnoreFilter
	observers  []Observer
	observeSem chan struct{}
	mu         sync.RWMutex
}

// NewEngine creates a new Engine with the given database connection.
func NewEngine(db *sqlite.DB) (*Engine, error) {
	if err := sqlite.ValidateDB(db, "graph-engine"); err != nil {
		return nil, err
	}
	return &Engine{
		db:         db,
		scanners:   make(map[string]FileScanner),
		observeSem: make(chan struct{}, 16),
	}, nil
}

// RegisterObserver registers an observer to receive graph lifecycle events.
func (e *Engine) RegisterObserver(o Observer) {
	e.mu.Lock()
	defer e.mu.Unlock()
	e.observers = append(e.observers, o)
}

func (e *Engine) notifyObservers(event GraphEvent) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	for _, o := range e.observers {
		// Notifies asynchronously with backpressure protection
		select {
		case e.observeSem <- struct{}{}:
			go func(observer Observer) {
				defer func() { <-e.observeSem }()
				observer.Notify(event)
			}(o)
		default:
			slog.Warn("observer backpressure: dropping event", "type", event.Type)
		}
	}
}

// RegisterScanner registers a file scanner for its supported file extensions.
func (e *Engine) RegisterScanner(s FileScanner) {
	for _, ext := range s.SupportedExtensions() {
		e.scanners[ext] = s
	}
}

// ScanProject scans the directory and coordinates registered scanners
func (e *Engine) ScanProject(ctx context.Context, root string) error {
	e.notifyObservers(GraphEvent{Type: EventScanStarted, Time: time.Now()})
	defer e.notifyObservers(GraphEvent{Type: EventScanCompleted, Time: time.Now()})

	// Initializes the sovereign filter based on .gitignore
	e.filter = utils.NewIgnoreFilter(root)

	filesChan := make(chan string)
	resultsChan := make(chan ScanResult)
	var wg sync.WaitGroup

	// 1. Start the Worker Pool
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

				// Incremental Hash Verification moved to Engine to be global
				res := e.scanFileWithIncrementalCheck(ctx, scanner, path)
				resultsChan <- res
			}
		}()
	}

	// 2. Result Collector (Serialized DB writes)
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

			err := e.persistResult(ctx, res)
			if err != nil {
				scanErr = err
			}
		}
		done <- true
	}()

	// 3. File Walker using the new dynamic filter
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
	return e.LinkDependencies(ctx)
}

func (e *Engine) scanFileWithIncrementalCheck(ctx context.Context, scanner FileScanner, path string) ScanResult {
	hash, err := utils.CalculateHash(path)
	if err != nil {
		return ScanResult{Err: fmt.Errorf("engine: hash failed for %s: %w", path, err)}
	}

	var existingHash string
	fileID := "file:" + path
	err = e.db.Conn.QueryRowContext(ctx, "SELECT hash FROM nodes WHERE id = ?", fileID).Scan(&existingHash)
	if err == nil && existingHash == hash {
		return ScanResult{}
	}

	res := scanner.Scan(path)
	if res.Err == nil && len(res.Nodes) > 0 {
		// Ensures the file node has the updated hash
		for i := range res.Nodes {
			if res.Nodes[i].ID == fileID {
				res.Nodes[i].Hash = hash
				break
			}
		}
	}
	return res
}

func (e *Engine) persistResult(ctx context.Context, res ScanResult) error {
	tx, err := e.db.Conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("engine: could not start transaction: %w", err)
	}

	filePath := res.Nodes[0].FilePath
	_, err = tx.ExecContext(ctx, "DELETE FROM nodes WHERE file_path = ? AND type != 'file'", filePath)
	if err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("engine: failed to prune symbols in %s: %w", filePath, err)
	}

	for _, n := range res.Nodes {
		err = e.upsertNodeTx(ctx, tx, n)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	for _, ed := range res.Edges {
		err = e.createEdgeTx(ctx, tx, ed)
		if err != nil {
			_ = tx.Rollback()
			return err
		}
	}

	err = tx.Commit()
	if err != nil {
		return fmt.Errorf("engine: commit failed: %w", err)
	}

	// Notifies observers after successful commit
	for _, n := range res.Nodes {
		e.notifyObservers(GraphEvent{Type: EventNodeUpserted, Payload: n, Time: time.Now()})
	}
	for _, ed := range res.Edges {
		e.notifyObservers(GraphEvent{Type: EventEdgeCreated, Payload: ed, Time: time.Now()})
	}

	return nil
}

func (e *Engine) upsertNodeTx(ctx context.Context, tx *sql.Tx, n Node) error {
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
	_, err := tx.ExecContext(ctx, query, n.ID, n.Name, n.Type, n.FilePath, n.StartLine, n.EndLine, n.Hash)
	if err != nil {
		return fmt.Errorf("engine: upsert failed for %s: %w", n.ID, err)
	}
	return nil
}

func (e *Engine) createEdgeTx(ctx context.Context, tx *sql.Tx, ed Edge) error {
	query := `INSERT OR IGNORE INTO edges (from_node_id, to_node_id, relation_type) VALUES (?, ?, ?)`
	_, err := tx.ExecContext(ctx, query, ed.From, ed.To, ed.Rel)
	if err != nil {
		return fmt.Errorf("engine: edge creation failed: %w", err)
	}
	return nil
}
