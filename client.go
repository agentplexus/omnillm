package omnillm

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/grokify/mogo/log/slogutil"
	"github.com/grokify/sogo/database/kvs"

	"github.com/agentplexus/omnillm/provider"
)

// ChatClient is the main client interface that wraps a Provider
type ChatClient struct {
	provider provider.Provider
	memory   *MemoryManager
	hook     ObservabilityHook
	logger   *slog.Logger
}

// ClientConfig holds configuration for creating a client
type ClientConfig struct {
	Provider ProviderName
	APIKey   string
	BaseURL  string
	Region   string // For AWS Bedrock

	// Timeout sets the HTTP client timeout for API calls.
	// If zero, providers use their default timeouts.
	// Recommended: 300 * time.Second for reasoning models.
	Timeout time.Duration

	// HTTPClient is an optional HTTP client with custom transport (e.g., retry transport).
	// If nil, providers will create clients using the Timeout value above.
	// This can be used to add retry logic, tracing, or other middleware.
	// Example with retry:
	//   rt := retryhttp.NewWithOptions(retryhttp.WithMaxRetries(3))
	//   config.HTTPClient = &http.Client{Transport: rt}
	HTTPClient *http.Client

	// Memory configuration (optional)
	Memory       kvs.Client
	MemoryConfig *MemoryConfig

	// Direct provider injection (for 3rd party providers)
	CustomProvider provider.Provider

	// ObservabilityHook is called before/after LLM calls (optional)
	ObservabilityHook ObservabilityHook

	// Logger for internal logging (optional, defaults to null logger)
	Logger *slog.Logger

	// Provider-specific configurations can be added here
	Extra map[string]any
}

// NewClient creates a new ChatClient based on the provider
func NewClient(config ClientConfig) (*ChatClient, error) {
	var prov provider.Provider
	var err error

	// Check for direct provider injection first
	if config.CustomProvider != nil {
		prov = config.CustomProvider
	} else {
		// Fall back to built-in providers
		switch config.Provider {
		case ProviderNameOpenAI:
			prov, err = newOpenAIProvider(config)
		case ProviderNameAnthropic:
			prov, err = newAnthropicProvider(config)
		case ProviderNameBedrock:
			return nil, ErrBedrockExternal
		case ProviderNameOllama:
			prov, err = newOllamaProvider(config)
		case ProviderNameGemini:
			prov, err = newGeminiProvider(config)
		case ProviderNameXAI:
			prov, err = newXAIProvider(config)
		default:
			return nil, ErrUnsupportedProvider
		}

		if err != nil {
			return nil, err
		}
	}

	// Initialize logger (default to null logger if not provided)
	logger := config.Logger
	if logger == nil {
		logger = slogutil.Null()
	}

	client := &ChatClient{
		provider: prov,
		hook:     config.ObservabilityHook,
		logger:   logger,
	}

	// Initialize memory if provided
	if config.Memory != nil {
		memoryConfig := DefaultMemoryConfig()
		if config.MemoryConfig != nil {
			memoryConfig = *config.MemoryConfig
		}
		client.memory = NewMemoryManager(config.Memory, memoryConfig)
	}

	return client, nil
}

// CreateChatCompletion creates a chat completion
func (c *ChatClient) CreateChatCompletion(ctx context.Context, req *provider.ChatCompletionRequest) (*provider.ChatCompletionResponse, error) {
	info := LLMCallInfo{
		CallID:       newCallID(),
		ProviderName: c.provider.Name(),
		StartTime:    time.Now(),
	}

	// Hook: before request
	if c.hook != nil {
		ctx = c.hook.BeforeRequest(ctx, info, req)
	}

	resp, err := c.provider.CreateChatCompletion(ctx, req)

	// Hook: after response
	if c.hook != nil {
		c.hook.AfterResponse(ctx, info, req, resp, err)
	}

	return resp, err
}

