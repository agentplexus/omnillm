# Observability

OmniLLM supports observability hooks for tracing, logging, and metrics without modifying the core library.

## ObservabilityHook Interface

```go
type ObservabilityHook interface {
    // BeforeRequest is called before each LLM call
    BeforeRequest(ctx context.Context, info LLMCallInfo, req *provider.ChatCompletionRequest) context.Context

    // AfterResponse is called after each LLM call completes
    AfterResponse(ctx context.Context, info LLMCallInfo, req *provider.ChatCompletionRequest, resp *provider.ChatCompletionResponse, err error)

    // WrapStream wraps a stream for observability
    WrapStream(ctx context.Context, info LLMCallInfo, req *provider.ChatCompletionRequest, stream provider.ChatCompletionStream) provider.ChatCompletionStream
}

type LLMCallInfo struct {
    CallID       string    // Unique identifier for correlating
    ProviderName string    // e.g., "openai", "anthropic"
    StartTime    time.Time // When the call started
}
```

## Simple Logging Hook

```go
type LoggingHook struct{}

func (h *LoggingHook) BeforeRequest(ctx context.Context, info omnillm.LLMCallInfo, req *omnillm.ChatCompletionRequest) context.Context {
    log.Printf("[%s] LLM call started: provider=%s model=%s", info.CallID, info.ProviderName, req.Model)
    return ctx
}

func (h *LoggingHook) AfterResponse(ctx context.Context, info omnillm.LLMCallInfo, req *omnillm.ChatCompletionRequest, resp *omnillm.ChatCompletionResponse, err error) {
    duration := time.Since(info.StartTime)
    if err != nil {
        log.Printf("[%s] LLM call failed: duration=%v error=%v", info.CallID, duration, err)
    } else {
        log.Printf("[%s] LLM call completed: duration=%v tokens=%d", info.CallID, duration, resp.Usage.TotalTokens)
    }
}

func (h *LoggingHook) WrapStream(ctx context.Context, info omnillm.LLMCallInfo, req *omnillm.ChatCompletionRequest, stream omnillm.ChatCompletionStream) omnillm.ChatCompletionStream {
    return stream
}

// Use the hook
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers: []omnillm.ProviderConfig{
        {Provider: omnillm.ProviderNameOpenAI, APIKey: "your-api-key"},
    },
    ObservabilityHook: &LoggingHook{},
})
```

## OpenTelemetry Integration

```go
type OTelHook struct {
    tracer trace.Tracer
}

func (h *OTelHook) BeforeRequest(ctx context.Context, info omnillm.LLMCallInfo, req *omnillm.ChatCompletionRequest) context.Context {
    ctx, span := h.tracer.Start(ctx, "llm.chat_completion",
        trace.WithAttributes(
            attribute.String("llm.provider", info.ProviderName),
            attribute.String("llm.model", req.Model),
        ),
    )
    return ctx
}

func (h *OTelHook) AfterResponse(ctx context.Context, info omnillm.LLMCallInfo, req *omnillm.ChatCompletionRequest, resp *omnillm.ChatCompletionResponse, err error) {
    span := trace.SpanFromContext(ctx)
    defer span.End()

    if err != nil {
        span.RecordError(err)
        span.SetStatus(codes.Error, err.Error())
    } else if resp != nil {
        span.SetAttributes(
            attribute.Int("llm.tokens.total", resp.Usage.TotalTokens),
            attribute.Int("llm.tokens.prompt", resp.Usage.PromptTokens),
            attribute.Int("llm.tokens.completion", resp.Usage.CompletionTokens),
        )
    }
}
```

## OmniObserve Integration

For full LLM observability, use [OmniObserve](https://github.com/agentplexus/omniobserve):

```go
import "github.com/agentplexus/omniobserve/integrations/omnillm"

// Create omnillm hook from omniobserve provider
hook := omnillm.NewHook(omniobserveProvider)

client, err := omnillm.NewClient(omnillm.ClientConfig{
    Providers:         []omnillm.ProviderConfig{...},
    ObservabilityHook: hook,
})
```
