# OmniLLM Feature Roadmap

## High Value

### 1. Retry with Backoff ✅
Automatic retries for transient failures (rate limits, 5xx errors).

**Status:** Implemented in v0.7.0 via `ClientConfig.HTTPClient` with `retryhttp.RetryTransport`.

```go
rt := retryhttp.NewWithOptions(
    retryhttp.WithMaxRetries(5),
    retryhttp.WithInitialBackoff(500 * time.Millisecond),
)
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider:   omnillm.ProviderNameOpenAI,
    APIKey:     "...",
    HTTPClient: &http.Client{Transport: rt},
})
```

### 2. Request Timeouts ✅
Per-request timeout configuration.

**Status:** Implemented in v0.10.0 via `ClientConfig.Timeout`.

```go
ClientConfig{
    Timeout: 300 * time.Second,  // Recommended for reasoning models
}
```

### 3. Extended Sampling Parameters ✅
Additional sampling and generation parameters beyond Temperature and TopP.

**Status:** Implemented in v0.11.0 via `ChatCompletionRequest` fields.

| Parameter | Providers | Description |
|-----------|-----------|-------------|
| `TopK` | Anthropic, Gemini, Ollama | Limits token selection to top K candidates |
| `Seed` | OpenAI, X.AI, Ollama | Enables reproducible outputs |
| `N` | OpenAI | Number of completions to generate |
| `ResponseFormat` | OpenAI, Gemini | JSON mode (`{"type": "json_object"}`) |
| `Logprobs` | OpenAI | Return log probabilities of output tokens |
| `TopLogprobs` | OpenAI | Number of most likely tokens to return |

```go
req := &omnillm.ChatCompletionRequest{
    Model:    omnillm.ModelGPT4o,
    Messages: messages,
    TopK:     ptr(40),                                    // Anthropic, Gemini, Ollama
    Seed:     ptr(42),                                    // OpenAI, X.AI, Ollama
    ResponseFormat: &omnillm.ResponseFormat{Type: "json_object"}, // OpenAI, Gemini
}
```

### 4. Fallback Providers
Automatic failover when primary provider fails.

```go
ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    FallbackProviders: []ProviderConfig{
        {Provider: omnillm.ProviderNameAnthropic, APIKey: "..."},
    },
}
```

## Medium Value

### 5. Rate Limiting
Client-side rate limiter to respect provider limits.

### 6. Token Counting/Estimation
Estimate tokens before sending to avoid limit errors.

### 7. Response Caching
Cache identical requests to reduce costs (with TTL).

### 8. Circuit Breaker
Prevent cascading failures when provider is unhealthy.

## Nice to Have

### 9. Embeddings API
Unified interface for text embeddings.

### 10. Structured Output Validation
JSON schema validation for responses.

### 11. Batch Processing
Efficient batch request handling.

---

## Areas for Improvement

### Testing

- **Expand test coverage** - Integration tests currently require API keys; consider adding more mock-based unit tests
- **Add streaming tests** - Some providers lack comprehensive streaming response tests
- **Provider parity testing** - Ensure all providers have equivalent test coverage

### Documentation

- **Error type documentation** - More comprehensive docs on error types and handling patterns
- **Provider-specific quirks** - Document provider differences and edge cases
- **Migration guides** - Add guides for migrating from provider-specific SDKs

### Infrastructure

- **CI/CD pipeline** - Add GitHub Actions workflows for automated testing, linting, and releases
- **Coverage reporting** - Integrate coverage badges and reports into CI
- **Cross-platform testing** - Verify builds on Linux, macOS, and Windows

### Code Quality

- **Consistent error wrapping** - Ensure all errors include sufficient context
- **Interface compliance tests** - Add compile-time checks for provider interface compliance
- **Benchmarks** - Add performance benchmarks for critical paths

---

## Questions to Consider

- What's the primary use case? (chatbot, batch processing, real-time?)
- Is cost optimization important? (caching, token counting)
- How critical is uptime? (fallbacks, circuit breaker)
