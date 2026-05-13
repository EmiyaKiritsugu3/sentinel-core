package utils

import (
	"crypto/sha256"
	"fmt"
	"io"
	"os"
)

// CalculateHash generates a SHA256 hash of file content for change detection
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
