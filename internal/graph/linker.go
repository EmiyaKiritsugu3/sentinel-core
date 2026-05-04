package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LinkDependencies resolve imports temporários para referências reais entre arquivos.
func (e *Engine) LinkDependencies() error {
	// 1. Busca todos os imports pendentes
	rows, err := e.db.Conn.Query("SELECT id, name, file_path FROM nodes WHERE type = 'unresolved_import'")
	if err != nil {
		return fmt.Errorf("linker: failed to query pending imports: %w", err)
	}
	defer rows.Close()

	type pendingLink struct {
		id         string
		importPath string
		sourceFile string
	}

	var pending []pendingLink
	for rows.Next() {
		var p pendingLink
		if err := rows.Scan(&p.id, &p.importPath, &p.sourceFile); err == nil {
			pending = append(pending, p)
		}
	}

	fmt.Printf("🔗 Sentinel: Linking %d dependencies...\n", len(pending))

	for _, p := range pending {
		targetFile, resolved := e.resolveImport(p.sourceFile, p.importPath)
		if resolved {
			err := e.createRealEdge(p.sourceFile, targetFile)
			if err != nil {
				fmt.Printf("⚠️  Linker: failed to link %s -> %s: %v\n", p.sourceFile, targetFile, err)
				continue
			}
			// Remove o nó temporário após resolução bem sucedida
			_, _ = e.db.Conn.Exec("DELETE FROM nodes WHERE id = ?", p.id)
		}
	}

	return nil
}

func (e *Engine) resolveImport(sourceFile, importPath string) (string, bool) {
	// 1. Resolve Go internal imports
	modulePrefix := "github.com/EmiyaKiritsugu3/sentinel-core/"
	if strings.HasPrefix(importPath, modulePrefix) {
		relativePath := strings.TrimPrefix(importPath, modulePrefix)
		// Go internal imports resolve to directories.
		// We'll link to the directory path for now (as file:path/to/dir)
		// Or better, check if there's at least one .go file in that dir.
		if _, err := os.Stat(relativePath); err == nil {
			return relativePath, true
		}
	}

	// 2. Resolve TypeScript/JS relative imports
	if !strings.HasPrefix(importPath, ".") {
		return "", false
	}

	baseDir := filepath.Dir(sourceFile)
	targetBase := filepath.Join(baseDir, importPath)

	// Possíveis extensões em ordem de prioridade
	exts := []string{".tsx", ".ts", ".js", ".jsx", "/index.tsx", "/index.ts"}

	for _, ext := range exts {
		fullPath := targetBase + ext
		if strings.HasSuffix(ext, "index.tsx") || strings.HasSuffix(ext, "index.ts") {
			fullPath = filepath.Join(targetBase, filepath.Base(ext))
		}

		if _, err := os.Stat(fullPath); err == nil {
			return fullPath, true
		}
	}

	return "", false
}

func (e *Engine) createRealEdge(sourceFile, targetFile string) error {
	tx, err := e.db.Conn.Begin()
	if err != nil {
		return err
	}

	fromID := "file:" + sourceFile
	toID := "file:" + targetFile

	query := `INSERT OR IGNORE INTO edges (from_node_id, to_node_id, relation_type) VALUES (?, ?, ?)`
	_, err = tx.Exec(query, fromID, toID, "imports")
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}
