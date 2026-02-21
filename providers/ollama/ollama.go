// Package ollama provides Ollama API client implementation
package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client implements Ollama API client
type Client struct {
	baseURL string
	client  *http.Client
}

// New creates a new Ollama client
func New(baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 60 * time.Second} // Longer timeout for local models
	}

	return &Client{
		baseURL: baseURL,
		client:  httpClient,
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "ollama"
}

// CreateCompletion creates a chat completion
func (c *Client) CreateCompletion(ctx context.Context, req *Request) (*Response, error) {
	if req.Model == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

	req.Stream = boolPtr(false)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq) //nolint:gosec // G704: baseURL is configured at client init, not user-controlled per-request
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil {
			return nil, fmt.Errorf("ollama API error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &response, nil
}

// CreateCompletionStream creates a streaming chat completion
func (c *Client) CreateCompletionStream(ctx context.Context, req *Request) (*Stream, error) {
	if req.Model == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

	req.Stream = boolPtr(true)

	body, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq) //nolint:gosec // G704: baseURL is configured at client init, not user-controlled per-request
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		var errorResp ErrorResponse
		if json.Unmarshal(body, &errorResp) == nil {
			return nil, fmt.Errorf("ollama API error: %s", errorResp.Error)
		}
		return nil, fmt.Errorf("ollama API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return &Stream{
		reader: bufio.NewReader(resp.Body),
		closer: resp.Body,
	}, nil
}

// Close closes the client (no-op for Ollama)
func (c *Client) Close() error {
	return nil
}

// Stream represents a streaming response from Ollama
type Stream struct {
	reader io.Reader
	closer io.Closer
}

// Recv receives the next chunk from the stream
func (s *Stream) Recv() (*StreamResponse, error) {
	scanner := bufio.NewScanner(s.reader)
	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, err
		}
		return nil, io.EOF
	}

	line := scanner.Text()
	if line == "" {
		return nil, io.EOF
	}

	var chunk StreamResponse
	if err := json.Unmarshal([]byte(line), &chunk); err != nil {
		return nil, fmt.Errorf("failed to decode stream chunk: %w", err)
	}

	return &chunk, nil
}

// Close closes the stream
func (s *Stream) Close() error {
	return s.closer.Close()
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
