# OpenAI

## Overview

- **Models**: GPT-5, GPT-4.1, GPT-4o, GPT-4o-mini, GPT-4-turbo, GPT-3.5-turbo
- **Features**: Chat completions, streaming, function/tool calling

## Configuration

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOpenAI, APIKey: "your-openai-api-key"},
    },
})
```

## Available Models

| Model | Context Window | Description |
|-------|----------------|-------------|
| `omnillm.ModelGPT5` | 200K | Latest GPT-5 model |
| `omnillm.ModelGPT41` | 128K | GPT-4.1 |
| `omnillm.ModelGPT4o` | 128K | GPT-4o (recommended) |
| `omnillm.ModelGPT4oMini` | 128K | GPT-4o Mini (cost-effective) |
| `omnillm.ModelGPT4Turbo` | 128K | GPT-4 Turbo |
| `omnillm.ModelGPT35Turbo` | 16K | GPT-3.5 Turbo |

## Tool Calling

OpenAI supports function/tool calling for agentic workflows:

```go
response, err := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: omnillm.ModelGPT4o,
    Messages: []omnillm.Message{
        {Role: omnillm.RoleUser, Content: "What's the weather in Tokyo?"},
    },
    Tools: []omnillm.Tool{
        {
            Type: "function",
            Function: omnillm.ToolFunction{
                Name:        "get_weather",
                Description: "Get current weather for a location",
                Parameters: map[string]any{
                    "type": "object",
                    "properties": map[string]any{
                        "location": map[string]any{
                            "type":        "string",
                            "description": "City name",
                        },
                    },
                    "required": []string{"location"},
                },
            },
        },
    },
})
```

## Custom Endpoint

Use a custom OpenAI-compatible endpoint:

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {
            Provider: omnillm.ProviderNameOpenAI,
            APIKey:   "your-api-key",
            BaseURL:  "https://your-custom-endpoint.com/v1",
        },
    },
})
```
