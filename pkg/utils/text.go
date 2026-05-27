package utils

import (
	"regexp"
	"strings"
)

var (
	sanitizeIDReplacer = strings.NewReplacer(
		":", "_",
		"/", "_",
		".", "_",
		"-", "_",
		" ", "_",
	)

	slugifySpecialCharsRegex = regexp.MustCompile(`[^a-z0-9\s-]+`)
	slugifyDoubleHyphenRegex = regexp.MustCompile(`-+`)
)

// SanitizeID remove caracteres que quebram a sintaxe do Mermaid ou do SQLite
func SanitizeID(id string) string {
	return sanitizeIDReplacer.Replace(id)
}

// Slugify transforms a string into a file-name-friendly format
func Slugify(text string) string {
	// 1. Lowercase
	res := strings.ToLower(text)

	// 2. Remove special characters (keep only letters, numbers and spaces)
	res = slugifySpecialCharsRegex.ReplaceAllString(res, "")

	// 3. Replace spaces and underscores with hyphens
	res = strings.ReplaceAll(res, " ", "-")
	res = strings.ReplaceAll(res, "_", "-")

	// 4. Remove duplicate hyphens
	res = slugifyDoubleHyphenRegex.ReplaceAllString(res, "-")

	// 5. Trim hyphens at edges
	res = strings.Trim(res, "-")

	// Fallback in case the slug results in empty string
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
