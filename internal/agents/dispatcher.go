package agents

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// Dispatcher handles sub-task assignment and event reconciliation (Write Serializer).
type Dispatcher struct {
	Registry *RegistryManager
	Shield   *GitShield
	DB       *sqlite.DB
}

// NewDispatcher initializes the orchestration engine.
func NewDispatcher(registry *RegistryManager, shield *GitShield, db *sqlite.DB) *Dispatcher {
	return &Dispatcher{
		Registry: registry,
		Shield:   shield,
		DB:       db,
	}
}

// Dispatch selects a specialist, creates a worktree, and registers the sub-task.
func (d *Dispatcher) Dispatch(ctx context.Context, st *SubTask) error {
	spec, err := d.Registry.SelectBest(ctx, st.RequiredCapabilities)
	if err != nil {
		return fmt.Errorf("dispatcher: could not select specialist: %w", err)
	}

	st.SpecialistID = spec.ID
	path, err := d.Shield.CreateWorktree(st.ID, st.BranchName)
	if err != nil {
		return fmt.Errorf("dispatcher: worktree creation failed: %w", err)
	}
	st.WorktreePath = path

	// Marshal capabilities to JSON for persistence
	capsJSON, err := json.Marshal(st.RequiredCapabilities)
	if err != nil {
		return fmt.Errorf("dispatcher: failed to marshal capabilities: %w", err)
	}

	// Persistir sub-task no Ledger Central (Apenas o Dispatcher escreve aqui)
	query := `
		INSERT INTO sub_tasks (id, parent_task_id, specialist_id, description, status, worktree_path, branch_name, required_capabilities)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			specialist_id=excluded.specialist_id,
			status=excluded.status,
			worktree_path=excluded.worktree_path,
			branch_name=excluded.branch_name,
			required_capabilities=excluded.required_capabilities,
			updated_at=CURRENT_TIMESTAMP`
	_, err = d.DB.Conn.ExecContext(ctx, query, st.ID, st.ParentTaskID, st.SpecialistID, st.Description, "DISPATCHED", st.WorktreePath, st.BranchName, string(capsJSON))
	if err != nil {
		// Best-effort rollback of worktree on DB failure
		_ = d.Shield.RemoveWorktree(st.WorktreePath)
		return fmt.Errorf("dispatcher: failed to log sub-task: %w", err)
	}

	return nil
}

// ReconcileEvents reads atomic event files from sub-agents and updates the central Ledger.
func (d *Dispatcher) ReconcileEvents(ctx context.Context) error {
	eventDir := ".sentinel/events"
	entries, err := os.ReadDir(eventDir)
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return fmt.Errorf("dispatcher: could not read events: %w", err)
	}

	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		path := filepath.Join(eventDir, entry.Name())
		file, err := os.Open(path)
		if err != nil {
			continue
		}

		// Standard #01: Use buffered reader (bufio) to read the event file
		reader := bufio.NewReader(file)
		data, err := io.ReadAll(reader)
		file.Close()
		if err != nil {
			continue
		}

		var event struct {
			SubTaskID string `json:"sub_task_id"`
			Status    string `json:"status"`
			Result    string `json:"result"`
		}

		if err := json.Unmarshal(data, &event); err != nil {
			continue
		}

		// Atualização Atômica no Ledger (Standard #13)
		_, err = d.DB.Conn.ExecContext(ctx, "UPDATE sub_tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?", event.Status, event.SubTaskID)
		if err != nil {
			return fmt.Errorf("dispatcher: reconciliation failed for %s: %w", event.SubTaskID, err)
		}

		if err := os.Remove(path); err != nil {
			fmt.Fprintf(os.Stderr, "warning: dispatcher: could not remove event file %s: %v\n", path, err)
		}
	}

	return nil
}
