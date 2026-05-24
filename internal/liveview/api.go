// Package liveview provides a WebSocket-based live graph viewer.
package liveview

import (
	"database/sql"
	"encoding/json"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// GraphSnapshot is a JSON-serializable graph state.
type GraphSnapshot struct {
	Nodes []graph.Node `json:"nodes"`
	Edges []graph.Edge `json:"edges"`
}

// TaskStatus is the JSON response for GET /api/status.
// It reflects the most recent task from the SQLite tasks table.
type TaskStatus struct {
	ID           string  `json:"id"`
	Description  string  `json:"description"`
	Status       string  `json:"status"`
	Tier         *string `json:"tier,omitempty"`
	Verification *string `json:"verification,omitempty"`
	CreatedAt    *string `json:"created_at,omitempty"`
}

func handleGetGraph(db *sqlite.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers for local development
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Query Nodes
		nodesRows, err := db.Conn.Query("SELECT id, name, type, file_path, start_line, end_line, hash, last_indexed FROM nodes")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = nodesRows.Close() }()

		var nodes []graph.Node
		for nodesRows.Next() {
			var n graph.Node
			err := nodesRows.Scan(&n.ID, &n.Name, &n.Type, &n.FilePath, &n.StartLine, &n.EndLine, &n.Hash, &n.LastIndexed)
			if err == nil {
				nodes = append(nodes, n)
			}
		}

		// Query Edges
		edgesRows, err := db.Conn.Query("SELECT from_node_id, to_node_id, relation_type FROM edges")
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		defer func() { _ = edgesRows.Close() }()

		var edges []graph.Edge
		for edgesRows.Next() {
			var e graph.Edge
			err := edgesRows.Scan(&e.From, &e.To, &e.Rel)
			if err == nil {
				edges = append(edges, e)
			}
		}

		snapshot := GraphSnapshot{
			Nodes: nodes,
			Edges: edges,
		}

		// Standard #07 - Memory Integrity: Stream directly to ResponseWriter
		encoder := json.NewEncoder(w)
		if err := encoder.Encode(snapshot); err != nil {
			log.Printf("liveview: failed to encode graph snapshot: %v", err)
		}
	}
}

func handleGetStatus(db *sqlite.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		row := db.Conn.QueryRow(
			"SELECT id, description, status, tier, verification_command, created_at FROM tasks ORDER BY created_at DESC, rowid DESC LIMIT 1",
		)

		var status TaskStatus
		var tier, verification, createdAt sql.NullString
		err := row.Scan(&status.ID, &status.Description, &status.Status, &tier, &verification, &createdAt)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				// No tasks yet — return empty status.
				encoder := json.NewEncoder(w)
				if err := encoder.Encode(TaskStatus{}); err != nil {
					log.Printf("liveview: failed to encode empty status: %v", err)
				}
				return
			}
			log.Printf("liveview: failed to query task status: %v", err)
			w.WriteHeader(http.StatusInternalServerError)
			if err := json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"}); err != nil {
				log.Printf("liveview: failed to encode error response: %v", err)
			}
			return
		}

		if tier.Valid {
			status.Tier = &tier.String
		}
		if verification.Valid {
			v := verification.String
			status.Verification = &v
		}
		if createdAt.Valid {
			c := createdAt.String
			status.CreatedAt = &c
		}

		encoder := json.NewEncoder(w)
		if err := encoder.Encode(status); err != nil {
			log.Printf("liveview: failed to encode task status: %v", err)
		}
	}
}

