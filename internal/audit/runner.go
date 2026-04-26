package audit

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/shlex"
)

type Runner struct {
	db *sqlite.DB
}

func NewRunner(db *sqlite.DB) *Runner {
	return &Runner{db: db}
}

// ExecuteAudit roda o comando de verificação para uma tarefa específica com timeout e proteção de shell
func (r *Runner) ExecuteAudit(taskID string, command string) (bool, error) {
	fmt.Printf("🛡️ Sentinel: Auditing Task [%s]...\n", taskID)

	// 1. Safe Parsing: Remove dependência de shell e evita injeção
	args, err := shlex.Split(command)
	if err != nil {
		return false, fmt.Errorf("audit: failed to parse command: %w", err)
	}
	if len(args) == 0 {
		return false, fmt.Errorf("audit: empty verification command")
	}

	fmt.Printf("Executing: %v (Timeout: 30s)\n", args)

	// Cria contexto com timeout de 30 segundos
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// 2. Invocação Direta: Sem wrapper de shell
	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	exitCode := 0
	success := true

	if err != nil {
		if ctx.Err() == context.DeadlineExceeded {
			fmt.Println("🛑 ERROR: Audit Timeout Exceeded.")
			exitCode = 124 
		} else if exitError, ok := err.(*exec.ExitError); ok {
			exitCode = exitError.ExitCode()
		} else {
			exitCode = 1
		}
		success = false
	}

	// 3. Registro de Auditoria
	logQuery := `INSERT INTO audit_logs (task_id, command, output, exit_code) VALUES (?, ?, ?, ?)`
	_, dbErr := r.db.Conn.Exec(logQuery, taskID, command, out.String(), exitCode)
	if dbErr != nil {
		return false, fmt.Errorf("audit: failed to save log: %w", dbErr)
	}

	if success {
		fmt.Println("✅ Audit Passed. Gate open.")
	} else {
		fmt.Printf("❌ Audit Failed (Exit Code: %d). Gate locked.\n", exitCode)
	}

	return success, nil
}
