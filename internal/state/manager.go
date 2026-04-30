package state

import (
	"fmt"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/uuid"
)

type Task struct {
	ID          string
	Description string
	Status      string
	Tier        string
	CreatedAt   time.Time
}

type Manager struct {
	db *sqlite.DB
}

func NewManager(db *sqlite.DB) *Manager {
	return &Manager{db: db}
}

// CreateTask cria uma nova tarefa no banco
func (m *Manager) CreateTask(description string, tier string, verificationCmd string) (string, error) {
	id := uuid.New().String()[:8]
	query := `INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)`
	_, err := m.db.Conn.Exec(query, id, description, "PENDING", tier, verificationCmd)
	if err != nil {
		return "", fmt.Errorf("state: failed to create task: %w", err)
	}
	return id, nil
}

// StartTask marca a tarefa como em progresso
func (m *Manager) StartTask(id string) error {
	query := `UPDATE tasks SET status = 'IN_PROGRESS', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.Conn.Exec(query, id)
	if err != nil {
		return fmt.Errorf("state: failed to start task %s: %w", id, err)
	}
	return nil
}

// GetTaskByID busca uma tarefa específica
func (m *Manager) GetTaskByID(id string) (*Task, string, error) {
	query := `SELECT id, description, status, tier, verification_command FROM tasks WHERE id = ?`
	var t Task
	var cmd string
	err := m.db.Conn.QueryRow(query, id).Scan(&t.ID, &t.Description, &t.Status, &t.Tier, &cmd)
	if err != nil {
		return nil, "", fmt.Errorf("state: task %s not found: %w", id, err)
	}
	return &t, cmd, nil
}

// UpdateStatus muda o estado da tarefa
func (m *Manager) UpdateStatus(id string, nextStatus string) error {
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.Conn.Exec(query, nextStatus, id)
	if err != nil {
		return fmt.Errorf("state: failed to update status of task %s to %s: %w", id, nextStatus, err)
	}
	return nil
}

// GetActiveTask retorna a tarefa que está em progresso
func (m *Manager) GetActiveTask() (*Task, error) {
	query := `SELECT id, description, status, tier FROM tasks WHERE status = 'IN_PROGRESS' LIMIT 1`
	var t Task
	err := m.db.Conn.QueryRow(query).Scan(&t.ID, &t.Description, &t.Status, &t.Tier)
	if err != nil {
		return nil, fmt.Errorf("state: no active task: %w", err)
	}
	return &t, nil
}

// ListTasks retorna todas as tarefas registradas no banco
func (m *Manager) ListTasks() ([]Task, error) {
	query := `SELECT id, description, status, tier, created_at FROM tasks ORDER BY created_at DESC`
	rows, err := m.db.Conn.Query(query)
	if err != nil {
		return nil, fmt.Errorf("state: failed to list tasks: %w", err)
	}
	defer rows.Close()

	var tasks []Task
	for rows.Next() {
		var t Task
		var createdAt string
		if err := rows.Scan(&t.ID, &t.Description, &t.Status, &t.Tier, &createdAt); err != nil {
			return nil, fmt.Errorf("state: failed to scan task: %w", err)
		}
		// SQLite CURRENT_TIMESTAMP é "YYYY-MM-DD HH:MM:SS"
		t.CreatedAt, _ = time.Parse("2006-01-02 15:04:05", createdAt)
		tasks = append(tasks, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("state: row iteration error: %w", err)
	}

	return tasks, nil
}
