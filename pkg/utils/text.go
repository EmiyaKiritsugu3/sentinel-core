package utils

import (
	"strings"
)

// SanitizeID remove caracteres que quebram a sintaxe do Mermaid ou do SQLite
func SanitizeID(id string) string {
	replacer := strings.NewReplacer(
		":", "_",
		"/", "_",
		".", "_",
		"-", "_",
		" ", "_",
	)
	return replacer.Replace(id)
}

// Slugify transforms a string into a file-name-friendly format
func Slugify(text string) string {
	var b strings.Builder
	b.Grow(len(text))

	lastDash := true // true to prevent leading dash
	for i := 0; i < len(text); i++ {
		c := text[i]
		if (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') {
			b.WriteByte(c)
			lastDash = false
		} else if c >= 'A' && c <= 'Z' {
			b.WriteByte(c + ('a' - 'A'))
			lastDash = false
		} else if c == ' ' || c == '_' || c == '-' {
			if !lastDash {
				b.WriteByte('-')
				lastDash = true
			}
		}
	}

	res := b.String()
	res = strings.TrimRight(res, "-")

	if res == "" {
		return "unnamed-decision"
	}
	return res
}

// EscapeYAML prepares a string for safe use inside YAML double quotes
func EscapeYAML(text string) string {
	res := strings.ReplaceAll(text, "\\", "\\\\")
	res = strings.ReplaceAll(res, "\"", "\\\"")
	res = strings.ReplaceAll(res, "\n", " ")
	return res
}
