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
