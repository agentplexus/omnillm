# Release Notes - OmniLLM v0.11.0

**Release Date:** 2026-01-10
**Base Version:** v0.10.0

## Overview

Version 0.11.0 is a major feature release that adds four key reliability and cost optimization features: Fallback Providers, Circuit Breaker, Token Estimation, and Response Caching. This release also includes extended sampling parameters for fine-grained control over model outputs.

**Highlights:**

- **Fallback Providers**: Automatic failover to backup providers when primary fails
- **Circuit Breaker**: Prevent cascading failures by temporarily skipping unhealthy providers
- **Token Estimation**: Pre-flight validation to avoid context window limit errors
- **Response Caching**: Reduce API costs by caching identical requests
- **Extended Sampling Parameters**: TopK, Seed, N, ResponseFormat, Logprobs support

---

## New Features

### 1. Fallback Providers

Automatic failover to backup providers when the primary provider fails with retryable errors (rate limits, server errors, network issues).

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   "openai-key",
    FallbackProviders: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameAnthropic, APIKey: "anthropic-key"},
        {Provider: omnillm.ProviderNameGemini, APIKey: "gemini-key"},
    },
})

// If OpenAI fails, automatically tries Anthropic, then Gemini
response, err := client.CreateChatCompletion(ctx, request)
```

**Key Features:**

- Intelligent error classification (only retries on retryable errors)
- Auth errors (401/403) and invalid requests (400) do not trigger fallback
- `FallbackError` type provides detailed attempt tracking
- Works with both sync and streaming APIs

### 2. Circuit Breaker

Prevents cascading failures by temporarily skipping providers that are failing repeatedly.

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   "openai-key",
    FallbackProviders: []omnillm.ProviderConfig{...},
    CircuitBreakerConfig: &omnillm.CircuitBreakerConfig{
        FailureThreshold: 5,               // Open after 5 consecutive failures
        SuccessThreshold: 2,               // Close after 2 successes in half-open
        Timeout:          30 * time.Second, // Wait before trying again
    },
})
```

**Circuit States:**

| State | Description |
|-------|-------------|
| Closed | Normal operation, requests flow through |
| Open | Provider is failing, requests skip immediately |
| Half-Open | Testing if provider has recovered |

### 3. Token Estimation

Pre-flight token counting to validate requests before sending to the API.

```go
// Standalone estimation
estimator := omnillm.NewTokenEstimator(omnillm.DefaultTokenEstimatorConfig())
tokens, _ := estimator.EstimateTokens("gpt-4o", messages)
window := estimator.GetContextWindow("gpt-4o") // 128000

// Automatic validation in client
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider:       omnillm.ProviderNameOpenAI,
    APIKey:         "your-key",
    TokenEstimator: omnillm.NewTokenEstimator(omnillm.DefaultTokenEstimatorConfig()),
    ValidateTokens: true,
})
```

**Built-in Context Windows:**

- 40+ models supported (OpenAI, Anthropic, Gemini, X.AI, Ollama)
- Custom context windows via `CustomContextWindows` map
- Configurable characters-per-token ratio

### 4. Response Caching

Cache identical requests to reduce API costs with configurable TTL.

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   "your-key",
    Cache:    kvsClient, // Redis, DynamoDB, etc.
    CacheConfig: &omnillm.CacheConfig{
        TTL:       1 * time.Hour,
        KeyPrefix: "myapp:llm-cache",
    },
})

// Check if response was cached
if resp.ProviderMetadata["cache_hit"] == true {
    // Response came from cache
}
```

**Cache Key Generation:**

- SHA-256 hash of model, messages, and parameters
- Configurable inclusion of temperature and seed in cache key
- Model allowlist for selective caching
- Streaming requests skipped by default

### 5. Extended Sampling Parameters

New parameters for fine-grained control over model outputs:

| Parameter | Type | Providers | Description |
|-----------|------|-----------|-------------|
| `TopK` | `*int` | Anthropic, Gemini, Ollama | Top K token selection |
| `Seed` | `*int` | OpenAI, X.AI, Ollama | Reproducible outputs |
| `N` | `*int` | OpenAI | Number of completions |
| `ResponseFormat` | `*ResponseFormat` | OpenAI, Gemini | JSON mode |
| `Logprobs` | `*bool` | OpenAI | Return log probabilities |
| `TopLogprobs` | `*int` | OpenAI | Top logprobs count |

---

## New Types

### Error Classification

```go
type ErrorCategory int

