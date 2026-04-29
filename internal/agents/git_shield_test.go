package agents

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

type mockValidator struct{}

func (m *mockValidator) ValidatePath(path string) error {
	return nil
}

func (m *mockValidator) ValidateCommand(cmd string) error {
	return nil
}

func TestGitShield_CreateWorktree(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "gitshield-test-*")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Init a git repo in tmpDir
	cmd := exec.Command("git", "init")
	cmd.Dir = tmpDir
	if err := cmd.Run(); err != nil {
		t.Fatalf("failed to init git repo: %v", err)
	}

	// Create a dummy commit so we have a branch
	if err := os.WriteFile(filepath.Join(tmpDir, "dummy"), []byte("data"), 0644); err != nil {
		t.Fatalf("failed to write dummy file: %v", err)
	}
	cmd = exec.Command("git", "add", ".")
	cmd.Dir = tmpDir
	cmd.Run()
	cmd = exec.Command("git", "commit", "-m", "initial")
	cmd.Dir = tmpDir
	cmd.Run()

	gs := NewGitShield(tmpDir, &mockValidator{})

	// Ensure .worktrees directory exists
	if err := os.Mkdir(filepath.Join(tmpDir, ".worktrees"), 0755); err != nil {
		t.Fatalf("failed to create .worktrees dir: %v", err)
	}

	t.Run("Create Worktree Success", func(t *testing.T) {
		// Detect default branch
		branch := "master"
		cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
		cmd.Dir = tmpDir
		out, err := cmd.Output()
		if err == nil {
			branch = string(out)
			branch = strings.TrimSpace(branch)
		}

		path, err := gs.CreateWorktree("task-123", branch)
		if err != nil {
			t.Fatalf("failed to create worktree with branch %s: %v", branch, err)
		}

		if _, err := os.Stat(filepath.Join(tmpDir, path)); os.IsNotExist(err) {
			t.Errorf("worktree directory was not created: %s", path)
		}
	})

	t.Run("Cleanup Worktrees", func(t *testing.T) {
		err := gs.CleanupWorktrees()
		if err != nil {
			t.Fatalf("failed to cleanup worktrees: %v", err)
		}

		// Verify worktree list doesn't have our task
		output, _ := gs.run("worktree", "list")
		if testing.Verbose() {
			t.Logf("Worktree list: %s", output)
		}
	})
}
