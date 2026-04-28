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
}

// NewGitShield initializes a new GitShield agent.
func NewGitShield(workingDir string) *GitShield {
	return &GitShield{WorkingDir: workingDir}
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
