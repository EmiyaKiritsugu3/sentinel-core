// Package bridge connects the agent engine to external AI providers.
package bridge

import (
	"context"
	"errors"

	"github.com/google/generative-ai-go/genai"
)

// ErrNilClient is returned when a nil generative AI client is provided.
var ErrNilClient = errors.New("bridge: nil generative AI client")

// GenaiClient abstracts *genai.Client to enable mocking and structural separation.
type GenaiClient interface {
	GenerativeModel(model string) GenaiModel
	Close() error
}

// GenaiModel abstracts *genai.GenerativeModel.
type GenaiModel interface {
	SetTemperature(temp float32)
	SetSystemInstruction(instruction string)
	SetSystemInstructionContent(content *genai.Content)
	SetTools(tools []*genai.Tool)
	StartChat() MessageSender
	GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// MessageSender abstracts *genai.ChatSession for sending messages.
type MessageSender interface {
	SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error)
}

// sdkClient implements GenaiClient wrapping the concrete SDK Client.
type sdkClient struct {
	client *genai.Client
}

// NewSDKClient creates a new GenaiClient wrapping a real SDK Client.
func NewSDKClient(client *genai.Client) (GenaiClient, error) {
	if client == nil {
		return nil, ErrNilClient
	}
	return &sdkClient{client: client}, nil
}

// GenerativeModel implements GenaiClient.
func (c *sdkClient) GenerativeModel(model string) GenaiModel {
	return &sdkModel{model: c.client.GenerativeModel(model)}
}

// Close implements GenaiClient.
func (c *sdkClient) Close() error {
	return c.client.Close()
}

// sdkModel implements GenaiModel wrapping the concrete SDK GenerativeModel.
type sdkModel struct {
	model *genai.GenerativeModel
}

// SetTemperature implements GenaiModel.
func (m *sdkModel) SetTemperature(temp float32) {
	m.model.SetTemperature(temp)
}

// SetSystemInstruction implements GenaiModel.
func (m *sdkModel) SetSystemInstruction(instruction string) {
	m.model.SystemInstruction = genai.NewUserContent(genai.Text(instruction))
}

// SetSystemInstructionContent implements GenaiModel.
func (m *sdkModel) SetSystemInstructionContent(content *genai.Content) {
	m.model.SystemInstruction = content
}

// SetTools implements GenaiModel.
func (m *sdkModel) SetTools(tools []*genai.Tool) {
	m.model.Tools = tools
}

// StartChat implements GenaiModel.
func (m *sdkModel) StartChat() MessageSender {
	return &sdkChatSession{session: m.model.StartChat()}
}

// GenerateContent implements GenaiModel.
func (m *sdkModel) GenerateContent(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return m.model.GenerateContent(ctx, parts...)
}

// sdkChatSession implements MessageSender wrapping the concrete SDK ChatSession.
type sdkChatSession struct {
	session *genai.ChatSession
}

// SendMessage implements MessageSender.
func (s *sdkChatSession) SendMessage(ctx context.Context, parts ...genai.Part) (*genai.GenerateContentResponse, error) {
	return s.session.SendMessage(ctx, parts...)
}
