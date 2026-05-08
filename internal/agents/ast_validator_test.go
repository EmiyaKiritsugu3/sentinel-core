package agents

import (
	"testing"
)

func TestValidateASTIsomorphism(t *testing.T) {
	t.Run("empty path returns error", func(t *testing.T) {
		err := validateASTIsomorphism("", "package main")
		if err == nil {
			t.Fatal("expected error for empty path, got nil")
		}
		if err.Error() != "Gate B: empty path" {
			t.Errorf("unexpected error message: %v", err)
		}
	})

	t.Run("valid Go file returns nil", func(t *testing.T) {
		err := validateASTIsomorphism("main.go", "package main\n")
		if err != nil {
			t.Fatalf("expected nil for valid Go, got: %v", err)
		}
	})

	t.Run("invalid Go file returns structural audit error", func(t *testing.T) {
		err := validateASTIsomorphism("broken.go", "this is not valid go code !!!")
		if err == nil {
			t.Fatal("expected error for invalid Go, got nil")
		}
		if err.Error() == "Gate B: empty path" {
			t.Errorf("error should not be empty path: %v", err)
		}
	})

	t.Run("unsupported extension bypasses validation", func(t *testing.T) {
		for _, ext := range []string{".py", ".rs", ".java", ".rb", ".cpp"} {
			err := validateASTIsomorphism("file"+ext, "any content")
			if err != nil {
				t.Errorf("expected nil for unsupported extension %s, got: %v", ext, err)
			}
		}
	})

	t.Run("valid TypeScript file returns nil", func(t *testing.T) {
		err := validateASTIsomorphism("app.ts", "const x: number = 1;\n")
		if err != nil {
			t.Fatalf("expected nil for valid TS, got: %v", err)
		}
	})

	t.Run("valid TSX file returns nil", func(t *testing.T) {
		err := validateASTIsomorphism("component.tsx", "const App = () => <div>hello</div>;\n")
		if err != nil {
			t.Fatalf("expected nil for valid TSX, got: %v", err)
		}
	})

	t.Run("invalid TypeScript file returns structural audit error", func(t *testing.T) {
		err := validateASTIsomorphism("bad.ts", "function ( { !!! }}}")
		if err == nil {
			t.Fatal("expected error for invalid TS, got nil")
		}
	})

	t.Run("invalid TSX file returns structural audit error", func(t *testing.T) {
		err := validateASTIsomorphism("bad.tsx", "<<<<>>>")
		if err == nil {
			t.Fatal("expected error for invalid TSX, got nil")
		}
	})

	t.Run("Go file with valid syntax including comments", func(t *testing.T) {
		content := `package main

// Documentation comment
func main() {
	// inline comment
	println("hello")
}
`
		err := validateASTIsomorphism("main.go", content)
		if err != nil {
			t.Fatalf("expected nil for valid Go with comments, got: %v", err)
		}
	})
}
