// Package anthropic provides Anthropic Claude API client implementation
package anthropic

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Client implements Anthropic API client
type Client struct {
	apiKey  string
	baseURL string
	client  *http.Client
}

// New creates a new Anthropic client
func New(apiKey, baseURL string, httpClient *http.Client) *Client {
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	if httpClient == nil {
		httpClient = &http.Client{Timeout: 30 * time.Second}
	}

	return &Client{
		apiKey:  apiKey,
		baseURL: baseURL,
		client:  httpClient,
	}
}

// Name returns the provider name
func (c *Client) Name() string {
	return "anthropic"
}

// CreateCompletion creates a chat completion
func (c *Client) CreateCompletion(ctx context.Context, req *Request) (*Response, error) {
	if req.Model == "" {
		return nil, fmt.Errorf("model cannot be empty")
	}
	if len(req.Messages) == 0 {
		return nil, fmt.Errorf("messages cannot be empty")
	}

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)

	resp, err := c.client.Do(httpReq) //nolint:gosec // G704: baseURL is configured at client init, not user-controlled per-request
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, c.handleErrorResponse(resp)
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

	// Enable streaming
	req.Stream = boolPtr(true)

	reqBody, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/v1/messages", bytes.NewBuffer(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	c.setHeaders(httpReq)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := c.client.Do(httpReq) //nolint:gosec // G704: baseURL is configured at client init, not user-controlled per-request
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.handleErrorResponse(resp)
	}

	return &Stream{
		response: resp,
		scanner:  bufio.NewScanner(resp.Body),
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	return nil
}

// setHeaders sets the required headers for Anthropic API requests
func (c *Client) setHeaders(req *http.Request) {
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.apiKey)
	req.Header.Set("anthropic-version", "2023-06-01")
}

// handleErrorResponse handles error responses from Anthropic API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response")
	}

	var errorResp struct {
		Error struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &errorResp); err != nil {
		return fmt.Errorf("api error: %s", string(body))
	}

	return fmt.Errorf("anthropic api error: %s", errorResp.Error.Message)
}

// Stream implements streaming for Anthropic
type Stream struct {
	response *http.Response
	scanner  *bufio.Scanner
	closed   bool
}

// Recv receives the next chunk from the stream
func (s *Stream) Recv() (*StreamEvent, error) {
	if s.closed {
		return nil, fmt.Errorf("stream is closed")
	}

	var currentEvent string
	var currentData strings.Builder

	for s.scanner.Scan() {
		line := s.scanner.Text()

		// Empty line indicates end of event
		if line == "" {
			if currentEvent != "" && currentData.Len() > 0 {
				// Parse the event data
				var event StreamEvent
				if err := json.Unmarshal([]byte(currentData.String()), &event); err != nil {
					// Reset for next event
					currentEvent = ""
					currentData.Reset()
					continue
				}

				// Only return events we care about
				if event.Type == "content_block_delta" || event.Type == "message_start" ||
					event.Type == "message_delta" || event.Type == "message_stop" {
					return &event, nil
				}

				// Reset for next event
				currentEvent = ""
				currentData.Reset()
			}
			continue
		}

		// Parse event type
		if strings.HasPrefix(line, "event: ") {
			currentEvent = strings.TrimPrefix(line, "event: ")
			continue
		}

		// Parse data
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			currentData.WriteString(data)
			continue
		}
	}

	if err := s.scanner.Err(); err != nil {
		return nil, fmt.Errorf("stream error: %w", err)
	}

	return nil, io.EOF
}

// Close closes the stream
func (s *Stream) Close() error {
	if !s.closed {
		s.closed = true
		return s.response.Body.Close()
	}
	return nil
}

// boolPtr creates a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