const (
    ErrorCategoryUnknown ErrorCategory = iota
    ErrorCategoryRetryable    // Rate limits, server errors, network errors
    ErrorCategoryNonRetryable // Auth errors, invalid requests
)

func ClassifyError(err error) ErrorCategory
func IsRetryableError(err error) bool
func IsNonRetryableError(err error) bool
```

### Token Types

```go
type TokenEstimator interface {
    EstimateTokens(model string, messages []provider.Message) (int, error)
    GetContextWindow(model string) int
}

type TokenLimitError struct {
    EstimatedTokens int
    ContextWindow   int
    AvailableTokens int
    Model           string
}
```

### Fallback Types

```go
type FallbackError struct {
    Attempts  []FallbackAttempt
    LastError error
}

type FallbackAttempt struct {
    Provider string
    Error    error
    Duration time.Duration
}
```

---

## Updated ClientConfig

```go
type ClientConfig struct {
    // Existing fields...

    // NEW: Fallback & Circuit Breaker
    FallbackProviders    []ProviderConfig
    CircuitBreakerConfig *CircuitBreakerConfig

    // NEW: Token Estimation
    TokenEstimator TokenEstimator
    ValidateTokens bool

    // NEW: Response Caching
    Cache       kvs.Client
    CacheConfig *CacheConfig
}
```

---

## Upgrade Guide

### From v0.10.0

No breaking changes. All new features are opt-in.

```bash
go get github.com/agentplexus/omnillm@v0.11.0
go mod tidy
```

### Enable Fallback Providers

```go
// Before
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   apiKey,
})

// After (with fallback)
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   apiKey,
    FallbackProviders: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameAnthropic, APIKey: anthropicKey},
    },
})
```

### Enable Token Validation

```go
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider:       omnillm.ProviderNameOpenAI,
    APIKey:         apiKey,
    TokenEstimator: omnillm.NewTokenEstimator(omnillm.DefaultTokenEstimatorConfig()),
    ValidateTokens: true,
})
```

### Enable Response Caching

```go
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   apiKey,
    Cache:    kvsClient, // Your KVS implementation
    CacheConfig: &omnillm.CacheConfig{
        TTL: 1 * time.Hour,
    },
})
```

---

## New Files

| File | Description |
|------|-------------|
| `circuitbreaker.go` | Circuit breaker implementation |
| `circuitbreaker_test.go` | Circuit breaker tests |
| `fallback.go` | Fallback provider wrapper |
| `fallback_test.go` | Fallback tests |
| `tokens.go` | Token estimation |
| `tokens_test.go` | Token estimation tests |
| `cache.go` | Response caching |
| `cache_test.go` | Cache tests |

---

## Test Coverage

- Main package: **72.7%** coverage
- New feature code: **78-95%** coverage
- 45+ unit tests

---

## Performance Considerations

### Fallback Providers

- Fallback adds minimal latency when primary succeeds
- Circuit breaker prevents unnecessary attempts to failing providers
- Consider provider ordering by latency/cost

### Token Estimation

- Character-based estimation is fast but approximate
- Actual token count may vary by 5-15%
- Use conservative estimates for critical applications

### Response Caching

- Cache lookups add ~1ms latency
- TTL should match your freshness requirements
- Consider memory/storage costs for cache backend

---

## Related Documentation

- [README.md](README.md) - Full feature documentation
- [CHANGELOG.md](CHANGELOG.md) - Complete change history
- [ROADMAP.md](ROADMAP.md) - Future plans
