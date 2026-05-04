# Git Shield & Ephemeral Branches Implementation Plan [PID-SENTINEL]

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement the `GitShield` component in Go to handle task-specific branch creation and atomic commits safely, adhering to Project Sentinel's elite engineering standards.

**Architecture:** A dedicated `GitShield` struct in `internal/agents` that abstracts `git` CLI operations using shell-less execution (`exec.Command` directly with binary). It uses a `WorkingDir` to ensure operations occur in the correct project context and leverages `pkg/utils` for sanitization.

**Tech Stack:** Go (Standard Library `os/exec`).

---

## Task 1: Basic Structure and Initialization

**Files:**

- Create: `internal/agents/git_shield.go`

- [ ] **Step 1: Implement GitShield struct and constructor**
Define `GitShield` with `WorkingDir` and a constructor that defaults to the current directory if empty.

```go
package agents

import (
 "fmt"
 "os/exec"
 "strings"
 "home/emiyakiritsugu/Projetos_Antigravity/sentinel-core/pkg/utils"
)

type GitShield struct {
 WorkingDir string
}

func NewGitShield(workingDir string) *GitShield {
 return &GitShield{WorkingDir: workingDir}
}
```

- [ ] **Step 2: Implement run helper for Standard #10**
Add a private helper to execute git commands directly without shell wrapping.

```go
func (g *GitShield) run(args ...string) (string, error) {
 cmd := exec.Command("git", args...)
 cmd.Dir = g.WorkingDir
 output, err := cmd.CombinedOutput()
 if err != nil {
  return "", fmt.Errorf("git: command 'git %s' failed: %w (output: %s)", strings.Join(args, " "), err, string(output))
 }
 return string(output), nil
}
```

### Task 2: Task Branch Creation

**Files:**

- Modify: `internal/agents/git_shield.go`

- [ ] **Step 1: Implement CreateTaskBranch**
Use `utils.Slugify` to sanitize the taskID and create/checkout the branch.

```go
func (g *GitShield) CreateTaskBranch(taskID string) (string, error) {
 slug := utils.Slugify(taskID)
 branchName := fmt.Sprintf("sentinel/task-%s", slug)
 
 _, err := g.run("checkout", "-b", branchName)
 if err != nil {
  return "", err
 }
 return branchName, nil
}
```

### Task 3: Atomic Commit Implementation

**Files:**

- Modify: `internal/agents/git_shield.go`

- [ ] **Step 1: Implement AtomicCommit**
Run `git add .` and `git commit -m {message}` sequentially.

```go
func (g *GitShield) AtomicCommit(message string) error {
 if _, err := g.run("add", "."); err != nil {
  return err
 }
 
 if _, err := g.run("commit", "-m", message); err != nil {
  return err
 }
 return nil
}
```

### Task 4: Verification and Compilation

**Files:**

- N/A

- [ ] **Step 1: Run compilation check**
Run: `go build ./internal/agents/...`
Expected: Success

- [ ] **Step 2: Commit changes**
Commit message: `feat(agents): implement git shield with ephemeral branches`
