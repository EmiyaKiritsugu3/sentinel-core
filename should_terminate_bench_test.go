package main

import (
	"strings"
	"testing"
)

func shouldTerminateOriginal(toolCalls []map[string]interface{}, textResponses []string) bool {
	if len(toolCalls) == 0 {
		for _, text := range textResponses {
			if strings.Contains(strings.ToLower(text), "sovereign audit") {
				return true
			}
		}
	}
	return false
}

func shouldTerminateOptimized(toolCalls []map[string]interface{}, textResponses []string) bool {
	if len(toolCalls) == 0 {
		for _, text := range textResponses {
			if strings.Contains(text, "Sovereign Audit") || strings.Contains(text, "sovereign audit") {
				return true
			}
		}
	}
	return false
}

func BenchmarkShouldTerminate_Original(b *testing.B) {
	texts := []string{
		"Here is my detailed analysis of the codebase.",
		"I have looked at the files and everything looks okay.",
		"Nothing to see here.",
		"Final statement: Sovereign Audit Report: Done.",
	}
	for i := 0; i < b.N; i++ {
		_ = shouldTerminateOriginal(nil, texts)
	}
}

func BenchmarkShouldTerminate_Optimized(b *testing.B) {
	texts := []string{
		"Here is my detailed analysis of the codebase.",
		"I have looked at the files and everything looks okay.",
		"Nothing to see here.",
		"Final statement: Sovereign Audit Report: Done.",
	}
	for i := 0; i < b.N; i++ {
		_ = shouldTerminateOptimized(nil, texts)
	}
}

func BenchmarkShouldTerminateNoMatch_Original(b *testing.B) {
	texts := []string{
		"Here is my detailed analysis of the codebase.",
		"I have looked at the files and everything looks okay.",
		"Nothing to see here.",
		"Final statement: Everything looks ok.",
	}
	for i := 0; i < b.N; i++ {
		_ = shouldTerminateOriginal(nil, texts)
	}
}

func BenchmarkShouldTerminateNoMatch_Optimized(b *testing.B) {
	texts := []string{
		"Here is my detailed analysis of the codebase.",
		"I have looked at the files and everything looks okay.",
		"Nothing to see here.",
		"Final statement: Everything looks ok.",
	}
	for i := 0; i < b.N; i++ {
		_ = shouldTerminateOptimized(nil, texts)
	}
}
