# Token Estimation

OmniLLM provides pre-flight token estimation to validate requests before sending them to the API.

## Basic Usage

```go
// Create estimator with default config
estimator := omnillm.NewTokenEstimator(omnillm.DefaultTokenEstimatorConfig())

// Estimate tokens for messages
tokens, err := estimator.EstimateTokens("gpt-4o", messages)

// Get model's context window
window := estimator.GetContextWindow("gpt-4o") // Returns 128000
```

## Automatic Validation

Enable automatic token validation in client:

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOpenAI, APIKey: "your-key"},
    },
    TokenEstimator: omnillm.NewTokenEstimator(omnillm.DefaultTokenEstimatorConfig()),
    ValidateTokens: true, // Rejects requests that exceed context window
})

// Returns TokenLimitError if request exceeds model limits
response, err := client.CreateChatCompletion(ctx, request)
if tlErr, ok := err.(*omnillm.TokenLimitError); ok {
    fmt.Printf("Request has %d tokens, but model only supports %d\n",
        tlErr.EstimatedTokens, tlErr.ContextWindow)
}
```

## Built-in Context Windows

| Provider | Models | Context Window |
|----------|--------|----------------|
| OpenAI | GPT-4o, GPT-4o-mini | 128,000 |
| OpenAI | o1 | 200,000 |
| Anthropic | Claude 3/3.5/4 | 200,000 |
| Google | Gemini 2.5 | 1,000,000 |
| Google | Gemini 1.5 Pro | 2,000,000 |
| X.AI | Grok 3/4 | 128,000 |

## Custom Configuration

```go
config := omnillm.TokenEstimatorConfig{
    CharactersPerToken: 3.5, // More conservative estimate
    CustomContextWindows: map[string]int{
        "my-custom-model": 500000,
        "gpt-4o":          200000, // Override built-in
    },
}
estimator := omnillm.NewTokenEstimator(config)
```