func handleGetCode(db *sqlite.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		filePath := r.URL.Query().Get("path")
		if filePath == "" {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "missing required query param: path"})
			return
		}

			cleanPath := filepath.Clean(filePath)
		if filepath.IsAbs(cleanPath) || strings.HasPrefix(cleanPath, "..") {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
			return
		}

		baseDir, _ := os.Getwd()
		absPath, err := filepath.Abs(cleanPath)
		if err != nil || !strings.HasPrefix(absPath, baseDir+string(filepath.Separator)) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path outside root"})
			return
		}

		startStr := r.URL.Query().Get("start")
		endStr := r.URL.Query().Get("end")

		start := 1
		if startStr != "" {
			v, err := strconv.Atoi(startStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid start param"})
				return
			}
			if v > 0 {
				start = v
			}
		}

		content, err := os.ReadFile(absPath) //nolint:gosec // path is strictly validated against baseDir
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "file not found: " + filePath})
				return
			}
			slog.Error("liveview: failed to read file", "path", filePath, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		allLines := strings.Split(string(content), "\n")
		if len(allLines) > 0 && allLines[len(allLines)-1] == "" {
			allLines = allLines[:len(allLines)-1]
		}

		if len(allLines) == 0 {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"file": filePath, "lines": []string{}, "startLine": 0, "endLine": 0,
			})
			return
		}

		end := len(allLines)
		if endStr != "" {
			v, err := strconv.Atoi(endStr)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid end param"})
				return
			}
			if v > end {
				v = end
			}
			if v < 1 {
				v = 1
			}
			end = v
		}

		if start < 1 {
			start = 1
		}
		if start > len(allLines) {
			start = len(allLines)
		}
		if end < start {
			end = start
		}
		if end > len(allLines) {
			end = len(allLines)
		}

		lines := allLines[start-1 : end]

		resp := map[string]any{
			"file":      filePath,
			"lines":     lines,
			"startLine": start,
			"endLine":   end,
		}

		_ = json.NewEncoder(w).Encode(resp)
	}
}

func handleListADR(db *sqlite.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		entries, err := os.ReadDir("docs/architecture/adr")
		if err != nil {
			if os.IsNotExist(err) {
				_ = json.NewEncoder(w).Encode(map[string]any{"adrs": []any{}})
				return
			}
			slog.Error("liveview: failed to read ADR directory", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		adrs := make([]map[string]string, 0)
		for _, entry := range entries {
			if entry.IsDir() {
				continue
			}
			name := entry.Name()
			if !strings.HasPrefix(name, "ADR-") || !strings.HasSuffix(name, ".md") {
				continue
			}

			rest := strings.TrimPrefix(name, "ADR-")
			parts := strings.SplitN(rest, "-", 2)
			id := parts[0]

			title := ""
			if len(parts) > 1 {
				title = strings.TrimSuffix(parts[1], ".md")
				title = strings.ReplaceAll(title, "-", " ")
			}

			adrs = append(adrs, map[string]string{
				"id":       id,
				"title":    title,
				"filename": name,
			})
		}

		_ = json.NewEncoder(w).Encode(map[string]any{"adrs": adrs})
	}
}

func handleGetADR(db *sqlite.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		filename := strings.TrimPrefix(r.URL.Path, "/api/adr/")
		if filename == "" || strings.Contains(filename, "..") || strings.HasPrefix(filename, "/") {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
			return
		}

		fullPath := filepath.Join("docs/architecture/adr", filename)
		absPath, err := filepath.Abs(fullPath)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
			return
		}

		adrDir, err := filepath.Abs("docs/architecture/adr")
		if err != nil || !strings.HasPrefix(absPath, adrDir+string(filepath.Separator)) {
			w.WriteHeader(http.StatusBadRequest)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "invalid path"})
			return
		}

		content, err := os.ReadFile(absPath)
		if err != nil {
			if os.IsNotExist(err) {
				w.WriteHeader(http.StatusNotFound)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": "file not found: " + filename})
				return
			}
			slog.Error("liveview: failed to read ADR file", "filename", filename, "error", err)
			w.WriteHeader(http.StatusInternalServerError)
			_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		rest := strings.TrimPrefix(filename, "ADR-")
		parts := strings.SplitN(rest, "-", 2)
		id := parts[0]
		title := ""
		if len(parts) > 1 {
			title = strings.TrimSuffix(parts[1], ".md")
			title = strings.ReplaceAll(title, "-", " ")
		}

		resp := map[string]any{
			"id":       id,
			"title":    title,
			"content":  string(content),
			"filename": filename,
		}

		_ = json.NewEncoder(w).Encode(resp)
	}
}
