package utils

import "testing"

func TestSanitizeID(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"colons", "a:b", "a_b"},
		{"slashes", "a/b", "a_b"},
		{"dots", "a.b", "a_b"},
		{"hyphens", "a-b", "a_b"},
		{"spaces", "a b", "a_b"},
		{"mixed", "a:b/c.d-e f", "a_b_c_d_e_f"},
		{"normal", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SanitizeID(tt.input); got != tt.expected {
				t.Errorf("SanitizeID() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestSlugify(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"normal", "Hello World", "hello-world"},
		{"with special", "Hello, World!", "hello-world"},
		{"multiple spaces", "Hello   World", "hello-world"},
		{"underscores", "hello_world", "hello-world"},
		{"empty", "", "unnamed-decision"},
		{"only special", "!@#$", "unnamed-decision"},
		{"mixed", "Test 123_abc", "test-123-abc"},
		{"leading and trailing hyphens after replace", " -Hello World- ", "hello-world"},
		{"numbers", "123 456", "123-456"},
		{"mixed cases", "mIxEd CaSe", "mixed-case"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Slugify(tt.input); got != tt.expected {
				t.Errorf("Slugify() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestEscapeYAML(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{"newlines", "a\nb", "a b"},
		{"quotes", `a"b`, `a\"b`},
		{"backslashes", `a\b`, `a\\b`},
		{"mixed", "a\n\"b\\c\"", `a \"b\\c\"`},
		{"normal", "abc", "abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := EscapeYAML(tt.input); got != tt.expected {
				t.Errorf("EscapeYAML() = %v, want %v", got, tt.expected)
			}
		})
	}
}
