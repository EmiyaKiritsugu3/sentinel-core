// Package bridge connects the agent engine to external AI providers.
// internal/bridge/classifier.go
package bridge

import (
	"context"
	"log/slog"
	"strings"
	"sync"
)

// Intent represents a task intent category.
type Intent string

// Intent constants for task classification.
const (
	IntentDiagnose  Intent = "diagnose"
	IntentImplement Intent = "implement"
	IntentRefactor  Intent = "refactor"
	IntentReview    Intent = "review"
	IntentUnknown   Intent = "unknown"
)

var intentKeywords = map[Intent][]string{
	IntentDiagnose:  {"fix", "bug", "error", "broken", "failing", "crash", "debug", "investigate", "corrigir", "erro"},
	IntentImplement: {"add", "create", "build", "implement", "new", "adicionar", "criar", "implementar"},
	IntentRefactor:  {"refactor", "cleanup", "reorganize", "extract", "move", "simplify", "refatorar"},
	IntentReview:    {"review", "audit", "check", "verify", "analyze", "validate", "revisar", "auditar"},
}

// AIClassifier is the interface for AI-powered intent classification.
// The zero value (nil) means heuristic-only mode.
type AIClassifier interface {
	Classify(ctx context.Context, description string) (Intent, error)
}

// IntentClassifier classifies task intent using a tiered strategy:
// heuristic first, AI fallback when confidence is below threshold.
type IntentClassifier struct {
	ai        AIClassifier
	threshold float64
	cache     sync.Map // taskID → Intent, goroutine-safe
}

// NewIntentClassifier creates a new IntentClassifier with the given AI classifier and confidence threshold.
func NewIntentClassifier(ai AIClassifier, threshold float64) *IntentClassifier {
	return &IntentClassifier{ai: ai, threshold: threshold}
}

// Classify returns the Intent for a task. Results are cached by taskID.
func (c *IntentClassifier) Classify(ctx context.Context, taskID, description string) Intent {
	if v, ok := c.cache.Load(taskID); ok {
		return v.(Intent)
	}
	intent, confidence := heuristicClassify(description)
	if confidence < c.threshold {
		if c.ai != nil {
			if aiIntent, err := c.ai.Classify(ctx, description); err == nil {
				intent = aiIntent
			} else {
				slog.Warn("gemini fallback failed", "error", err)
			}
		} else {
			intent = IntentUnknown
		}
	}
	c.cache.Store(taskID, intent)
	return intent
}

func heuristicClassify(description string) (Intent, float64) {
	lower := strings.ToLower(description)
	words := strings.Fields(lower)

	hits := map[Intent]int{}
	for _, word := range words {
		for intent, keywords := range intentKeywords {
			for _, kw := range keywords {
				word = strings.Trim(word, ".,:;!?()[]{}\"'")
				if word == kw {
					hits[intent]++
				}
			}
		}
	}

	categoriesHit := 0
	var bestIntent Intent
	bestCount := 0
	for intent, count := range hits {
		if count > 0 {
			categoriesHit++
		}
		if count > bestCount {
			bestCount = count
			bestIntent = intent
		}
	}

	switch categoriesHit {
	case 0:
		return IntentUnknown, 0.00
	case 1:
		return bestIntent, 0.85
	default:
		return bestIntent, 0.30
	}
}

// NilClassifier is a null object for AIClassifier. Use in tests and
// when no AI key is configured.
type NilClassifier struct{}

// NewNilClassifier creates a new NilClassifier.
func NewNilClassifier() *NilClassifier { return &NilClassifier{} }

// Classify returns IntentUnknown with no error.
func (n *NilClassifier) Classify(_ context.Context, _ string) (Intent, error) {
	return IntentUnknown, nil
}
