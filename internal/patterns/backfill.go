package patterns

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type BackfillResult struct {
	Extracted int
	Inserted  int
	Skipped   int
	Errors    []string
}

type BackfillCandidate struct {
	Title       string
	Description string
	Category    string
	Source      string
	SourceRef   string
	Tags        string
	Impact      string
}

func (s *PatternStore) BackfillFromCognitiveDNA(baseDir string) (BackfillResult, error) {
	var result BackfillResult
	candidates, err := parseCognitiveDNA(filepath.Join(baseDir, "docs/process/COGNITIVE-DNA.md"))
	if err != nil {
		return result, fmt.Errorf("patterns: backfill cognitive-dna: %w", err)
	}
	result.Extracted = len(candidates)

	for _, c := range candidates {
		existing, _ := s.List(ListFilters{})
		dup := false
		for _, p := range existing {
			if strings.EqualFold(p.Title, c.Title) {
				dup = true
				break
			}
		}
		if dup {
			result.Skipped++
			continue
		}

		_, err := s.Create(&Pattern{
			Title:       c.Title,
			Description: c.Description,
			Category:    c.Category,
			Source:      SourceCognitiveDNA,
			SourceRef:   c.SourceRef,
			Tags:        c.Tags,
			Impact:      c.Impact,
		})
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", c.Title, err))
			continue
		}
		result.Inserted++
	}
	return result, nil
}

func (s *PatternStore) BackfillFromEvolutionInsights(baseDir string) (BackfillResult, error) {
	var result BackfillResult
	candidates, err := parseEvolutionInsights(filepath.Join(baseDir, "docs/process/EVOLUTION-INSIGHTS.md"))
	if err != nil {
		return result, fmt.Errorf("patterns: backfill evolution-insights: %w", err)
	}
	result.Extracted = len(candidates)

	for _, c := range candidates {
		existing, _ := s.List(ListFilters{})
		dup := false
		for _, p := range existing {
			if strings.EqualFold(p.Title, c.Title) {
				dup = true
				break
			}
		}
		if dup {
			result.Skipped++
			continue
		}

		_, err := s.Create(&Pattern{
			Title:       c.Title,
			Description: c.Description,
			Category:    c.Category,
			Source:      SourceEvolutionInsights,
			SourceRef:   c.SourceRef,
			Tags:        c.Tags,
			Impact:      c.Impact,
		})
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("%s: %v", c.Title, err))
			continue
		}
		result.Inserted++
	}
	return result, nil
}

func (s *PatternStore) BackfillFromSentinelLog(baseDir string, dryRun bool) ([]BackfillCandidate, error) {
	candidates, err := parseSentinelLog(filepath.Join(baseDir, "docs/process/sentinel-log.md"))
	if err != nil {
		return nil, fmt.Errorf("patterns: backfill sentinel-log: %w", err)
	}
	if dryRun {
		return candidates, nil
	}
	for _, c := range candidates {
		s.Create(&Pattern{
			Title:       c.Title,
			Description: c.Description,
			Category:    c.Category,
			Source:      SourceSentinelLog,
			SourceRef:   c.SourceRef,
			Tags:        c.Tags,
			Impact:      c.Impact,
		})
	}
	return candidates, nil
}

