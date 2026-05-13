package agents

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

type mockValidator struct{}

func (m *mockValidator) ValidatePath(path string) error   { return nil }
func (m *mockValidator) ValidateCommand(cmd string) error { return nil }

func TestGitShield_CreateWorktree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitshield-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer func() { _ = os.RemoveAll(tmpDir) }()

	// Init a git repo and initial commit
	runCmd := func(args ...string) {
		c := exec.CommandContext(context.Background(), "git", args...) //nolint:gosec // test fixture
		c.Dir = tmpDir
		if out, err := c.CombinedOutput(); err != nil {
			t.Fatalf("git %v failed: %v (output: %s)", args, err, out)
		}
	}
	runCmd("init")
	_ = os.WriteFile(filepath.Join(tmpDir, "dummy"), []byte("data"), 0644) //nolint:gosec // test fixture
	runCmd("add", ".")
	runCmd("commit", "-m", "initial")

	// Create a branch that is NOT checked out
	runCmd("branch", "subtask-branch")

	gs := NewGitShield(tmpDir, &mockValidator{})
	_ = os.Mkdir(filepath.Join(tmpDir, ".worktrees"), 0755) //nolint:gosec // test fixture

	t.Run("Create Worktree Success", func(t *testing.T) {
		path, err := gs.CreateWorktree("task-123", "subtask-branch")
		if err != nil {
			t.Fatalf("failed to create worktree: %v", err)
		}

		fullPath := filepath.Join(tmpDir, path)
		if _, err := os.Stat(fullPath); os.IsNotExist(err) {
			t.Errorf("worktree directory was not created: %s", fullPath)
		}
	})

	t.Run("Cleanup Worktrees", func(t *testing.T) {
		err := gs.CleanupWorktrees()
		if err != nil {
			t.Fatalf("failed to cleanup worktrees: %v", err)
		}
	})
}
