// Package state manages the task lifecycle and persistence.
package state

import (
	"context"
	"fmt"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/uuid"
)

// Task represents a tracked work item.
type Task struct {
	ID          string
	Description string
	Status      string
	Tier        string
	CreatedAt   time.Time
}

// Manager manages task lifecycle and persistence.
type Manager struct {
	db *sqlite.DB
}

// NewManager creates a new Manager with the given DB.
func NewManager(db *sqlite.DB) (*Manager, error) {
	if err := sqlite.ValidateDB(db, "state-manager"); err != nil {
		return nil, err
	}
	return &Manager{db: db}, nil
}

// CreateTask creates a new task in the database
func (m *Manager) CreateTask(ctx context.Context, description string, tier string, verificationCmd string) (string, error) {
	id := uuid.New().String()
	query := `INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)`
	_, err := m.db.Conn.ExecContext(ctx, query, id, description, "PENDING", tier, verificationCmd)
	if err != nil {
		return "", fmt.Errorf("state: failed to create task: %w", err)
	}
	return id, nil
}

// StartTask marks the task as in progress
func (m *Manager) StartTask(ctx context.Context, id string) error {
	query := `UPDATE tasks SET status = 'IN_PROGRESS', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.Conn.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("state: failed to start task %s: %w", id, err)
	}
	return nil
}

// GetTaskByID fetches a specific task
func (m *Manager) GetTaskByID(ctx context.Context, id string) (*Task, string, error) {
	query := `SELECT id, description, status, tier, verification_command FROM tasks WHERE id = ?`
	var t Task
	var cmd string
	err := m.db.Conn.QueryRowContext(ctx, query, id).Scan(&t.ID, &t.Description, &t.Status, &t.Tier, &cmd)
	if err != nil {
		return nil, "", fmt.Errorf("state: task %s not found: %w", id, err)
	}
	return &t, cmd, nil
}

// UpdateStatus muda o estado da tarefa
func (m *Manager) UpdateStatus(ctx context.Context, id string, nextStatus string) error {
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.Conn.ExecContext(ctx, query, nextStatus, id)
	if err != nil {
		return fmt.Errorf("state: failed to update status of task %s to %s: %w", id, nextStatus, err)
	}
	return nil
}

// GetActiveTask returns the task that is in progress
func (m *Manager) GetActiveTask(ctx context.Context) (*Task, error) {
	query := `SELECT id, description, status, tier FROM tasks WHERE status = 'IN_PROGRESS' LIMIT 1`
	var t Task
	err := m.db.Conn.QueryRowContext(ctx, query).Scan(&t.ID, &t.Description, &t.Status, &t.Tier)
	if err != nil {
		return nil, fmt.Errorf("state: no active task: %w", err)
	}
	return &t, nil
}

// ListTasks retorna todas as tarefas registradas no banco
func (m *Manager) ListTasks(ctx context.Context) ([]Task, error) {
	query := `SELECT id, description, status, tier, created_at FROM tasks ORDER BY created_at DESC`
	rows, err := m.db.Conn.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("state: failed to list tasks: %w", err)
	}
	defer func() { _ = rows.Close() }()

	var tasks []Task
	for rows.Next() {
		var t Task
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Description, &t.Status, &t.Tier, &createdAt); err != nil {
			return nil, fmt.Errorf("state: failed to scan task: %w", err)
		}
		// SQLite CURRENT_TIMESTAMP is "YYYY-MM-DD HH:MM:SS"
		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("state: row iteration error: %w", err)
	}

	return tasks, nil
}
