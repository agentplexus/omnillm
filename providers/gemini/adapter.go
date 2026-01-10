// Package gemini provides Google Gemini provider adapter for the OmniLLM unified interface
package gemini

import (
	"context"
	"io"

	"github.com/agentplexus/omnillm/provider"
)

// Provider represents the Gemini provider adapter
type Provider struct {
	client *Client
}

// NewProvider creates a new Gemini provider adapter
func NewProvider(apiKey string) provider.Provider {
	client := New(apiKey)
	return &Provider{client: client}
}

// NewProviderWithContext creates a new Gemini provider adapter with context
func NewProviderWithContext(ctx context.Context, apiKey string) (provider.Provider, error) {
	client, err := NewWithContext(ctx, apiKey)
	if err != nil {
		return nil, err
	}
	return &Provider{client: client}, nil
}

// Name returns the provider name
func (p *Provider) Name() string {
	return p.client.Name()
}

// CreateChatCompletion creates a chat completion
func (p *Provider) CreateChatCompletion(ctx context.Context, req *provider.ChatCompletionRequest) (*provider.ChatCompletionResponse, error) {
	// Convert from unified format to Gemini format
	geminiReq := &Request{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stop:        req.Stop,
	}

	// Convert response format if provided
	if req.ResponseFormat != nil {
		geminiReq.ResponseFormat = &ResponseFormat{
			Type: req.ResponseFormat.Type,
		}
	}

	// Convert messages
	for _, msg := range req.Messages {
		geminiReq.Messages = append(geminiReq.Messages, Message{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		})
	}

	resp, err := p.client.CreateCompletion(ctx, geminiReq)
	if err != nil {
		return nil, err
	}

	// Convert back to unified format
	unifiedResp := &provider.ChatCompletionResponse{
		ID:      resp.ID,
		Object:  resp.Object,
		Created: resp.Created,
		Model:   resp.Model,
		Usage: provider.Usage{
			PromptTokens:     resp.Usage.PromptTokens,
			CompletionTokens: resp.Usage.CompletionTokens,
			TotalTokens:      resp.Usage.TotalTokens,
		},
	}

	// Convert choices
	for _, choice := range resp.Choices {
		unifiedChoice := provider.ChatCompletionChoice{
			Index: choice.Index,
			Message: provider.Message{
				Role:    provider.Role(choice.Message.Role),
				Content: choice.Message.Content,
				Name:    choice.Message.Name,
			},
			FinishReason: choice.FinishReason,
		}
		unifiedResp.Choices = append(unifiedResp.Choices, unifiedChoice)
	}

	return unifiedResp, nil
}

// CreateChatCompletionStream creates a streaming chat completion
func (p *Provider) CreateChatCompletionStream(ctx context.Context, req *provider.ChatCompletionRequest) (provider.ChatCompletionStream, error) {
	// Convert from unified format to Gemini format
	geminiReq := &Request{
		Model:       req.Model,
		MaxTokens:   req.MaxTokens,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		TopK:        req.TopK,
		Stop:        req.Stop,
	}

	// Convert response format if provided
	if req.ResponseFormat != nil {
		geminiReq.ResponseFormat = &ResponseFormat{
			Type: req.ResponseFormat.Type,
		}
	}

	// Convert messages
	for _, msg := range req.Messages {
		geminiReq.Messages = append(geminiReq.Messages, Message{
			Role:    string(msg.Role),
			Content: msg.Content,
			Name:    msg.Name,
		})
	}

	stream, err := p.client.CreateCompletionStream(ctx, geminiReq)
	if err != nil {
		return nil, err
	}

	return &StreamAdapter{stream: stream}, nil
}

// Close closes the provider
func (p *Provider) Close() error {
	return p.client.Close()
}

// StreamAdapter adapts Gemini stream to unified interface
type StreamAdapter struct {
	stream *Stream
}

// Recv receives the next chunk from the stream
func (s *StreamAdapter) Recv() (*provider.ChatCompletionChunk, error) {
	chunk, err := s.stream.Recv()
	if err != nil {
		if err == io.EOF {
			return nil, io.EOF
		}
		return nil, err
	}

	// Convert to unified format
	result := &provider.ChatCompletionChunk{
		ID:      chunk.ID,
		Object:  chunk.Object,
		Created: chunk.Created,
		Model:   chunk.Model,
	}

	if chunk.Usage != nil {
		result.Usage = &provider.Usage{
			PromptTokens:     chunk.Usage.PromptTokens,
			CompletionTokens: chunk.Usage.CompletionTokens,
			TotalTokens:      chunk.Usage.TotalTokens,
		}
	}

	for _, choice := range chunk.Choices {
		unifiedChoice := provider.ChatCompletionChoice{
			Index:        choice.Index,
			FinishReason: choice.FinishReason,
		}

		if choice.Delta != nil {
			unifiedChoice.Delta = &provider.Message{
				Role:    provider.Role(choice.Delta.Role),
				Content: choice.Delta.Content,
				Name:    choice.Delta.Name,
			}
		}

		result.Choices = append(result.Choices, unifiedChoice)
	}

	return result, nil
}

// Close closes the stream
func (s *StreamAdapter) Close() error {
	return s.stream.Close()
}
