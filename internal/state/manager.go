package state

import (
	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/uuid"
)

type Task struct {
	ID          string
	Description string
	Status      string
	Tier        string
}

type Manager struct {
	db *sqlite.DB
}

func NewManager(db *sqlite.DB) *Manager {
	return &Manager{db: db}
}

// CreateTask cria uma nova tarefa no banco
func (m *Manager) CreateTask(description string, tier string) (string, error) {
	id := uuid.New().String()[:8] // ID curto para facilidade no CLI
	query := `INSERT INTO tasks (id, description, status, tier) VALUES (?, ?, ?, ?)`
	_, err := m.db.Conn.Exec(query, id, description, "PENDING", tier)
	if err != nil {
		return "", err
	}
	return id, nil
}

// UpdateStatus muda o estado da tarefa garantindo que não se pule etapas
func (m *Manager) UpdateStatus(id string, nextStatus string) error {
	// Aqui poderíamos adicionar a lógica de "Gate": 
	// Só pode ir para DONE se o Audit_Runner der OK.
	query := `UPDATE tasks SET status = ?, updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.Conn.Exec(query, nextStatus, id)
	return err
}

// GetActiveTask retorna a tarefa que está em progresso
func (m *Manager) GetActiveTask() (*Task, error) {
	query := `SELECT id, description, status, tier FROM tasks WHERE status = 'IN_PROGRESS' LIMIT 1`
	var t Task
	err := m.db.Conn.QueryRow(query).Scan(&t.ID, &t.Description, &t.Status, &t.Tier)
	if err != nil {
		return nil, err
	}
	return &t, nil
}
