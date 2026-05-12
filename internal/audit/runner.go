package audit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/shlex"
)

type Runner struct {
	db *sqlite.DB
}

func NewRunner(db *sqlite.DB) (*Runner, error) {
	if err := sqlite.ValidateDB(db, "audit-runner"); err != nil {
		return nil, err
	}
	return &Runner{db: db}, nil
}

// ExecuteAudit roda o comando de verificação para uma tarefa específica com timeout e proteção de shell
func (r *Runner) ExecuteAudit(taskID string, command string) (bool, error) {
	fmt.Printf("🛡️ Sentinel: Auditing Task [%s]...\n", taskID)

	args, err := shlex.Split(command)
	if err != nil {
		return false, fmt.Errorf("audit: failed to parse command: %w", err)
	}
	if len(args) == 0 {
		return false, fmt.Errorf("audit: empty verification command")
	}

	fmt.Printf("Executing: %v (Timeout: 30s)\n", args)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...)
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	exitCode := 0
	success := true

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			fmt.Println("🛑 ERROR: Audit Timeout Exceeded.")
			exitCode = 124
		} else {
			// Uso do errors.As para detecção robusta de erro de saída
			var exitError *exec.ExitError
			if errors.As(err, &exitError) {
				exitCode = exitError.ExitCode()
			} else {
				exitCode = 1
			}
		}
		success = false
	}

	logQuery := `INSERT INTO audit_logs (task_id, command, output, exit_code) VALUES (?, ?, ?, ?)`
	_, dbErr := r.db.Conn.Exec(logQuery, taskID, command, out.String(), exitCode)
	if dbErr != nil {
		return false, fmt.Errorf("audit: failed to save log for task %s: %w", taskID, dbErr)
	}

	if success {
		fmt.Println("✅ Audit Passed. Gate open.")
	} else {
		fmt.Printf("❌ Audit Failed (Exit Code: %d). Gate locked.\n", exitCode)
	}

	return success, nil
}
