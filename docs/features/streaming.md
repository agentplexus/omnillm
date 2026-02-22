# Streaming

OmniLLM supports real-time response streaming for all providers.

## Basic Streaming

```go
stream, err := client.CreateChatCompletionStream(context.Background(), &omnillm.ChatCompletionRequest{
    Model: omnillm.ModelGPT4o,
    Messages: []omnillm.Message{
        {Role: omnillm.RoleUser, Content: "Tell me a short story about AI."},
    },
    MaxTokens:   &[]int{200}[0],
    Temperature: &[]float64{0.8}[0],
})
if err != nil {
    log.Fatal(err)
}
defer stream.Close()

fmt.Print("AI Response: ")
for {
    chunk, err := stream.Recv()
    if err == io.EOF {
        break
    }
    if err != nil {
        log.Fatal(err)
    }

    if len(chunk.Choices) > 0 && chunk.Choices[0].Delta != nil {
        fmt.Print(chunk.Choices[0].Delta.Content)
    }
}
fmt.Println()
```

## Stream Interface

```go
type ChatCompletionStream interface {
    // Recv receives the next chunk from the stream
    Recv() (*ChatCompletionStreamResponse, error)

    // Close closes the stream
    Close() error
}
```

## Provider Support

| Provider | Streaming |
|----------|-----------|
| OpenAI | Yes |
| Anthropic | Yes (SSE) |
| Google Gemini | Yes |
| X.AI | Yes |
| Ollama | Yes |
| AWS Bedrock | Yes |

## Streaming with Observability

When using observability hooks, wrap the stream to track streaming metrics:

```go
func (h *MyHook) WrapStream(ctx context.Context, info omnillm.LLMCallInfo, req *omnillm.ChatCompletionRequest, stream omnillm.ChatCompletionStream) omnillm.ChatCompletionStream {
    return &observableStream{
        stream:    stream,
        ctx:       ctx,
        info:      info,
        startTime: time.Now(),
    }
}
```
