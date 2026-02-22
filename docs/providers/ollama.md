# Ollama

## Overview

- **Models**: Llama 3, Mistral, CodeLlama, Gemma, Qwen2.5, DeepSeek-Coder
- **Features**: Local inference, no API keys required, optimized for Apple Silicon

## Configuration

```go
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOllama, BaseURL: "http://localhost:11434"},
    },
})
```

## Running Ollama

1. Install Ollama from [ollama.ai](https://ollama.ai)
2. Pull a model: `ollama pull llama3`
3. Ollama runs automatically on `localhost:11434`

## Available Models

| Model | Description |
|-------|-------------|
| `llama3` | Meta's Llama 3 |
| `llama3:70b` | Llama 3 70B (larger) |
| `mistral` | Mistral 7B |
| `codellama` | Code-specialized Llama |
| `gemma` | Google's Gemma |
| `qwen2.5` | Alibaba's Qwen 2.5 |
| `deepseek-coder` | DeepSeek Coder |

## Example

```go
response, err := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: "llama3",
    Messages: []omnillm.Message{
        {Role: omnillm.RoleUser, Content: "Explain quantum computing simply."},
    },
})
```

## Streaming

```go
stream, err := client.CreateChatCompletionStream(ctx, &omnillm.ChatCompletionRequest{
    Model: "llama3",
    Messages: messages,
})

for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

## Custom Ollama Server

Connect to a remote Ollama instance:

```go
{Provider: omnillm.ProviderNameOllama, BaseURL: "http://192.168.1.100:11434"}
```
