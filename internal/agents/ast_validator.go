package agents

import (
	"context"
	"fmt"
	"go/parser"
	"go/token"
	"path/filepath"

	sitter "github.com/smacker/go-tree-sitter"
	"github.com/smacker/go-tree-sitter/typescript/tsx"
	"github.com/smacker/go-tree-sitter/typescript/typescript"
)

func validateASTIsomorphism(path string, content string) error {
	if path == "" {
		return fmt.Errorf("Gate B: empty path")
	}

	ext := filepath.Ext(path)

	// Go validation using standard library (as in ScannerGo)
	if ext == ".go" {
		fset := token.NewFileSet()
		if _, err := parser.ParseFile(fset, path, content, parser.ParseComments); err != nil {
			return fmt.Errorf("Gate B: Structural Audit Failed for Go file. Syntax error: %v. Fix the syntax before writing.", err)
		}
		return nil
	}

	// Tree-sitter validation for TS/TSX
	var lang *sitter.Language
	switch ext {
	case ".ts":
		lang = typescript.GetLanguage()
	case ".tsx":
		lang = tsx.GetLanguage()
	default:
		// Unsupported language, bypass validation
		return nil
	}
	if lang == nil {
		return fmt.Errorf("Gate B: no Tree-sitter language available for %s", ext)
	}

	parserTs := sitter.NewParser()
	if parserTs == nil {
		return fmt.Errorf("Gate B: failed to initialize Tree-sitter parser")
	}
	parserTs.SetLanguage(lang)
	tree, err := parserTs.ParseCtx(context.Background(), nil, []byte(content))
	if err != nil {
		return fmt.Errorf("Gate B: parsing failed: %v", err)
	}
	if tree == nil {
		return fmt.Errorf("Gate B: parsing failed: Tree-sitter returned nil tree")
	}
	defer tree.Close()

	root := tree.RootNode()
	if root == nil {
		return fmt.Errorf("Gate B: parsing failed: Tree-sitter returned nil root node")
	}
	if root.HasError() {
		return fmt.Errorf("Gate B: Structural Audit Failed for TS/TSX. Generated code has invalid syntax (ERROR/MISSING node detected). Fix the syntax before writing.")
	}

	return nil
}
