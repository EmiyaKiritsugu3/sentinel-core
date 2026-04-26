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
		return err
	}
	edges, err := v.getEdges()
	if err != nil {
		return err
	}

	content := "# Project Master Architecture [PID-SENTINEL]\n\n"
	content += "> [!IMPORTANT]\n> This is an auto-generated live map of the codebase.\n\n"
	content += "```mermaid\ngraph TD\n"
	content += v.formatMermaid(nodes, edges)
	content += "```\n"

	path := "docs/architecture/MASTER-GRAPH.md"
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, []byte(content), 0644)
}

// GenerateTaskSnapshot gera um diagrama focado nos nós impactados por uma tarefa
func (v *Visualizer) GenerateTaskSnapshot(taskID, description string, impactFiles []string) error {
	var nodes []Node
	for _, file := range impactFiles {
		fileNodes, _ := v.getNodes(file)
		nodes = append(nodes, fileNodes...)
	}
	
	edges, _ := v.getEdges()

	content := fmt.Sprintf("# Task Snapshot: %s [PID-SENTINEL]\n\n", taskID)
	content += fmt.Sprintf("## Goal: %s\n\n", description)
	content += "```mermaid\ngraph TD\n"
	content += v.formatMermaid(nodes, edges)
	content += "```\n"

	path := fmt.Sprintf("docs/architecture/tasks/%s-GRAPH.md", taskID)
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, []byte(content), 0644)
}

type Node struct {
	ID   string
	Name string
	Type string
}

type Edge struct {
	From string
	To   string
	Rel  string
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
		return nil, err
	}
	defer rows.Close()

	var nodes []Node
	for rows.Next() {
		var n Node
		if err := rows.Scan(&n.ID, &n.Name, &n.Type); err != nil {
			return nil, err
		}
		nodes = append(nodes, n)
	}
	return nodes, nil
}

func (v *Visualizer) getEdges() ([]Edge, error) {
	rows, err := v.db.Conn.Query("SELECT from_node_id, to_node_id, relation_type FROM edges")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var edges []Edge
	for rows.Next() {
		var e Edge
		if err := rows.Scan(&e.From, &e.To, &e.Rel); err != nil {
			return nil, err
		}
		edges = append(edges, e)
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
