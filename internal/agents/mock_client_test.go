package agents

import (
	"context"

	"github.com/EmiyaKiritsugu3/sentinel-core/internal/bridge"
	"github.com/google/generative-ai-go/genai"
)

// MockClient mocks bridge.GenaiClient for unit tests.
type MockClient struct {
	Model *MockModel
}

// GenerativeModel implements bridge.GenaiClient.
func (m *MockClient) GenerativeModel(model string) bridge.GenaiModel {
	if m.Model == nil {
		m.Model = &MockModel{}
	}
	return m.Model
}

// Close implements bridge.GenaiClient.
func (m *MockClient) Close() error {
	return nil
}

// MockModel mocks bridge.GenaiModel for unit tests.
type MockModel struct {
	Temp        float32
	Instruction string
	Tools       []*genai.Tool
	Session     *MockSession
	// ChatSession overrides Session when non-nil, allowing any MessageSender
	// implementation to be injected (e.g. cancelOnSecondCallSession).
	ChatSession bridge.MessageSender
	Response    *genai.GenerateContentResponse
	Err         error
}

// SetTemperature implements bridge.GenaiModel.
func (m *MockModel) SetTemperature(temp float32) {
	m.Temp = temp
}

// SetSystemInstruction implements bridge.GenaiModel.
func (m *MockModel) SetSystemInstruction(inst string) {
	m.Instruction = inst
}

// SetSystemInstructionContent implements bridge.GenaiModel.
func (m *MockModel) SetSystemInstructionContent(content *genai.Content) {
	if content != nil && len(content.Parts) > 0 {
		if txt, ok := content.Parts[0].(genai.Text); ok {
			m.Instruction = string(txt)
		}
	}
}

// SetTools implements bridge.GenaiModel.
func (m *MockModel) SetTools(tools []*genai.Tool) {
	m.Tools = tools
}

// StartChat implements bridge.GenaiModel.
func (m *MockModel) StartChat() bridge.MessageSender {
	if m.ChatSession != nil {
		return m.ChatSession
	}
	if m.Session == nil {
		m.Session = &MockSession{}
	}
	return m.Session
}

// GenerateContent implements bridge.GenaiModel.
func (m *MockModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.Response, m.Err
}

// MockSession mocks bridge.MessageSender for unit tests.
type MockSession struct {
	Responses []*genai.GenerateContentResponse
	Idx       int
	Err       error
}

// SendMessage implements bridge.MessageSender.
func (s *MockSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	if s.Err != nil {
		return nil, s.Err
	}
	if s.Idx >= len(s.Responses) {
		return &genai.GenerateContentResponse{}, nil
	}
	r := s.Responses[s.Idx]
	s.Idx++
	return r, nil
}

// closeTrackingClient is a GenaiClient that records Close() calls.
type closeTrackingClient struct {
	onClose func()
}

// GenerativeModel implements bridge.GenaiClient.
func (c *closeTrackingClient) GenerativeModel(model string) bridge.GenaiModel {
	return &MockModel{}
}

// Close implements bridge.GenaiClient and invokes the onClose callback.
func (c *closeTrackingClient) Close() error {
	if c.onClose != nil {
		c.onClose()
	}
	return nil
}

// cancelOnSecondCallSession cancels a context on the second SendMessage call.
// This lets tests deterministically exercise the select-case <-ctx.Done() branch
// in the Execute loop, which only fires on the SECOND iteration.
type cancelOnSecondCallSession struct {
	Response *genai.GenerateContentResponse
	cancel   context.CancelFunc
	calls    int
}

// SendMessage implements bridge.MessageSender.
func (s *cancelOnSecondCallSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	s.calls++
	if s.calls == 2 {
		s.cancel()
	}
	if s.Response != nil {
		return s.Response, nil
	}
	return &genai.GenerateContentResponse{}, nil
}
