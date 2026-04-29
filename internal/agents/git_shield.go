package agents

import (
	"fmt"
	"os/exec"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/utils"
)

// GitShield handles task-specific branch management and atomic commits (Standard #10).
type GitShield struct {
	WorkingDir string
	validator  Validator
}

// NewGitShield initializes a new GitShield agent.
func NewGitShield(workingDir string, v Validator) *GitShield {
	return &GitShield{WorkingDir: workingDir, validator: v}
}

// run executes a git command directly without shell wrapping (Standard #10).
func (g *GitShield) run(args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = g.WorkingDir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("git: command 'git %s' failed: %w (output: %s)", strings.Join(args, " "), err, string(output))
	}
	return string(output), nil
}

// CreateTaskBranch creates and switches to a sanitized task-specific branch.
func (g *GitShield) CreateTaskBranch(taskID string) (string, error) {
	slug := utils.Slugify(taskID)
	branchName := fmt.Sprintf("sentinel/task-%s", slug)

	_, err := g.run("checkout", "-b", branchName)
	if err != nil {
		// If branch exists, just switch to it
		if strings.Contains(err.Error(), "already exists") {
			_, err = g.run("checkout", branchName)
			if err != nil {
				return "", err
			}
			return branchName, nil
		}
		return "", err
	}
	return branchName, nil
}

// CreateWorktree creates a new isolated worktree for a task.
func (g *GitShield) CreateWorktree(taskID string, branch string) (string, error) {
	slug := utils.Slugify(taskID)
	path := fmt.Sprintf(".worktrees/sentinel-task-%s", slug)

	// Standard #10: Security - validate path before execution
	if err := g.validator.ValidatePath(path); err != nil {
		return "", fmt.Errorf("git: invalid worktree path: %w", err)
	}

	_, err := g.run("worktree", "add", path, branch)
	if err != nil {
		return "", fmt.Errorf("git: failed to create worktree: %w", err)
	}

	return path, nil
}

// CleanupWorktrees removes all sentinel-task worktrees (Sovereign GC).
func (g *GitShield) CleanupWorktrees() error {
	output, err := g.run("worktree", "list", "--porcelain")
	if err != nil {
		return err
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if strings.HasPrefix(line, "worktree ") {
			path := strings.TrimPrefix(line, "worktree ")
			if strings.Contains(path, "sentinel-task-") {
				if _, err := g.run("worktree", "remove", "--force", path); err != nil {
					// Log error but continue (Standard #05: Error governance - logging for async-like cleanup)
					continue
				}
			}
		}
	}
	return nil
}

// AtomicCommit stages all changes and creates a commit with the given message.
func (g *GitShield) AtomicCommit(message string) error {
	if _, err := g.run("add", "."); err != nil {
		return err
	}

	if _, err := g.run("commit", "-m", message); err != nil {
		// Handle "nothing to commit" case gracefully
		if strings.Contains(err.Error(), "nothing to commit") {
			return nil
		}
		return err
	}
	return nil
}
