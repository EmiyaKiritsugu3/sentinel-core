// Package reflect provides structural validation and compliance checking.
package reflect

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// Violation represents a standard violation found during validation.
type Violation struct {
	StandardID string
	FilePath   string
	Line       int
	Reason     string
}

// Validator checks project files against Sentinel standards.
type Validator struct {
	db *sqlite.DB
}

// NewValidator creates a new Validator with the given DB.
func NewValidator(db *sqlite.DB) (*Validator, error) {
	if err := sqlite.ValidateDB(db, "reflect-validator"); err != nil {
		return nil, err
	}
	return &Validator{db: db}, nil
}

// ValidateProject varre o projeto em busca de violações de Standards
func (v *Validator) ValidateProject(root string) ([]Violation, error) {
	var violations []Violation

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return fmt.Errorf("validator: walk error at %s: %w", path, err)
		}
		if info.IsDir() || isIgnored(path) {
			return nil
		}

		fileViolations, err := v.checkFile(path)
		if err != nil {
			return fmt.Errorf("validator: check failed for %s: %w", path, err)
		}
		violations = append(violations, fileViolations...)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("validator: project validation failed: %w", err)
	}

	return violations, nil
}

// ValidatePath garante que o caminho fornecido pelo agente é seguro (Standard #10).
func (v *Validator) ValidatePath(path string) error {
	cleanPath := filepath.Clean(path)

	// 1. Bloqueia caminhos absolutos
	if filepath.IsAbs(cleanPath) {
		return fmt.Errorf("security: absolute paths are forbidden: %s", path)
	}

	// 2. Bloqueia tentativa de sair do diretório do projeto (Path Traversal)
	if strings.HasPrefix(cleanPath, "..") {
		return fmt.Errorf("security: path traversal attempt detected: %s", path)
	}

	return nil
}

// ValidateCommand valida se o commando shell é permitido e não contém injeções.
func (v *Validator) ValidateCommand(cmd string) error {
	forbidden := []string{"|", "&&", ";", ">", ">>", "<", "`", "$("}
	for _, char := range forbidden {
		if strings.Contains(cmd, char) {
			return fmt.Errorf("security: forbidden shell character '%s' in command: %s", char, cmd)
		}
	}
	return nil
}

func (v *Validator) checkFile(path string) ([]Violation, error) {
	file, err := os.Open(path) //nolint:gosec // path from caller
	if err != nil {
		return nil, fmt.Errorf("validator: failed to open file %s: %w", path, err)
	}
	defer func() { _ = file.Close() }()

	// Quis custodiet ipsos custodes?
	isValidatorItself := strings.Contains(path, "internal/reflect/validator.go")

	var violations []Violation
	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()

		// Standard #01: Anti-os.ReadFile
		if !isValidatorItself && strings.Contains(line, "os.ReadFile") && !strings.Contains(path, "legacy") {
			violations = append(violations, Violation{
				StandardID: "STD-01",
				FilePath:   path,
				Line:       lineNum,
				Reason:     "Violation of Standard #01: Use of os.ReadFile detected. Use buffered readers (bufio) instead.",
			})
		}

		// Standard #05: Anti-Silent Errors
		if !isValidatorItself && strings.Contains(line, "return nil, err") && !strings.Contains(path, "legacy") {
			violations = append(violations, Violation{
				StandardID: "STD-05",
				FilePath:   path,
				Line:       lineNum,
				Reason:     "Violation of Standard #05: Raw error return detected. Use fmt.Errorf wrapping.",
			})
		}

		lineNum++
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("validator: scanner error at %s: %w", path, err)
	}

	return violations, nil
}

func isIgnored(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".go" {
		return true
	}
	ignored := []string{"vendor", "node_modules", ".git", "legacy"}
	for _, i := range ignored {
		if strings.Contains(path, i) {
			return true
		}
	}
	return false
}
