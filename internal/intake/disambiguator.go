// internal/intake/disambiguator.go
package intake

import (
	"fmt"
	"math"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

const (
	weightLength   = 0.25
	weightVerb     = 0.20
	weightPronoun  = 0.15
	weightAnchor   = 0.40
	scoreThreshold = 0.50
)

var genericVerbs = []string{
	"fix", "improve", "update", "change", "make", "handle", "check",
	"corrigir", "melhorar", "atualizar", "mudar",
}

var vaguePronouns = []string{
	"it", "this", "the issue", "the bug", "the problem",
	"isso", "ele", "o problema", "o erro",
}

// Suggestion is a graph-anchored alternative for a vague task description.
type Suggestion struct {
	NodeName string
	FilePath string
}

// Disambiguator analyzes task descriptions for vagueness and suggests
// graph-anchored alternatives.
type Disambiguator struct {
	db *sqlite.DB // nil = skip graph phase (Phase 2)
}

func NewDisambiguator(db *sqlite.DB) *Disambiguator {
	return &Disambiguator{db: db}
}

// Analyze returns whether the description is vague and any graph suggestions.
func (d *Disambiguator) Analyze(description string) (vague bool, suggestions []Suggestion) {
	score := d.VaguenessScore(description)
	if score <= scoreThreshold {
		return false, nil
	}
	if d.db != nil {
		suggestions = d.queryGraph(description)
	}
	return true, suggestions
}

// VaguenessScore returns a score in [0.0, 1.0]. Values > 0.50 trigger suggestion.
func (d *Disambiguator) VaguenessScore(description string) float64 {
	score := lengthSignal(description) +
		verbSignal(description) +
		pronounSignal(description) +
		d.anchorSignal(description)
	return math.Min(score, 1.0)
}

func lengthSignal(description string) float64 {
	n := len(strings.Fields(description))
	switch {
	case n < 3:
		return weightLength // 0.25
	case n <= 5:
		return 0.18
	case n <= 10:
		return 0.08
	default:
		return 0.00
	}
}

func verbSignal(description string) float64 {
	lower := strings.ToLower(description)
	for _, v := range genericVerbs {
		if strings.Contains(lower, v) {
			return weightVerb // 0.20
		}
	}
	return 0.00
}

func pronounSignal(description string) float64 {
	lower := strings.ToLower(description)
	for _, p := range vaguePronouns {
		if strings.Contains(lower, p) {
			return weightPronoun // 0.15
		}
	}
	return 0.00
}

func (d *Disambiguator) anchorSignal(description string) float64 {
	lower := strings.ToLower(description)

	// Phase 1: lexical anchors (zero DB)
	if strings.Contains(lower, "internal/") ||
		strings.Contains(lower, "pkg/") ||
		strings.Contains(lower, ".go") {
		return 0.00
	}
	// line reference: colon followed by digit
	for i, ch := range description {
		if ch == ':' && i+1 < len(description) && description[i+1] >= '0' && description[i+1] <= '9' {
			return 0.00
		}
	}

	// Phase 2: graph-anchored (DB query)
	if d.db == nil {
		return weightAnchor // 0.40 — no graph available
	}

	var count int
	if err := d.db.Conn.QueryRow("SELECT COUNT(*) FROM nodes").Scan(&count); err != nil || count == 0 {
		return weightAnchor // graph not indexed
	}

	keywords := extractKeywords(description)
	if len(keywords) == 0 {
		return weightAnchor
	}

	matched := 0
	for _, kw := range keywords {
		var n int
		_ = d.db.Conn.QueryRow(
			"SELECT COUNT(*) FROM nodes WHERE LOWER(name) LIKE ?",
			fmt.Sprintf("%%%s%%", strings.ToLower(kw)),
		).Scan(&n)
		if n > 0 {
			matched++
		}
	}

	matchedRatio := float64(matched) / float64(len(keywords))
	return weightAnchor * (1.0 - matchedRatio)
}

func (d *Disambiguator) queryGraph(description string) []Suggestion {
	keywords := extractKeywords(description)
	var suggestions []Suggestion
	seen := map[string]bool{}

	for _, kw := range keywords {
		rows, err := d.db.Conn.Query(
			"SELECT name, file_path FROM nodes WHERE LOWER(name) LIKE ? LIMIT 3",
			fmt.Sprintf("%%%s%%", strings.ToLower(kw)),
		)
		if err != nil {
			continue
		}
		for rows.Next() {
			var s Suggestion
			if err := rows.Scan(&s.NodeName, &s.FilePath); err == nil && !seen[s.NodeName] {
				suggestions = append(suggestions, s)
				seen[s.NodeName] = true
			}
		}
		rows.Close()
		if len(suggestions) >= 5 {
			break
		}
	}
	return suggestions
}

func extractKeywords(description string) []string {
	lower := strings.ToLower(description)
	// Remove common stop words and return remaining tokens >= 3 chars
	stopWords := map[string]bool{
		"the": true, "a": true, "an": true, "in": true, "of": true,
		"to": true, "and": true, "or": true, "for": true, "with": true,
		"fix": true, "add": true, "new": true, "o": true,
	}
	var keywords []string
	for _, w := range strings.Fields(lower) {
		w = strings.Trim(w, ".,!?")
		if len(w) >= 3 && !stopWords[w] {
			keywords = append(keywords, w)
		}
	}
	return keywords
}