func parseCognitiveDNA(path string) ([]BackfillCandidate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var candidates []BackfillCandidate
	scanner := bufio.NewScanner(f)
	var currentPMO BackfillCandidate
	var inPMO bool
	var pmoBody strings.Builder

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "[AP-") {
			parts := strings.Split(line, "|")
			if len(parts) >= 5 {
				idPart := strings.TrimSpace(parts[1])
				namePart := strings.TrimSpace(parts[2])
				descPart := strings.TrimSpace(parts[3])
				id := strings.TrimPrefix(strings.TrimSuffix(idPart, "**"), "**")
				name := strings.TrimPrefix(strings.TrimSuffix(namePart, "**"), "**")
				desc := strings.TrimPrefix(strings.TrimSuffix(descPart, "**"), "**")

				candidates = append(candidates, BackfillCandidate{
					Title:       fmt.Sprintf("%s: %s", id, name),
					Description: desc,
					Category:    CategoryAntiPattern,
					SourceRef:   fmt.Sprintf("COGNITIVE-DNA.md:%s", id),
					Tags:        "anti-pattern,cognitive-dna",
					Impact:      ImpactHigh,
				})
			}
		}

		if strings.HasPrefix(line, "### PMO-") {
			if inPMO && pmoBody.Len() > 0 {
				currentPMO.Description = strings.TrimSpace(pmoBody.String())
				candidates = append(candidates, currentPMO)
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
			if strings.Contains(line, "- **Regra:**") {
				pmoBody.WriteString(strings.TrimPrefix(line, "- **Regra:**"))
				pmoBody.WriteString(" ")
			} else if strings.Contains(line, "- **Modus Operandi:**") {
				pmoBody.WriteString(strings.TrimPrefix(line, "- **Modus Operandi:**"))
			}
		}
	}

	if inPMO && pmoBody.Len() > 0 {
		currentPMO.Description = strings.TrimSpace(pmoBody.String())
		candidates = append(candidates, currentPMO)
	}

	return candidates, scanner.Err()
}

func parseEvolutionInsights(path string) ([]BackfillCandidate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var candidates []BackfillCandidate
	scanner := bufio.NewScanner(f)
	inGaps := false
	inCognitive := false

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "Gaps Estruturais") {
			inGaps = true
			inCognitive = false
			continue
		}
		if strings.Contains(line, "Cognitive Patterns") {
			inGaps = false
			inCognitive = true
			continue
		}
		if strings.HasPrefix(line, "##") {
			inGaps = false
			inCognitive = false
			continue
		}

		if (inGaps || inCognitive) && strings.HasPrefix(line, "- ") {
			clean := strings.TrimPrefix(line, "- ")
			clean = strings.TrimPrefix(clean, "[x] ")
			clean = strings.TrimPrefix(clean, "[ ] ")
			clean = strings.TrimPrefix(clean, "**")
			clean = strings.TrimSuffix(clean, "**")

			parts := strings.SplitN(clean, ":", 2)
			title := strings.TrimSpace(parts[0])
			desc := title
			if len(parts) > 1 {
				desc = strings.TrimSpace(parts[1])
			}

			if title == "" {
				continue
			}

			if strings.Contains(line, "~~") {
				continue
			}

			category := CategoryStructuralPrinciple
			if inCognitive {
				category = CategoryCognitivePattern
			}

			candidates = append(candidates, BackfillCandidate{
				Title:       title,
				Description: desc,
				Category:    category,
				SourceRef:   "EVOLUTION-INSIGHTS.md",
				Tags:        "evolution-insights",
				Impact:      ImpactMedium,
			})
		}
	}

	return candidates, scanner.Err()
}

func parseSentinelLog(path string) ([]BackfillCandidate, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var candidates []BackfillCandidate
	scanner := bufio.NewScanner(f)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, "Filtro A") || strings.Contains(line, "Filtro B") || strings.Contains(line, "Filtro C") {
			clean := strings.TrimPrefix(line, "- ")
			clean = strings.TrimPrefix(clean, "* ")
			clean = strings.TrimPrefix(clean, "**")

			if len(clean) > 10 {
				filtro := "unknown"
				if strings.Contains(line, "Filtro A") {
					filtro = "A"
				} else if strings.Contains(line, "Filtro B") {
					filtro = "B"
				} else if strings.Contains(line, "Filtro C") {
					filtro = "C"
				}

				candidates = append(candidates, BackfillCandidate{
					Title:       clean,
					Description: clean,
					Category:    CategoryRoutingPrinciple,
					SourceRef:   fmt.Sprintf("sentinel-log.md:Filtro-%s", filtro),
					Tags:        "epiphany,sentinel-log",
					Impact:      ImpactMedium,
				})
			}
		}
	}

	return candidates, scanner.Err()
}
