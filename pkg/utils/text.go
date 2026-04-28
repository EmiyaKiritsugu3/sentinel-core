package utils

import (
	"regexp"
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

// Slugify transforma uma string em um formato amigável para nomes de arquivos
func Slugify(text string) string {
	// 1. Lowercase
	res := strings.ToLower(text)
	
	// 2. Remove caracteres especiais (mantém apenas letras, números e espaços)
	reg := regexp.MustCompile(`[^a-z0-9\s-]+`)
	res = reg.ReplaceAllString(res, "")
	
	// 3. Substitui espaços e underscores por hífens
	res = strings.ReplaceAll(res, " ", "-")
	res = strings.ReplaceAll(res, "_", "-")
	
	// 4. Remove hífens duplicados
	regDouble := regexp.MustCompile(`-+`)
	res = regDouble.ReplaceAllString(res, "-")
	
	// 5. Trim hífens nas extremidades
	res = strings.Trim(res, "-")

	// Fallback para caso o slug resulte em vazio
	if res == "" {
		return "unnamed-decision"
	}
	return res
}

// EscapeYAML prepara uma string para ser usada com segurança dentro de aspas duplas no YAML
func EscapeYAML(text string) string {
	res := strings.ReplaceAll(text, "\\", "\\\\")
	res = strings.ReplaceAll(res, "\"", "\\\"")
	res = strings.ReplaceAll(res, "\n", " ")
	return res
}
