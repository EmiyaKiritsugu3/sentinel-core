// Package bridge connects the agent engine to external AI providers.
package bridge

import (
	"context"
	"fmt"
	"log/slog"
	"strings"

	"github.com/google/generative-ai-go/genai"
)

// GeminiClassifier implements AIClassifier using the Gemini API.
type GeminiClassifier struct {
	client *genai.Client
}

// defaultModelID is the Gemini model used for intent classification.
const defaultModelID = "gemini-1.5-flash"

// NewGeminiClassifier creates a new GeminiClassifier with the given genai client.
func NewGeminiClassifier(client *genai.Client) (*GeminiClassifier, error) {
	if client == nil {
		return nil, fmt.Errorf("gemini-classifier: nil genai client")
	}
	return &GeminiClassifier{client: client}, nil
}

// Classify uses the Gemini API to classify the task description into an Intent.
func (g *GeminiClassifier) Classify(ctx context.Context, description string) (Intent, error) {
	if g == nil || g.client == nil {
		return IntentUnknown, nil
	}
	model := g.client.GenerativeModel(defaultModelID)
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
		slog.Warn("unrecognized AI intent", "intent", parsed)
		return IntentUnknown, nil
	}
}
