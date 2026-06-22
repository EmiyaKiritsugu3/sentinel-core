// Package report generates compliance dashboards and project statistics.
package report

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/state"
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// TaskInfo wraps a state.Task with its ADR path.
type TaskInfo struct {
	state.Task
	ADRPath string
}

// ProjectStats aggregates project compliance statistics.
type ProjectStats struct {
	TotalNodes     int
	TotalFiles     int
	TotalFunctions int
	TotalStructs   int
	TotalTasks     int
	CompletedTasks int
	FailedTasks    int
	SuccessRate    float64
	AvgMathDelta   float64
	Tasks          []TaskInfo
}

// Aggregator collects and reports project statistics.
type Aggregator struct {
	db *sqlite.DB
}

// NewAggregator creates a new Aggregator with the given DB.
func NewAggregator(db *sqlite.DB) (*Aggregator, error) {
	if err := sqlite.ValidateDB(db, "report-aggregator"); err != nil {
		return nil, err
	}
	return &Aggregator{db: db}, nil
}

// FetchStats consolidates all SQLite data
func (a *Aggregator) FetchStats(ctx context.Context) (*ProjectStats, error) {
	stats := &ProjectStats{}

	// 1. Consolidate Node counts into a single query to fix N+1 issue
	nodeQuery := `
		SELECT
			COUNT(*),
			SUM(CASE WHEN type = 'file' THEN 1 ELSE 0 END),
			SUM(CASE WHEN type = 'function' THEN 1 ELSE 0 END),
			SUM(CASE WHEN type = 'struct' THEN 1 ELSE 0 END)
		FROM nodes`

	var totalFiles, totalFunctions, totalStructs sql.NullInt64
	if err := a.db.Conn.QueryRowContext(ctx, nodeQuery).Scan(&stats.TotalNodes, &totalFiles, &totalFunctions, &totalStructs); err != nil {
		return nil, fmt.Errorf("aggregator: failed to count nodes: %w", err)
	}

	stats.TotalFiles = int(totalFiles.Int64)
	stats.TotalFunctions = int(totalFunctions.Int64)
	stats.TotalStructs = int(totalStructs.Int64)

	// 2. Consolidate Task counts and avg into a single query
	taskQuery := `
		SELECT
			COUNT(*),
			SUM(CASE WHEN status = 'DONE' THEN 1 ELSE 0 END),
			SUM(CASE WHEN status = 'FAILED' THEN 1 ELSE 0 END),
			AVG(CASE WHEN status = 'DONE' THEN math_delta ELSE NULL END)
		FROM tasks`

	var completedTasks, failedTasks sql.NullInt64
	var avgDelta sql.NullFloat64
	if err := a.db.Conn.QueryRowContext(ctx, taskQuery).Scan(&stats.TotalTasks, &completedTasks, &failedTasks, &avgDelta); err != nil {
		return nil, fmt.Errorf("aggregator: failed to count tasks: %w", err)
	}

	stats.CompletedTasks = int(completedTasks.Int64)
	stats.FailedTasks = int(failedTasks.Int64)

	// 3. Success Rate and SME calculation
	if stats.TotalTasks > 0 {
		stats.SuccessRate = float64(stats.CompletedTasks) / float64(stats.TotalTasks) * 100

		if avgDelta.Valid {
			stats.AvgMathDelta = avgDelta.Float64
		}
	}

	// 4. Detailed Task Listing (Sovereign Link Discovery)
	mgr, err := state.NewManager(a.db)
	if err != nil {
		return nil, fmt.Errorf("aggregator: failed to create manager: %w", err)
	}
	tasks, err := mgr.ListTasks(ctx)
	if err != nil {
		return nil, fmt.Errorf("aggregator: failed to list tasks: %w", err)
	}

	for _, t := range tasks {
		info := TaskInfo{Task: t}
		// Attempts to find ADR via pattern on disk
		pattern := filepath.Join("docs/architecture/adr", fmt.Sprintf("ADR-%s-*.md", t.ID))
		matches, _ := filepath.Glob(pattern)
		if len(matches) > 0 {
			info.ADRPath = matches[0]
		}
		stats.Tasks = append(stats.Tasks, info)
	}

	return stats, nil
}

// GenerateMarkdown generates the persistence dashboard file
func (a *Aggregator) GenerateMarkdown(stats *ProjectStats) error {
	content := "# Sentinel Compliance Dashboard 📊 [PID-SENTINEL]\n\n"
	content += "> [!NOTE]\n> Este relatório é gerado automaticamente pelo Guardião.\n\n"

	content += "## 🏁 Key Performance Indicators (KPIs)\n\n"
	content += "| Métrica | Valor |\n"
	content += "| :--- | :--- |\n"
	content += fmt.Sprintf("| **Engineering Success Rate** | %.2f%% |\n", stats.SuccessRate)
	content += fmt.Sprintf("| **Sovereign Math Engine (Δ)** | %+.2f |\n", stats.AvgMathDelta)
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
	if err := os.MkdirAll(filepath.Dir(path), 0750); err != nil {
		return fmt.Errorf("aggregate: failed to create directory: %w", err)
	}
	return os.WriteFile(path, []byte(content), 0600)
}
