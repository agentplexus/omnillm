# Anthropic (Claude)

## Overview

- **Models**: Claude-Opus-4.1, Claude-Opus-4, Claude-Sonnet-4, Claude-3.7-Sonnet, Claude-3.5-Haiku, Claude-3-Opus, Claude-3-Sonnet, Claude-3-Haiku
- **Features**: Chat completions, streaming, system message support

## Configuration

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameAnthropic, APIKey: "your-anthropic-api-key"},
    },
})
```

## Available Models

| Model | Context Window | Description |
|-------|----------------|-------------|
| `omnillm.ModelClaudeOpus41` | 200K | Claude Opus 4.1 (most capable) |
| `omnillm.ModelClaudeOpus4` | 200K | Claude Opus 4 |
| `omnillm.ModelClaudeSonnet4` | 200K | Claude Sonnet 4 |
| `omnillm.ModelClaude37Sonnet` | 200K | Claude 3.7 Sonnet |
| `omnillm.ModelClaude35Haiku` | 200K | Claude 3.5 Haiku (fast) |
| `omnillm.ModelClaude3Opus` | 200K | Claude 3 Opus |
| `omnillm.ModelClaude3Sonnet` | 200K | Claude 3 Sonnet |
| `omnillm.ModelClaude3Haiku` | 200K | Claude 3 Haiku |

## TopK Sampling

Anthropic supports TopK sampling:

```go
response, err := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: omnillm.ModelClaude3Sonnet,
    Messages: messages,
    TopK: &[]int{40}[0], // Consider only top 40 tokens
})
```

## System Messages

System messages are fully supported:

```go
messages := []omnillm.Message{
    {Role: omnillm.RoleSystem, Content: "You are a helpful assistant."},
    {Role: omnillm.RoleUser, Content: "Hello!"},
}
```
