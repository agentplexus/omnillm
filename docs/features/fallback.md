# Fallback & Circuit Breaker

OmniLLM supports automatic failover to backup providers and circuit breaker patterns to handle provider failures gracefully.

## Fallback Providers

```go
// Providers[0] is primary, Providers[1+] are fallbacks
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOpenAI, APIKey: "openai-key"},       // Primary
        {Provider: omnillm.ProviderNameAnthropic, APIKey: "anthropic-key"}, // Fallback 1
        {Provider: omnillm.ProviderNameGemini, APIKey: "gemini-key"},       // Fallback 2
    },
})

// If OpenAI fails with a retryable error, automatically tries Anthropic, then Gemini
response, err := client.CreateChatCompletion(ctx, request)
```

## Error Classification

Fallback uses intelligent error classification:

| Error Type | Triggers Fallback |
|------------|-------------------|
| Rate limits (429) | Yes |
| Server errors (5xx) | Yes |
| Network errors | Yes |
| Auth errors (401/403) | No |
| Invalid requests (400) | No |

## Circuit Breaker

The circuit breaker pattern prevents cascading failures by temporarily skipping providers that are unhealthy.

### Configuration

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOpenAI, APIKey: "openai-key"},
        {Provider: omnillm.ProviderNameAnthropic, APIKey: "anthropic-key"},
    },
    CircuitBreakerConfig: &omnillm.CircuitBreakerConfig{
        FailureThreshold:     5,               // Open after 5 consecutive failures
        SuccessThreshold:     2,               // Close after 2 successes in half-open
        Timeout:              30 * time.Second, // Wait before trying again
        FailureRateThreshold: 0.5,             // 50% failure rate opens circuit
        MinimumRequests:      10,              // Minimum requests for rate calculation
    },
})
```

### States

- **Closed**: Normal operation, requests flow through
- **Open**: Provider is failing, requests skip it immediately
- **Half-Open**: Testing if provider has recovered

### State Transitions

```
         success          failure
   ┌──────────────┐   ┌──────────────┐
   │              │   │              │
   ▼              │   ▼              │
CLOSED ─────────► OPEN ─────────► HALF-OPEN
   ▲    failures     │   timeout      │
   │                 │                │
   └─────────────────┴────────────────┘
         success         failure
```
