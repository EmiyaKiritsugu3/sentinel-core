package report

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type TaskInfo struct {
	state.Task
	ADRPath string
}

type ProjectStats struct {
	TotalNodes     int
	TotalFiles     int
	TotalFunctions int
	TotalStructs   int
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	SuccessRate    float64
	Tasks          []TaskInfo
}

type Aggregator struct {
	db *sqlite.DB
}

func NewAggregator(db *sqlite.DB) *Aggregator {
	return &Aggregator{db: db}
}

// FetchStats consolida todos os dados do SQLite
func (a *Aggregator) FetchStats() (*ProjectStats, error) {
	stats := &ProjectStats{}

	// 1. Contagem de Nós
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&stats.TotalNodes)
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'file'").Scan(&stats.TotalFiles)
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'function'").Scan(&stats.TotalFunctions)
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes WHERE type = 'struct'").Scan(&stats.TotalStructs)

	// 2. Contagem de Tasks
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&stats.TotalTasks)
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'DONE'").Scan(&stats.CompletedTasks)
	a.db.Conn.QueryRow("SELECT COUNT(*) FROM tasks WHERE status = 'FAILED'").Scan(&stats.FailedTasks)

	// 3. Cálculo de Success Rate
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100
	}

	// 4. Listagem Detalhada de Tasks (Sovereign Link Discovery)
	mgr := state.NewManager(a.db)
	tasks, err := mgr.ListTasks()
	if err != nil {
		return nil, err
	}

	for _, t := range tasks {
		info := TaskInfo{Task: t}
		// Tenta encontrar o ADR via padrão no disco
		pattern := filepath.Join("docs/architecture/adr", fmt.Sprintf("ADR-%s-*.md", t.ID))
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			info.ADRPath = matches[0]
		}
		stats.Tasks = append(stats.Tasks, info)
	}

	return stats, nil
}

// GenerateMarkdown gera o arquivo de dashboard persistente
func (a *Aggregator) GenerateMarkdown(stats *ProjectStats) error {
	content := "# Sentinel Compliance Dashboard 📊 [PID-SENTINEL]\n\n"
	content += "> [!NOTE]\n> Este relatório é gerado automaticamente pelo Guardião.\n\n"

	content += "## 🏁 Key Performance Indicators (KPIs)\n\n"
	content += "| Métrica | Valor |\n"
	content += "| :--- | :--- |\n"
	content += fmt.Sprintf("| **Engineering Success Rate** | %.2f%% |\n", stats.SuccessRate)
	content += fmt.Sprintf("| **Total Architecture Nodes** | %d |\n", stats.TotalNodes)
	content += fmt.Sprintf("| **Files Tracked** | %d |\n", stats.TotalFiles)
	content += fmt.Sprintf("| **Functions & Structs** | %d |\n", stats.TotalFunctions+stats.TotalStructs)

	content += "\n## 🛡️ Task Lifecycle Status\n\n"
	content += fmt.Sprintf("- ✅ **Completed**: %d\n", stats.CompletedTasks)
	content += fmt.Sprintf("- 🛑 **Failed**: %d\n", stats.FailedTasks)
	content += fmt.Sprintf("- 🕒 **Total Attempts**: %d\n", stats.TotalTasks)

	content += "\n## 📝 Detailed Intent Inventory\n\n"
	content += "| ID | Tier | Status | Description | Decision Record |\n"
	content += "| :--- | :--- | :--- | :--- | :--- |\n"
	for _, t := range stats.Tasks {
		adrLink := "N/A"
		if t.ADRPath != "" {
			relPath, _ := filepath.Rel("docs/process", t.ADRPath)
			adrLink = fmt.Sprintf("[View ADR](%s)", relPath)
		}
		content += fmt.Sprintf("| `%s` | %s | %s | %s | %s |\n", t.ID, t.Tier, t.Status, t.Description, adrLink)
	}

	path := "docs/process/COMPLIANCE-DASHBOARD.md"
	os.MkdirAll(filepath.Dir(path), 0755)
	return os.WriteFile(path, []byte(content), 0644)
}
