package utils

import (
	"testing"
)

func BenchmarkSlugify(b *testing.B) {
	text := "Hello World! This is a Test!!! 123_456"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Slugify(text)
	}
}

func BenchmarkSanitizeID(b *testing.B) {
	text := "Hello World! This is a Test!!! 123_456"
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		SanitizeID(text)
	}
}
