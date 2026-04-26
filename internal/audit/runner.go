package audit

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type Runner struct {
	db *sqlite.DB
}

func NewRunner(db *sqlite.DB) *Runner {
	return &Runner{db: db}
}

// ExecuteAudit roda o comando de verificação para uma tarefa específica com timeout
func (r *Runner) ExecuteAudit(taskID string, command string) (bool, error) {
	fmt.Printf("🛡️ Sentinel: Auditing Task [%s]...\n", taskID)
	fmt.Printf("Executing: %s (Timeout: 30s)\n", command)

	// Cria contexto com timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Executa o comando de shell usando o contexto
	cmd := exec.CommandContext(ctx, "sh", "-c", command)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()
	exitCode := 0
	success := true

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("🛑 ERROR: Audit Timeout Exceeded.")
			exitCode = 124 // Padrão coreutils para timeout
		} else if exitError, ok := err.(*exec.ExitError); ok {
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
