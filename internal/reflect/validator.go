package reflect

import (
	"bufio"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

type Violation struct {
	StandardID string
	FilePath   string
	Line       int
	Reason     string
}

type Validator struct {
	db *sqlite.DB
}

func NewValidator(db *sqlite.DB) *Validator {
	return &Validator{db: db}
}

// ValidateProject varre o projeto em busca de violações de Standards
func (v *Validator) ValidateProject(root string) ([]Violation, error) {
	// 1. Busca os standards ativos (SEALED ou AUDITED)
	// (Simplificado para o MVP: bloqueamos padrões críticos via código)
	
	var violations []Violation

	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || isIgnored(path) {
			return nil
		}

		fileViolations, _ := v.checkFile(path)
		violations = append(violations, fileViolations...)
		return nil
	})

	return violations, err
}

func (v *Validator) checkFile(path string) ([]Violation, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var violations []Violation
	scanner := bufio.NewScanner(file)
	lineNum := 1

	for scanner.Scan() {
		line := scanner.Text()

		// Standard #01: Anti-os.ReadFile (Obrigatório usar bufio)
		if strings.Contains(line, "os.ReadFile") && !strings.Contains(path, "legacy") {
			violations = append(violations, Violation{
				StandardID: "STD-01",
				FilePath:   path,
				Line:       lineNum,
				Reason:     "Violation of Standard #01: Use of os.ReadFile detected. Use buffered readers (bufio) instead.",
			})
		}
		
		// Standard #03: Anti-Silent Errors (Obrigatório wrapping)
		if strings.Contains(line, "return nil, err") && !strings.Contains(path, "legacy") {
			violations = append(violations, Violation{
				StandardID: "STD-05",
				FilePath:   path,
				Line:       lineNum,
				Reason:     "Violation of Standard #05: Raw error return detected. Use fmt.Errorf wrapping.",
			})
		}

		lineNum++
	}

	return violations, nil
}

func isIgnored(path string) bool {
	ext := filepath.Ext(path)
	if ext != ".go" {
		return true
	}
	ignored := []string{"vendor", "node_modules", ".git", "legacy", "pkg/utils"} // pkg/utils é onde os padrões são definidos
	for _, i := range ignored {
		if strings.Contains(path, i) {
			return true
		}
	}
	return false
}
