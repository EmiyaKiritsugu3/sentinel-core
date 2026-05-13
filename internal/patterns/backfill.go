// Package patterns manages architectural pattern extraction and storage.
package patterns

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/EmiyaKiritsugu3/sentinel-core/pkg/sqlite"
)

// BackfillResult holds the outcome of a backfill operation.
type BackfillResult struct {
	Extracted  int
	Inserted   int
	Skipped    int
	Errors     []string
	Candidates []BackfillCandidate
}

// BackfillCandidate represents a single pattern extracted from a source document.
type BackfillCandidate struct {
	Title       string
	Description string
	Category    string
	Source      string
	SourceRef   string
	Tags        string
	Impact      string
}

func (s *PatternStore) insertIfNew(ctx context.Context, c BackfillCandidate, source string, result *BackfillResult) {
	existing, _ := s.List(ctx, ListFilters{})
	for _, p := range existing {
		if strings.EqualFold(p.Title, c.Title) {
			result.Skipped++
			return
		}
	}

	_, err := s.Create(ctx, &Pattern{
		Title:       c.Title,
		Description: c.Description,
		Category:    c.Category,
		Source:      source,
		SourceRef:   c.SourceRef,
		Tags:        c.Tags,
		Impact:      c.Impact,
	})
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", c.Title, err))
		return
	}
	result.Inserted++
}

// BackfillFromCognitiveDNA extracts patterns from the COGNITIVE-DNA.md file and inserts them.
func (s *PatternStore) BackfillFromCognitiveDNA(ctx context.Context, baseDir string) (BackfillResult, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.BackfillFromCognitiveDNA"); err != nil {
		return BackfillResult{}, err
	}
	var result BackfillResult
	candidates, err := parseCognitiveDNA(filepath.Join(baseDir, "docs/process/COGNITIVE-DNA.md"))
	if err != nil {
		return result, fmt.Errorf("patterns: backfill cognitive-dna: %w", err)
	}
	result.Extracted = len(candidates)

	for _, c := range candidates {
		s.insertIfNew(ctx, c, SourceCognitiveDNA, &result)
	}
	return result, nil
}

// BackfillFromEvolutionInsights extracts patterns from the EVOLUTION-INSIGHTS.md file and inserts them.
func (s *PatternStore) BackfillFromEvolutionInsights(ctx context.Context, baseDir string) (BackfillResult, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.BackfillFromEvolutionInsights"); err != nil {
		return BackfillResult{}, err
	}
	var result BackfillResult
	candidates, err := parseEvolutionInsights(filepath.Join(baseDir, "docs/process/EVOLUTION-INSIGHTS.md"))
	if err != nil {
		return result, fmt.Errorf("patterns: backfill evolution-insights: %w", err)
	}
	result.Extracted = len(candidates)

	for _, c := range candidates {
		s.insertIfNew(ctx, c, SourceEvolutionInsights, &result)
	}
	return result, nil
}

// BackfillFromSentinelLog extracts patterns from the sentinel-log.md file and inserts them.
func (s *PatternStore) BackfillFromSentinelLog(ctx context.Context, baseDir string, dryRun bool) (BackfillResult, error) {
	if err := sqlite.ValidateDB(s.db, "pattern-store.BackfillFromSentinelLog"); err != nil {
		return BackfillResult{}, err
	}
	candidates, err := parseSentinelLog(filepath.Join(baseDir, "docs/process/sentinel-log.md"))
	if err != nil {
		return BackfillResult{}, fmt.Errorf("patterns: backfill sentinel-log: %w", err)
	}

	var result BackfillResult
	result.Extracted = len(candidates)
	result.Candidates = candidates

	if dryRun {
		return result, nil
	}
	for _, c := range candidates {
		s.insertIfNew(ctx, c, SourceSentinelLog, &result)
	}
	return result, nil
}

func parseAPTableRow(line string) (BackfillCandidate, bool) {
	if !strings.Contains(line, "[AP-") {
		return BackfillCandidate{}, false
	}
	parts := strings.Split(line, "|")
	if len(parts) < 5 {
		return BackfillCandidate{}, false
	}

	stripBold := func(s string) string {
		return strings.TrimPrefix(strings.TrimSuffix(strings.TrimSpace(s), "**"), "**")
	}
	id := stripBold(parts[1])
	name := stripBold(parts[2])
	desc := stripBold(parts[3])

	return BackfillCandidate{
		Title:       fmt.Sprintf("%s: %s", id, name),
		Description: desc,
		Category:    CategoryAntiPattern,
		SourceRef:   fmt.Sprintf("COGNITIVE-DNA.md:%s", id),
		Tags:        "anti-pattern,cognitive-dna",
		Impact:      ImpactHigh,
	}, true
}

func parsePMOBodyLine(line string, pmoBody *strings.Builder) {
	switch {
	case strings.Contains(line, "- **Regra:**"):
		pmoBody.WriteString(strings.TrimPrefix(line, "- **Regra:**"))
		pmoBody.WriteString(" ")
	case strings.Contains(line, "- **Modus Operandi:**"):
		pmoBody.WriteString(strings.TrimPrefix(line, "- **Modus Operandi:**"))
	}
}

