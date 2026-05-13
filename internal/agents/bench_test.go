package agents

import (
	"strings"
	"testing"
)

// --- isExplicitThoughtBlock benchmarks ---

func BenchmarkIsExplicitThoughtBlock_CoT(b *testing.B) {
	text := " geschichtenこれは内部推論です"
	b.ResetTimer()
	for b.Loop() {
		isExplicitThoughtBlock(text)
	}
}

func BenchmarkIsExplicitThoughtBlock_CodeBlock(b *testing.B) {
	text := "```thought\nLet me analyze this step by step\n```"
	b.ResetTimer()
	for b.Loop() {
		isExplicitThoughtBlock(text)
	}
}

func BenchmarkIsExplicitThoughtBlock_Normal(b *testing.B) {
	text := "This is a normal response without any thought markers"
	b.ResetTimer()
	for b.Loop() {
		isExplicitThoughtBlock(text)
	}
}

func BenchmarkIsExplicitThoughtBlock_Whitespace(b *testing.B) {
	text := "   geschichten prefixed with spaces   "
	b.ResetTimer()
	for b.Loop() {
		isExplicitThoughtBlock(text)
	}
}

// --- pacWorstCase benchmarks ---

func BenchmarkPacWorstCase_AllProceed(b *testing.B) {
	for b.Loop() {
		pacWorstCase(PACProceed, PACProceed, PACProceed)
	}
}

func BenchmarkPacWorstCase_Mixed(b *testing.B) {
	for b.Loop() {
		pacWorstCase(PACProceed, PACSimplify, PACPivot)
	}
}

func BenchmarkPacWorstCase_Escalate(b *testing.B) {
	for b.Loop() {
		pacWorstCase(PACSimplify, PACPivot, PACEscalate)
	}
}

// --- PACRecommendation.String benchmarks ---

func BenchmarkPACRecommendation_String(b *testing.B) {
	for b.Loop() {
		_ = PACEscalate.String()
		_ = PACProceed.String()
		_ = PACSimplify.String()
		_ = PACPivot.String()
	}
}

// --- parseResponseParts benchmark (synthetic) ---

func BenchmarkStringsContains_LongText(b *testing.B) {
	longText := strings.Repeat("The agent should proceed with the implementation carefully. ", 100)
	b.ResetTimer()
	for b.Loop() {
		strings.Contains(longText, "implementation")
	}
}

// --- countThoughtActionTokens benchmark (using string scanning) ---

func BenchmarkStringsHasPrefix_ManyPrefixes(b *testing.B) {
	texts := []string{
		" geschichtenThis is a thought block",
		"```thought\nstep 1\n```",
		"Normal response text",
		" _INDENT_this is code",
		"Another normal response",
	}
	b.ResetTimer()
	for b.Loop() {
		for _, t := range texts {
			trimmed := strings.TrimSpace(t)
			_ = strings.HasPrefix(trimmed, " geschichten") || strings.HasPrefix(trimmed, "```thought")
		}
	}
}
