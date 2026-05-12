// internal/bridge/gemini_classifier.go
package bridge

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// GeminiClassifier implements AIClassifier using the Gemini API.
type GeminiClassifier struct {
	client *genai.Client
}

func NewGeminiClassifier(client *genai.Client) (*GeminiClassifier, error) {
	if client == nil {
		return nil, fmt.Errorf("gemini-classifier: nil genai client")
	}
	return &GeminiClassifier{client: client}, nil
}

func (g *GeminiClassifier) Classify(ctx context.Context, description string) (Intent, error) {
	if g == nil || g.client == nil {
		return IntentUnknown, nil
	}
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
	if len(resp.Candidates) == 0 {
		return IntentUnknown, nil
	}

	cand := resp.Candidates[0]
	if cand.Content == nil || len(cand.Content.Parts) == 0 {
		return IntentUnknown, nil
	}
	part, ok := cand.Content.Parts[0].(genai.Text)
	if !ok {
		return IntentUnknown, nil
	}
	raw := string(part)
	parsed := Intent(strings.ToLower(strings.TrimSpace(raw)))
	switch parsed {
	case IntentDiagnose, IntentImplement, IntentRefactor, IntentReview:
		return parsed, nil
	default:
		fmt.Fprintf(os.Stderr, "warning: classifier: unrecognized ai intent: %q\n", parsed)
		return IntentUnknown, nil
	}
}
