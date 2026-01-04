package omnillm

import (
	"net/http"

	"github.com/agentplexus/omnillm/provider"
	"github.com/agentplexus/omnillm/providers/anthropic"
	"github.com/agentplexus/omnillm/providers/gemini"
	"github.com/agentplexus/omnillm/providers/ollama"
	"github.com/agentplexus/omnillm/providers/openai"
	"github.com/agentplexus/omnillm/providers/xai"
)

// getHTTPClient returns the HTTPClient from config, or creates one with the
// configured Timeout. Returns nil if neither is set (provider will use defaults).
func getHTTPClient(config ClientConfig) *http.Client {
	if config.HTTPClient != nil {
		return config.HTTPClient
	}
	if config.Timeout > 0 {
		return &http.Client{Timeout: config.Timeout}
	}
	return nil
}

// newOpenAIProvider creates a new OpenAI provider adapter
func newOpenAIProvider(config ClientConfig) (provider.Provider, error) {
	if config.APIKey == "" {
		return nil, ErrEmptyAPIKey
	}
	return openai.NewProvider(config.APIKey, config.BaseURL, getHTTPClient(config)), nil
}

// newAnthropicProvider creates a new Anthropic provider adapter
func newAnthropicProvider(config ClientConfig) (provider.Provider, error) {
	if config.APIKey == "" {
		return nil, ErrEmptyAPIKey
	}
	return anthropic.NewProvider(config.APIKey, config.BaseURL, getHTTPClient(config)), nil
}

// newOllamaProvider creates a new Ollama provider adapter
func newOllamaProvider(config ClientConfig) (provider.Provider, error) { //nolint:unparam // `error` added to fulfill interface requirements
	return ollama.NewProvider(config.BaseURL, getHTTPClient(config)), nil
}

// newGeminiProvider creates a new Gemini provider adapter
func newGeminiProvider(config ClientConfig) (provider.Provider, error) {
	if config.APIKey == "" {
		return nil, ErrEmptyAPIKey
	}
	return gemini.NewProvider(config.APIKey), nil
}

// newXAIProvider creates a new X.AI provider adapter
func newXAIProvider(config ClientConfig) (provider.Provider, error) {
	if config.APIKey == "" {
		return nil, ErrEmptyAPIKey
	}
	return xai.NewProvider(config.APIKey, config.BaseURL, getHTTPClient(config)), nil
}