func flushPMO(currentPMO BackfillCandidate, pmoBody *strings.Builder, candidates *[]BackfillCandidate) {
	if pmoBody.Len() > 0 {
		currentPMO.Description = strings.TrimSpace(pmoBody.String())
		*candidates = append(*candidates, currentPMO)
	}
}

func parseCognitiveDNA(path string) ([]BackfillCandidate, error) {
	f, err := os.Open(path) //nolint:gosec // path from caller
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var candidates []BackfillCandidate
	scanner := bufio.NewScanner(f)
	var currentPMO BackfillCandidate
	var inPMO bool
	var pmoBody strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if c, ok := parseAPTableRow(line); ok {
			candidates = append(candidates, c)
			continue
		}

		if strings.HasPrefix(line, "### PMO-") {
			if inPMO {
				flushPMO(currentPMO, &pmoBody, &candidates)
			}
			title := strings.TrimPrefix(line, "### ")
			currentPMO = BackfillCandidate{
				Title:     title,
				Category:  CategoryStructuralPrinciple,
				SourceRef: fmt.Sprintf("COGNITIVE-DNA.md:%s", title),
				Tags:      "modus-operandi,cognitive-dna",
				Impact:    ImpactMedium,
			}
			inPMO = true
			pmoBody.Reset()
			continue
		}

		if inPMO {
			parsePMOBodyLine(line, &pmoBody)
		}
	}

	if inPMO {
		flushPMO(currentPMO, &pmoBody, &candidates)
	}

	return candidates, scanner.Err()
}

func parseEvolutionLine(line string, inGaps, inCognitive bool) (BackfillCandidate, bool) {
	if !inGaps && !inCognitive {
		return BackfillCandidate{}, false
	}
	if !strings.HasPrefix(line, "- ") {
		return BackfillCandidate{}, false
	}
	if strings.Contains(line, "~~") {
		return BackfillCandidate{}, false
	}

	clean := strings.TrimPrefix(line, "- ")
	clean = strings.TrimPrefix(clean, "[x] ")
	clean = strings.TrimPrefix(clean, "[ ] ")
	clean = strings.TrimPrefix(clean, "**")
	clean = strings.TrimSuffix(clean, "**")

	parts := strings.SplitN(clean, ":", 2)
	title := strings.TrimSpace(parts[0])
	if title == "" {
		return BackfillCandidate{}, false
	}

	desc := title
	if len(parts) > 1 {
		desc = strings.TrimSpace(parts[1])
	}

	category := CategoryStructuralPrinciple
	if inCognitive {
		category = CategoryCognitivePattern
	}

	return BackfillCandidate{
		Title:       title,
		Description: desc,
		Category:    category,
		SourceRef:   "EVOLUTION-INSIGHTS.md",
		Tags:        "evolution-insights",
		Impact:      ImpactMedium,
	}, true
}

func parseEvolutionInsights(path string) ([]BackfillCandidate, error) {
	f, err := os.Open(path) //nolint:gosec // path from caller
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var candidates []BackfillCandidate
	scanner := bufio.NewScanner(f)
	inGaps := false
	inCognitive := false

	for scanner.Scan() {
		line := scanner.Text()

		switch {
		case strings.Contains(line, "Gaps Estruturais"):
			inGaps = true
			inCognitive = false
		case strings.Contains(line, "Cognitive Patterns"):
			inGaps = false
			inCognitive = true
		case strings.HasPrefix(line, "##"):
			inGaps = false
			inCognitive = false
		default:
			if c, ok := parseEvolutionLine(line, inGaps, inCognitive); ok {
				candidates = append(candidates, c)
			}
		}
	}

	return candidates, scanner.Err()
}

func detectFiltro(line string) string {
	switch {
	case strings.Contains(line, "Filtro A"):
		return "A"
	case strings.Contains(line, "Filtro B"):
		return "B"
	case strings.Contains(line, "Filtro C"):
		return "C"
	default:
		return "unknown"
	}
}

func parseSentinelLine(line string) (BackfillCandidate, bool) {
	hasFiltro := strings.Contains(line, "Filtro A") || strings.Contains(line, "Filtro B") || strings.Contains(line, "Filtro C")
	if !hasFiltro {
		return BackfillCandidate{}, false
	}

	clean := strings.TrimPrefix(line, "- ")
	clean = strings.TrimPrefix(clean, "* ")
	clean = strings.TrimPrefix(clean, "**")

	if len(clean) <= 10 {
		return BackfillCandidate{}, false
	}

	return BackfillCandidate{
		Title:       clean,
		Description: clean,
		Category:    CategoryRoutingPrinciple,
		SourceRef:   fmt.Sprintf("sentinel-log.md:Filtro-%s", detectFiltro(line)),
		Tags:        "epiphany,sentinel-log",
		Impact:      ImpactMedium,
	}, true
}

func parseSentinelLog(path string) ([]BackfillCandidate, error) {
	f, err := os.Open(path) //nolint:gosec // path from caller
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

	var candidates []BackfillCandidate
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if c, ok := parseSentinelLine(line); ok {
			candidates = append(candidates, c)
		}
	}

	return candidates, scanner.Err()
}
