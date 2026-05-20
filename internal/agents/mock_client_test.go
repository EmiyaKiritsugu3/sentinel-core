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
func (m *MockModel) StartChat() bridge.GenaiChatSession {
	if m.Session == nil {
		m.Session = &MockSession{}
	}
	return m.Session
}

// GenerateContent implements bridge.GenaiModel.
func (m *MockModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.Response, m.Err
}

// MockSession mocks bridge.GenaiChatSession for unit tests.
type MockSession struct {
	Responses []*genai.GenerateContentResponse
	Idx       int
	Err       error
}

// SendMessage implements bridge.GenaiChatSession.
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