// CreateChatCompletionStream creates a streaming chat completion
func (c *ChatClient) CreateChatCompletionStream(ctx context.Context, req *provider.ChatCompletionRequest) (provider.ChatCompletionStream, error) {
	info := LLMCallInfo{
		CallID:       newCallID(),
		ProviderName: c.provider.Name(),
		StartTime:    time.Now(),
	}

	// Hook: before request
	if c.hook != nil {
		ctx = c.hook.BeforeRequest(ctx, info, req)
	}

	stream, err := c.provider.CreateChatCompletionStream(ctx, req)
	if err != nil {
		if c.hook != nil {
			c.hook.AfterResponse(ctx, info, req, nil, err)
		}
		return nil, err
	}

	// Hook: wrap stream for observability
	if c.hook != nil {
		stream = c.hook.WrapStream(ctx, info, req, stream)
	}

	return stream, nil
}

// Close closes the client
func (c *ChatClient) Close() error {
	return c.provider.Close()
}

// Provider returns the underlying provider
func (c *ChatClient) Provider() provider.Provider {
	return c.provider
}

// Memory returns the memory manager (nil if not configured)
func (c *ChatClient) Memory() *MemoryManager {
	return c.memory
}

// HasMemory returns true if memory is configured
func (c *ChatClient) HasMemory() bool {
	return c.memory != nil
}

// Logger returns the client's logger
func (c *ChatClient) Logger() *slog.Logger {
	return c.logger
}

