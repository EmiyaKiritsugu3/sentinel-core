// internal/bridge/gemini_classifier.go
package bridge

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// GeminiClassifier implements AIClassifier using the Gemini API.
type GeminiClassifier struct {
	client *genai.Client
}

func NewGeminiClassifier(client *genai.Client) *GeminiClassifier {
	return &GeminiClassifier{client: client}
}

func (g *GeminiClassifier) Classify(ctx context.Context, description string) (Intent, error) {
	model := g.client.GenerativeModel("gemini-1.5-flash")
	model.SystemInstruction = &genai.Content{
		Parts: []genai.Part{genai.Text("You are a task classifier. Respond with exactly one word.")},
	}
	prompt := fmt.Sprintf(
		"Classify this software task into exactly one word: diagnose, implement, refactor, or review.\nTask: %s",
		description,
	)
	resp, err := model.GenerateContent(ctx, genai.Text(prompt))
	if err != nil {
		return IntentUnknown, fmt.Errorf("gemini classifier: %w", err)
	}
	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return IntentUnknown, nil
	}
	raw := fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0])
	parsed := Intent(strings.ToLower(strings.TrimSpace(raw)))
	switch parsed {
	case IntentDiagnose, IntentImplement, IntentRefactor, IntentReview:
		return parsed, nil
	default:
		return IntentUnknown, nil
	}
}
