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
	query := "SELECT id, name, type, file_path FROM nodes"
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
		if err := rows.Scan(&n.ID, &n.Name, &n.Type, &n.FilePath); err != nil {
			return nil, fmt.Errorf("viz: row scan error (node): %w", err)
		}
		nodes = append(nodes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("viz: row iteration error (nodes): %w", err)
	}

	return nodes, nil
}

// GenerateC4ContainerDiagram gera um diagrama C4 de Nível 2 (Container)
func (v *Visualizer) GenerateC4ContainerDiagram() error {
	nodes, err := v.getNodes("")
	if err != nil {
		return fmt.Errorf("viz: failed to fetch nodes: %w", err)
	}
	edges, err := v.getEdges()
	if err != nil {
		return fmt.Errorf("viz: failed to fetch edges: %w", err)
	}

	content := "# System Container Architecture (C4 Level 2) [PID-SENTINEL]\n\n"
	content += "Este diagrama mostra os containers lógicos do Sentinel e como eles se comunicam.\n\n"
	content += "```mermaid\nC4Container\n"
	content += "    title Container diagram for Sentinel Core\n\n"
	content += v.formatC4Mermaid(nodes, edges)
	content += "```\n"

	path := "docs/architecture/C4-CONTAINER-GRAPH.md"
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return fmt.Errorf("viz: failed to create architecture dir: %w", err)
	}
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		return fmt.Errorf("viz: failed to write C4 graph: %w", err)
	}
	return nil
}

func (v *Visualizer) formatC4Mermaid(nodes []Node, edges []Edge) string {
	type container struct {
		id   string
		name string
		desc string
		isDb bool
	}

	containers := map[string]container{
		"CLI":      {id: "cli", name: "CLI Application", desc: "Interface Go/Cobra para desenvolvedores"},
		"Agents":   {id: "agents", name: "Agent Engine", desc: "Orquestração de loops cognitivos ReAct"},
		"Graph":    {id: "graph", name: "Graph Engine", desc: "Análise AST e extração semântica"},
		"Audit":    {id: "audit", name: "Compliance Guard", desc: "Validação de padrões e Hard Gates"},
		"State":    {id: "state", name: "State Manager", desc: "Gerenciamento de tarefas e histórico"},
		"Frontend": {id: "frontend", name: "Legacy Frontend", desc: "Componentes legados em TypeScript"},
		"Database": {id: "db", name: "SQLite Graph", desc: "Persistência de nós, arestas e tarefas", isDb: true},
	}

	var sb strings.Builder
	for _, c := range containers {
		if c.isDb {
			sb.WriteString(fmt.Sprintf("    ContainerDb(%s, \"%s\", \"SQLite\", \"%s\")\n", c.id, c.name, c.desc))
		} else {
			sb.WriteString(fmt.Sprintf("    Container(%s, \"%s\", \"Go\", \"%s\")\n", c.id, c.name, c.desc))
		}
	}
	sb.WriteString("\n")

	// Mapeia node ID para container ID
	nodeToContainer := make(map[string]string)

	// Helper para extrair CID do caminho
	getContainerID := func(path string) string {
		if strings.Contains(path, "cmd/sentinel") {
			return "cli"
		} else if strings.Contains(path, "internal/agents") {
			return "agents"
		} else if strings.Contains(path, "internal/graph") {
			return "graph"
		} else if strings.Contains(path, "internal/audit") || strings.Contains(path, "internal/reflect") {
			return "audit"
		} else if strings.Contains(path, "internal/state") {
			return "state"
		} else if strings.Contains(path, "pkg/sqlite") {
			return "db"
		} else if strings.Contains(path, "legacy/ts") {
			return "frontend"
		}
		return ""
	}

	for _, n := range nodes {
		cid := getContainerID(n.FilePath)
		if cid == "" && strings.HasPrefix(n.ID, "file:") {
			cid = getContainerID(strings.TrimPrefix(n.ID, "file:"))
		}

		if cid != "" {
			nodeToContainer[n.ID] = cid
		}
	}

	// Adiciona mapeamento para alvos de imports que podem não ser nós de símbolos
	for _, e := range edges {
		if _, ok := nodeToContainer[e.To]; !ok && strings.HasPrefix(e.To, "file:") {
			path := strings.TrimPrefix(e.To, "file:")
			if cid := getContainerID(path); cid != "" {
				nodeToContainer[e.To] = cid
			}
		}
		if _, ok := nodeToContainer[e.From]; !ok && strings.HasPrefix(e.From, "file:") {
			path := strings.TrimPrefix(e.From, "file:")
			if cid := getContainerID(path); cid != "" {
				nodeToContainer[e.From] = cid
			}
		}
	}

	// Agrega relações entre containers
	type relKey struct {
		from, to string
	}
	rels := make(map[relKey]string)
	for _, e := range edges {
		fromC, okF := nodeToContainer[e.From]
		toC, okT := nodeToContainer[e.To]

		if okF && okT && fromC != toC {
			key := relKey{fromC, toC}
			rels[key] = e.Rel
		}
	}

	for k, rel := range rels {
		sb.WriteString(fmt.Sprintf("    Rel(%s, %s, \"%s\")\n", k.from, k.to, rel))
	}

	return sb.String()
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
