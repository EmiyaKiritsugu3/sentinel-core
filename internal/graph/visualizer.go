package graph

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
)

type Visualizer struct {
	db *sqlite.DB
}

func NewVisualizer(db *sqlite.DB) *Visualizer {
	return &Visualizer{db: db}
}

// GenerateMasterDiagram gera o C4 holístico do projeto
func (v *Visualizer) GenerateMasterDiagram() error {
	nodes, err := v.getNodes("")
	if err != nil {
		return fmt.Errorf("viz: failed to fetch master nodes: %w", err)
	}
	edges, err := v.getEdges()
	if err != nil {
		return fmt.Errorf("viz: failed to fetch master edges: %w", err)
	}

	content := "# Project Master Architecture [PID-SENTINEL]\n\n"
	content += "> [!IMPORTANT]\n> This is an auto-generated live map of the codebase.\n\n"
	content += "```mermaid\ngraph TD\n"
	content += v.formatMermaid(nodes, edges)
	content += "```\n"

	path := "docs/architecture/MASTER-GRAPH.md"
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("viz: failed to create architecture dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("viz: failed to write master graph: %w", err)
	}
	return nil
}

// GenerateTaskSnapshot gera um diagrama focado nos nós impactados por uma tarefa
func (v *Visualizer) GenerateTaskSnapshot(taskID, description string, impactFiles []string) error {
	var nodes []Node
	for _, file := range impactFiles {
		fileNodes, err := v.getNodes(file)
		if err != nil {
			return fmt.Errorf("viz: failed to fetch nodes for file %s: %w", file, err)
		}
		nodes = append(nodes, fileNodes...)
	}

	edges, err := v.getEdges()
	if err != nil {
		return fmt.Errorf("viz: failed to fetch snapshot edges: %w", err)
	}

	content := fmt.Sprintf("# Task Snapshot: %s [PID-SENTINEL]\n\n", taskID)
	content += fmt.Sprintf("## Goal: %s\n\n", description)
	content += "```mermaid\ngraph TD\n"
	content += v.formatMermaid(nodes, edges)
	content += "```\n"

	path := fmt.Sprintf("docs/architecture/tasks/%s-GRAPH.md", taskID)
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("viz: failed to create task dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("viz: failed to write task snapshot: %w", err)
	}
	return nil
}

func (v *Visualizer) getNodes(filterFile string) ([]Node, error) {
	query := "SELECT id, name, type FROM nodes"
	var args []interface{}
	if filterFile != "" {
		query += " WHERE file_path = ?"
		args = append(args, filterFile)
	}

	rows, err := v.db.Conn.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("viz: db query error (nodes): %w", err)
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Name, &n.Type); err != nil {
			return nil, fmt.Errorf("viz: row scan error (node): %w", err)
		}
		nodes = append(nodes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("viz: row iteration error (nodes): %w", err)
	}

	return nodes, nil
}

func (v *Visualizer) getEdges() ([]Edge, error) {
	rows, err := v.db.Conn.Query("SELECT from_node_id, to_node_id, relation_type FROM edges")
	if err != nil {
		return nil, fmt.Errorf("viz: db query error (edges): %w", err)
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.From, &e.To, &e.Rel); err != nil {
			return nil, fmt.Errorf("viz: row scan error (edge): %w", err)
		}
		edges = append(edges, e)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("viz: row iteration error (edges): %w", err)
	}

	return edges, nil
}

func (v *Visualizer) formatMermaid(nodes []Node, edges []Edge) string {
	var sb strings.Builder
	nodeMap := make(map[string]bool)

	for _, n := range nodes {
		nodeMap[n.ID] = true
		style := ""
		if n.Type == "struct" {
			style = ":::struct"
		} else if n.Type == "function" {
			style = ":::func"
		}

		safeID := utils.SanitizeID(n.ID)
		sb.WriteString(fmt.Sprintf("    %s[\"%s (%s)\"]%s\n", safeID, n.Name, n.Type, style))
	}

	for _, e := range edges {
		safeFrom := utils.SanitizeID(e.From)
		safeTo := utils.SanitizeID(e.To)

		if nodeMap[e.From] && nodeMap[e.To] {
			sb.WriteString(fmt.Sprintf("    %s -->|%s| %s\n", safeFrom, e.Rel, safeTo))
		}
	}

	sb.WriteString("\n    classDef struct fill:#f9f,stroke:#333,stroke-width:2px;\n")
	sb.WriteString("    classDef func fill:#bbf,stroke:#333,stroke-width:1px;\n")

	return sb.String()
}
