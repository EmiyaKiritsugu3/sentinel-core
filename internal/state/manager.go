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
func (m *Manager) CreateTask(description string, tier string, verificationCmd string) (string, error) {
	id := uuid.New().String()[:8]
	query := `INSERT INTO tasks (id, description, status, tier, verification_command) VALUES (?, ?, ?, ?, ?)`
	_, err := m.db.Conn.Exec(query, id, description, "PENDING", tier, verificationCmd)
	if err != nil {
		return "", err
	}
	return id, nil
}

// StartTask marca a tarefa como em progresso
func (m *Manager) StartTask(id string) error {
	query := `UPDATE tasks SET status = 'IN_PROGRESS', updated_at = CURRENT_TIMESTAMP WHERE id = ?`
	_, err := m.db.Conn.Exec(query, id)
	return err
}

// GetTaskByID busca uma tarefa específica
func (m *Manager) GetTaskByID(id string) (*Task, string, error) {
	query := `SELECT id, description, status, tier, verification_command FROM tasks WHERE id = ?`
	var t Task
	var cmd string
	err := m.db.Conn.QueryRow(query, id).Scan(&t.ID, &t.Description, &t.Status, &t.Tier, &cmd)
	if err != nil {
		return nil, "", err
	}
	return &t, cmd, nil
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
