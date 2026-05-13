// Package liveview provides a WebSocket-based live graph viewer.
package liveview

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/graph"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// GraphSnapshot is a JSON-serializable graph state.
type GraphSnapshot struct {
	Nodes []graph.Node `json:"nodes"`
	Edges []graph.Edge `json:"edges"`
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
