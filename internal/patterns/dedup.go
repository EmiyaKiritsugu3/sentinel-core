package patterns

import (
	"fmt"
	"strings"
)

func levenshteinDistance(a, b string) int {
	la, lb := len(a), len(b)
	if la == 0 {
		return lb
	}
	if lb == 0 {
		return la
	}

	prev := make([]int, lb+1)
	curr := make([]int, lb+1)

	for j := 0; j <= lb; j++ {
		prev[j] = j
	}

	for i := 1; i <= la; i++ {
		curr[0] = i
		for j := 1; j <= lb; j++ {
			cost := 1
			if a[i-1] == b[j-1] {
				cost = 0
			}
			curr[j] = min(
				prev[j]+1,
				curr[j-1]+1,
				prev[j-1]+cost,
			)
		}
		prev, curr = curr, prev
	}
	return prev[lb]
}

func tagOverlap(a, b string) float64 {
	tagsA := parseTags(a)
	tagsB := parseTags(b)
	if len(tagsA) == 0 || len(tagsB) == 0 {
		return 0.0
	}

	setB := make(map[string]bool, len(tagsB))
	for _, t := range tagsB {
		setB[strings.ToLower(t)] = true
	}

	matches := 0
	for _, t := range tagsA {
		if setB[strings.ToLower(t)] {
			matches++
		}
	}
	return float64(matches) / float64(len(tagsA))
}

func parseTags(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}

const (
	levenshteinThreshold  = 3
	tagOverlapThreshold   = 0.5
)

func (s *PatternStore) FindSimilar(title string, tags []string) ([]Pattern, error) {
	all, err := s.List(ListFilters{})
	if err != nil {
		return nil, fmt.Errorf("patterns: find similar: %w", err)
	}

	tagsStr := strings.Join(tags, ",")
	var similar []Pattern
	for _, p := range all {
		if levenshteinDistance(strings.ToLower(title), strings.ToLower(p.Title)) <= levenshteinThreshold {
			similar = append(similar, p)
			continue
		}
		if tagOverlap(tagsStr, p.Tags) >= tagOverlapThreshold {
			similar = append(similar, p)
		}
	}
	return similar, nil
}