// CreateChatCompletionWithMemory creates a chat completion using conversation memory
func (c *ChatClient) CreateChatCompletionWithMemory(ctx context.Context, sessionID string, req *provider.ChatCompletionRequest) (*provider.ChatCompletionResponse, error) {
	if !c.HasMemory() {
		return c.CreateChatCompletion(ctx, req)
	}

	// Load existing conversation
	conversation, err := c.memory.LoadConversation(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Merge stored messages with request messages
	allMessages := append(conversation.Messages, req.Messages...)

	// Create new request with combined messages
	memoryReq := *req
	memoryReq.Messages = allMessages

	// Get response (use client method to ensure hook is called)
	response, err := c.CreateChatCompletion(ctx, &memoryReq)
	if err != nil {
		return nil, err
	}

	// Save the conversation with new messages and response
	if len(response.Choices) > 0 {
		// Save request messages and response
		messagesToSave := append(req.Messages, response.Choices[0].Message)
		err = c.memory.AppendMessages(ctx, sessionID, messagesToSave)
		if err != nil {
			slogutil.LoggerFromContext(ctx, c.logger).Error("failed to save conversation to memory",
				slog.String("session_id", sessionID),
				slog.String("error", err.Error()))
		}
	}

	return response, nil
}

// CreateChatCompletionStreamWithMemory creates a streaming chat completion using conversation memory
func (c *ChatClient) CreateChatCompletionStreamWithMemory(ctx context.Context, sessionID string, req *provider.ChatCompletionRequest) (provider.ChatCompletionStream, error) {
	if !c.HasMemory() {
		return c.CreateChatCompletionStream(ctx, req)
	}

	// Load existing conversation
	conversation, err := c.memory.LoadConversation(ctx, sessionID)
	if err != nil {
		return nil, err
	}

	// Merge stored messages with request messages
	allMessages := append(conversation.Messages, req.Messages...)

	// Create new request with combined messages
	memoryReq := *req
	memoryReq.Messages = allMessages

	// Get stream response (use client method to ensure hook is called)
	stream, err := c.CreateChatCompletionStream(ctx, &memoryReq)
	if err != nil {
		return nil, err
	}

	// Wrap the stream to capture the response for memory storage
	return &memoryAwareStream{
		stream:      stream,
		memory:      c.memory,
		sessionID:   sessionID,
		reqMessages: req.Messages,
		ctx:         ctx,
		logger:      c.logger,
	}, nil
}

// LoadConversation loads a conversation from memory
func (c *ChatClient) LoadConversation(ctx context.Context, sessionID string) (*ConversationMemory, error) {
	if !c.HasMemory() {
		return nil, fmt.Errorf("memory not configured")
	}
	return c.memory.LoadConversation(ctx, sessionID)
}

// SaveConversation saves a conversation to memory
func (c *ChatClient) SaveConversation(ctx context.Context, conversation *ConversationMemory) error {
	if !c.HasMemory() {
		return fmt.Errorf("memory not configured")
	}
	return c.memory.SaveConversation(ctx, conversation)
}

// AppendMessage appends a message to a conversation in memory
func (c *ChatClient) AppendMessage(ctx context.Context, sessionID string, message provider.Message) error {
	if !c.HasMemory() {
		return fmt.Errorf("memory not configured")
	}
	return c.memory.AppendMessage(ctx, sessionID, message)
}

// GetConversationMessages retrieves messages from a conversation
func (c *ChatClient) GetConversationMessages(ctx context.Context, sessionID string) ([]provider.Message, error) {
	if !c.HasMemory() {
		return nil, fmt.Errorf("memory not configured")
	}
	return c.memory.GetMessages(ctx, sessionID)
}

// CreateConversationWithSystemMessage creates a new conversation with a system message
func (c *ChatClient) CreateConversationWithSystemMessage(ctx context.Context, sessionID, systemMessage string) error {
	if !c.HasMemory() {
		return fmt.Errorf("memory not configured")
	}
	return c.memory.CreateConversationWithSystemMessage(ctx, sessionID, systemMessage)
}

// DeleteConversation removes a conversation from memory
func (c *ChatClient) DeleteConversation(ctx context.Context, sessionID string) error {
	if !c.HasMemory() {
		return fmt.Errorf("memory not configured")
	}
	return c.memory.DeleteConversation(ctx, sessionID)
}

// memoryAwareStream wraps a ChatCompletionStream to capture responses for memory storage
type memoryAwareStream struct {
	stream      provider.ChatCompletionStream
	memory      *MemoryManager
	sessionID   string
	reqMessages []provider.Message
	ctx         context.Context
	logger      *slog.Logger

	// Buffer to collect the complete response
	responseBuffer strings.Builder
	streamClosed   bool
}

// Recv receives the next chunk from the stream and buffers the response
func (s *memoryAwareStream) Recv() (*provider.ChatCompletionChunk, error) {
	chunk, err := s.stream.Recv()
	if err != nil {
		// If we hit EOF and haven't saved the response yet, save it now
		if err.Error() == "EOF" && !s.streamClosed {
			s.saveBufferedResponse()
			s.streamClosed = true
		}
		return chunk, err
	}

	// Buffer the response content
	if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
		s.responseBuffer.WriteString(chunk.Choices[0].Delta.Content)
	}

	return chunk, nil
}

// Close closes the stream and saves the complete response to memory
func (s *memoryAwareStream) Close() error {
	if !s.streamClosed {
		s.saveBufferedResponse()
		s.streamClosed = true
	}
	return s.stream.Close()
}

// saveBufferedResponse saves the complete buffered response to memory
func (s *memoryAwareStream) saveBufferedResponse() {
	if s.responseBuffer.Len() > 0 {
		// Create assistant message from buffered response
		assistantMessage := provider.Message{
			Role:    provider.RoleAssistant,
			Content: s.responseBuffer.String(),
		}

		// Save request messages and response
		messagesToSave := append(s.reqMessages, assistantMessage)
		err := s.memory.AppendMessages(s.ctx, s.sessionID, messagesToSave)
		if err != nil {
			slogutil.LoggerFromContext(s.ctx, s.logger).Error("failed to save streaming response to memory",
				slog.String("session_id", s.sessionID),
				slog.String("error", err.Error()))
		}
	}
}
