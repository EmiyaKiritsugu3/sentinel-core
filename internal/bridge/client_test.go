package bridge_test

import (
	"context"
	"errors"
	"testing"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

func TestNewSDKClient_NilClient(t *testing.T) {
	c, err := bridge.NewSDKClient(nil)
	if !errors.Is(err, bridge.ErrNilClient) {
		t.Fatalf("expected ErrNilClient, got: %v", err)
	}
	if c != nil {
		t.Fatalf("expected nil client wrapper, got: %v", c)
	}
}

func TestSDKClient_Methods(t *testing.T) {
	// Create a real genai.Client with a fake API key
	client, err := genai.NewClient(context.Background(), option.WithAPIKey("fake-key"))
	if err != nil {
		t.Fatalf("failed to create genai client: %v", err)
	}
	defer client.Close()

	sdkClt, err := bridge.NewSDKClient(client)
	if err != nil {
		t.Fatalf("failed to create SDK client wrapper: %v", err)
	}

	// Test GenerativeModel wrapper method
	model := sdkClt.GenerativeModel("gemini-1.5-flash")
	if model == nil {
		t.Fatal("expected non-nil model wrapper")
	}

	// Test model settings wrapper methods
	model.SetTemperature(0.7)
	model.SetSystemInstruction("test instruction")
	model.SetSystemInstructionContent(&genai.Content{
		Parts: []genai.Part{genai.Text("test")},
	})
	model.SetTools([]*genai.Tool{})

	// Test StartChat wrapper method
	session := model.StartChat()
	if session == nil {
		t.Fatal("expected non-nil session wrapper")
	}

	// Test GenerateContent and SendMessage (these will return network errors but cover the statements)
	_, _ = model.GenerateContent(context.Background(), genai.Text("test"))
	_, _ = session.SendMessage(context.Background(), genai.Text("test"))

	// Test Close wrapper method
	if err := sdkClt.Close(); err != nil {
		t.Errorf("expected nil error on Close, got: %v", err)
	}
}
