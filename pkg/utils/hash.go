package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// CalculateHash gera um hash SHA256 do conteúdo de um arquivo para detecção de mudanças
func CalculateHash(path string) (string, error) {
	f, err := os.Open(path) //nolint:gosec // path from caller
	if err != nil {
		return "", fmt.Errorf("could not open file for hashing: %w", err)
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return "", fmt.Errorf("could not copy file content to hash: %w", err)
	}

	return fmt.Sprintf("%x", h.Sum(nil)), nil
}
