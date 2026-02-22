# Response Caching

OmniLLM supports response caching to reduce API costs for identical requests.

## Basic Usage

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOpenAI, APIKey: "your-key"},
    },
    Cache: kvsClient, // Your KVS implementation (Redis, DynamoDB, etc.)
    CacheConfig: &omnillm.CacheConfig{
        TTL:       1 * time.Hour,
        KeyPrefix: "myapp:llm-cache",
    },
})

// First call hits the API
response1, _ := client.CreateChatCompletion(ctx, request)

// Second identical call returns cached response
response2, _ := client.CreateChatCompletion(ctx, request)

// Check if response was from cache
if response2.ProviderMetadata["cache_hit"] == true {
    fmt.Println("Response was cached!")
}
```

## Configuration

```go
cacheConfig := &omnillm.CacheConfig{
    TTL:                1 * time.Hour,       // Time-to-live
    KeyPrefix:          "omnillm:cache",     // Key prefix
    SkipStreaming:      true,                // Don't cache streaming (default)
    CacheableModels:    []string{"gpt-4o"},  // Only cache specific models (nil = all)
    IncludeTemperature: true,                // Temperature affects cache key
    IncludeSeed:        true,                // Seed affects cache key
}
```

## Cache Key Generation

Cache keys are generated from a SHA-256 hash of:

- Model name
- Messages (role, content, name, tool_call_id)
- MaxTokens, Temperature, TopP, TopK, Seed, Stop sequences

Different parameter values = different cache keys.

## Cache Backends

Caching uses the same KVS backend as conversation memory:

- **Redis**: High-performance distributed caching
- **DynamoDB**: AWS-native caching
- **In-Memory**: Development and testing
- **Custom**: Any Sogo KVS implementation
