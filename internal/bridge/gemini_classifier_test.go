package bridge_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/google/generative-ai-go/genai"
)

type mockClassifierClient struct {
	model *mockClassifierModel
}

func (m *mockClassifierClient) GenerativeModel(model string) bridge.GenaiModel {
	return m.model
}

func (m *mockClassifierClient) Close() error {
	return nil
}

type mockClassifierModel struct {
	instruction *genai.Content
	response    *genai.GenerateContentResponse
	err         error
}

func (m *mockClassifierModel) SetTemperature(temp float32) {}

func (m *mockClassifierModel) SetSystemInstruction(instruction string) {}

func (m *mockClassifierModel) SetSystemInstructionContent(content *genai.Content) {
	m.instruction = content
}

func (m *mockClassifierModel) SetTools(tools []*genai.Tool) {}

func (m *mockClassifierModel) StartChat() bridge.GenaiChatSession {
	return nil
}

func (m *mockClassifierModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.response, m.err
}

func TestNewGeminiClassifier_NilClient(t *testing.T) {
	c, err := bridge.NewGeminiClassifier(nil)
	if err == nil {
		t.Fatal("expected error for nil client, got nil")
	}
	if c != nil {
		t.Error("expected nil classifier")
	}
}

func TestGeminiClassifier_Classify_Success(t *testing.T) {
	mockM := &mockClassifierModel{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("  IMPLEMENT  \n"),
						},
					},
				},
			},
		},
	}
	mockC := &mockClassifierClient{model: mockM}

	g, err := bridge.NewGeminiClassifier(mockC)
	if err != nil {
		t.Fatalf("failed to create classifier: %v", err)
	}

	intent, err := g.Classify(context.Background(), "add auth validation")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentImplement {
		t.Errorf("expected IMPLEMENT, got %s", intent)
	}

	if mockM.instruction == nil || len(mockM.instruction.Parts) == 0 {
		t.Fatal("expected system instruction to be set")
	}
	sysText, ok := mockM.instruction.Parts[0].(genai.Text)
	if !ok || !strings.Contains(string(sysText), "task classifier") {
		t.Errorf("unexpected system instruction: %v", mockM.instruction.Parts[0])
	}
}

func TestGeminiClassifier_Classify_Unrecognized(t *testing.T) {
	mockM := &mockClassifierModel{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{
				{
					Content: &genai.Content{
						Parts: []genai.Part{
							genai.Text("banana"),
						},
					},
				},
			},
		},
	}
	mockC := &mockClassifierClient{model: mockM}

	g, err := bridge.NewGeminiClassifier(mockC)
	if err != nil {
		t.Fatalf("failed to create classifier: %v", err)
	}

	intent, err := g.Classify(context.Background(), "something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentUnknown {
		t.Errorf("expected UNKNOWN for unrecognized text, got %s", intent)
	}
}

func TestGeminiClassifier_Classify_Error(t *testing.T) {
	expectedErr := errors.New("gemini api down")
	mockM := &mockClassifierModel{
		err: expectedErr,
	}
	mockC := &mockClassifierClient{model: mockM}

	g, err := bridge.NewGeminiClassifier(mockC)
	if err != nil {
		t.Fatalf("failed to create classifier: %v", err)
	}

	_, err = g.Classify(context.Background(), "something")
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected wrapped error %v, got %v", expectedErr, err)
	}
}

func TestGeminiClassifier_Classify_Empty(t *testing.T) {
	// Scenario A: empty candidates
	mockM := &mockClassifierModel{
		response: &genai.GenerateContentResponse{
			Candidates: []*genai.Candidate{},
		},
	}
	mockC := &mockClassifierClient{model: mockM}
	g, _ := bridge.NewGeminiClassifier(mockC)
	intent, err := g.Classify(context.Background(), "something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentUnknown {
		t.Errorf("expected UNKNOWN, got %s", intent)
	}

	// Scenario B: nil content
	mockM.response.Candidates = []*genai.Candidate{{Content: nil}}
	intent, err = g.Classify(context.Background(), "something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentUnknown {
		t.Errorf("expected UNKNOWN, got %s", intent)
	}

	// Scenario C: empty parts
	mockM.response.Candidates = []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{}}}}
	intent, err = g.Classify(context.Background(), "something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentUnknown {
		t.Errorf("expected UNKNOWN, got %s", intent)
	}

	// Scenario D: non-text part type
	mockM.response.Candidates = []*genai.Candidate{{Content: &genai.Content{Parts: []genai.Part{genai.Blob{}}}}}
	intent, err = g.Classify(context.Background(), "something")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentUnknown {
		t.Errorf("expected UNKNOWN, got %s", intent)
	}
}

func TestGeminiClassifier_Classify_Nil(t *testing.T) {
	var g *bridge.GeminiClassifier
	intent, err := g.Classify(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if intent != bridge.IntentUnknown {
		t.Errorf("expected UNKNOWN, got %s", intent)
	}
}
