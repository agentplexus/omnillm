# Google Gemini

## Overview

- **Models**: Gemini-2.5-Pro, Gemini-2.5-Flash, Gemini-1.5-Pro, Gemini-1.5-Flash
- **Features**: Chat completions, streaming, massive context windows

## Configuration

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameGemini, APIKey: "your-gemini-api-key"},
    },
})
```

## Available Models

| Model | Context Window | Description |
|-------|----------------|-------------|
| `omnillm.ModelGemini25Pro` | 1M | Gemini 2.5 Pro |
| `omnillm.ModelGemini25Flash` | 1M | Gemini 2.5 Flash (fast) |
| `omnillm.ModelGemini15Pro` | 2M | Gemini 1.5 Pro (largest context) |
| `omnillm.ModelGemini15Flash` | 1M | Gemini 1.5 Flash |

## JSON Mode

Gemini supports JSON mode for structured outputs:

```go
response, err := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: omnillm.ModelGemini25Pro,
    Messages: messages,
    ResponseFormat: &omnillm.ResponseFormat{Type: "json_object"},
})
```

## Large Context

Gemini 1.5 Pro supports up to 2 million tokens of context, making it ideal for:

- Long document analysis
- Large codebase understanding
- Extended conversation history
