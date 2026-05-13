// Package audit provides compliance verification and gate execution.
package audit

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os/exec"
	"time"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
	"github.com/google/shlex"
)

// Runner runs audit verifications and records results.
type Runner struct {
	db *sqlite.DB
}

// NewRunner creates a new Runner with the given DB.
func NewRunner(db *sqlite.DB) (*Runner, error) {
	if err := sqlite.ValidateDB(db, "audit-runner"); err != nil {
		return nil, err
	}
	return &Runner{db: db}, nil
}

// ExecuteAudit runs the verification command for a specific task with timeout and shell protection
func (r *Runner) ExecuteAudit(taskID string, command string) (bool, error) {
	slog.Info("auditing task", "task", taskID)

	args, err := shlex.Split(command)
	if err != nil {
		return false, fmt.Errorf("audit: failed to parse command: %w", err)
	}
	if len(args) == 0 {
		return false, fmt.Errorf("audit: empty verification command")
	}

	slog.Info("executing verification", "command", args, "timeout", "30s")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	cmd := exec.CommandContext(ctx, args[0], args[1:]...) //nolint:gosec // intentional: command from ADR
	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err = cmd.Run()
	exitCode := 0
	success := true

	if err != nil {
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			slog.Error("audit timeout exceeded")
			exitCode = 124
		} else {
			// Use errors.As for robust exit error detection
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
		slog.Info("audit passed", "gate", "open")
	} else {
		slog.Error("audit failed", "exit_code", exitCode)
	}

	return success, nil
}
