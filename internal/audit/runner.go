package audit

import (
	"bytes"
	"fmt"
	"os/exec"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type Runner struct {
	db *sqlite.DB
}

func NewRunner(db *sqlite.DB) *Runner {
	return &Runner{db: db}
}

// ExecuteAudit roda o comando de verificação para uma tarefa específica
func (r *Runner) ExecuteAudit(taskID string, command string) (bool, error) {
	fmt.Printf("🛡️ Sentinel: Auditing Task [%s]...\n", taskID)
	fmt.Printf("Executing: %s\n", command)

	// Executa o comando de shell
	cmd := exec.Command("sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	exitCode := 0
	success := true

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
		success = false
	}

	// Salva o Log de Auditoria no SQLite
	logQuery := `INSERT INTO audit_logs (task_id, command, output, exit_code) VALUES (?, ?, ?, ?)`
	_, dbErr := r.db.Conn.Exec(logQuery, taskID, command, out.String(), exitCode)
	if dbErr != nil {
		return false, fmt.Errorf("failed to save audit log: %w", dbErr)
	}

	if success {
		fmt.Println("✅ Audit Passed. Gate open.")
	} else {
		fmt.Printf("❌ Audit Failed (Exit Code: %d). Gate locked.\n", exitCode)
	}

	return success, nil
}
