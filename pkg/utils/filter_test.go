package utils

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIgnoreFilter(t *testing.T) {
	// Create a temporary directory for testing
	tempDir, err := os.MkdirTemp("", "filter_test")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create a .gitignore file
	gitignoreContent := `
*.log
node_modules/
/build
.env
# this is a comment
`
	err = os.WriteFile(filepath.Join(tempDir, ".gitignore"), []byte(gitignoreContent), 0644)
	if err != nil {
		t.Fatalf("Failed to write .gitignore: %v", err)
	}

	filter := NewIgnoreFilter(tempDir)

	tests := []struct {
		path     string
		expected bool
	}{
		// Hardcoded internal ignores
		{".git/config", true},
		{".sentinel/config.json", true},
		{"vendor/github.com/foo/bar", true},
		{"dist/output.js", true},
		{"build/output.js", true},
		{"src/build/something.txt", true},

		// .gitignore patterns
		{"app.log", true},
		{"src/app.log", true},
		{"node_modules/express/index.js", true},
		{"build", true},
		{".env", true},

		// Hidden files
		{".hidden_file", true},
		{"src/.hidden_file", true},

		// Should not be ignored
		{"src/main.go", false},
		{"README.md", false},
		{"build_script.sh", false},
		{"package.json", false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if got := filter.IsIgnored(test.path); got != test.expected {
				t.Errorf("IsIgnored(%q) = %v, want %v", test.path, got, test.expected)
			}
		})
	}
}

func TestIgnoreFilter_Fallback(t *testing.T) {
	filter := &IgnoreFilter{
		patterns: []string{"*.log", "temp/"},
	}

	tests := []struct {
		path     string
		expected bool
	}{
		{"test.log", true},
		{"temp/file.txt", true},
		{"src/main.go", false},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			if got := filter.IsIgnored(test.path); got != test.expected {
				t.Errorf("IsIgnored(%q) = %v, want %v", test.path, got, test.expected)
			}
		})
	}
}

func TestNewIgnoreFilter_NoGitignore(t *testing.T) {
	tempDir, err := os.MkdirTemp("", "filter_test_empty")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	filter := NewIgnoreFilter(tempDir)
	if filter.IsIgnored("main.go") {
		t.Error("Expected main.go to not be ignored")
	}
	if !filter.IsIgnored(".git/config") {
		t.Error("Expected .git/config to be ignored")
	}
}
