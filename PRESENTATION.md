---
marp: true
theme: agentplexus
paginate: true
---

<!-- _class: lead -->

# OmniLLM

## Unified Go SDK for Large Language Models

A single interface to OpenAI, Anthropic, Gemini, X.AI, Ollama & more

---

# The Problem

Building AI applications often requires:

- **Multiple LLM providers** for redundancy, cost optimization, or capability needs
- **Different APIs** with incompatible request/response formats
- **Vendor lock-in** when tightly coupled to one provider
- **Code duplication** for streaming, error handling, and conversation management

---

# The Solution: OmniLLM

**One unified API** that works across all major LLM providers

```go
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,  // Switch providers here
    APIKey:   "your-api-key",
})

response, _ := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model:    omnillm.ModelGPT4o,
    Messages: []omnillm.Message{{Role: omnillm.RoleUser, Content: "Hello!"}},
})
```

---

# Key Features

- **Multi-Provider Support** - OpenAI, Anthropic, Gemini, X.AI, Ollama, Bedrock
- **Unified API** - Same interface across all providers
- **Streaming Support** - Real-time response streaming
- **Conversation Memory** - Persistent history via Key-Value Stores
- **Observability Hooks** - Extensible tracing, logging, and metrics
- **Retry with Backoff** - Automatic retries for transient failures
- **Type Safe** - Full Go type safety with comprehensive error handling
- **Extensible** - Easy to add new providers

---

<!-- _class: section-divider -->

# Architecture

---

# Modular Design

```
omnillm/
├── client.go            # Main ChatClient wrapper
├── providers.go         # Factory functions for built-in providers
├── memory.go            # Conversation memory management
├── observability.go     # ObservabilityHook interface
├── provider/            # Public interface for external providers
│   ├── interface.go     # Provider interface definition
│   └── types.go         # Unified request/response types
└── providers/           # Individual provider implementations
    ├── openai/          # OpenAI implementation
    ├── anthropic/       # Anthropic (Claude) implementation
    ├── gemini/          # Google Gemini implementation
    ├── xai/             # X.AI (Grok) implementation
    └── ollama/          # Ollama implementation
```

---

# Provider Pattern

Each provider follows the same structure:

| File | Purpose |
|------|---------|
| `provider.go` | HTTP client implementation |
| `types.go` | Provider-specific request/response types |
| `adapter.go` | `provider.Provider` interface implementation |
| `*_test.go` | Unit and integration tests |

**Benefits**: Clean separation, testability, easy extension

---

# The Provider Interface

```go
type Provider interface {
    // Name returns the provider identifier
    Name() string

    // CreateChatCompletion performs a synchronous chat completion
    CreateChatCompletion(ctx context.Context,
        req *ChatCompletionRequest) (*ChatCompletionResponse, error)

    // CreateChatCompletionStream performs streaming chat completion
    CreateChatCompletionStream(ctx context.Context,
        req *ChatCompletionRequest) (ChatCompletionStream, error)

    // Close releases any resources
    Close() error
}
```

---

<!-- _class: section-divider -->

# Supported Providers

---

# Provider Coverage

| Provider | Models | Features |
|----------|--------|----------|
| **OpenAI** | GPT-5, GPT-4.1, GPT-4o, GPT-4o-mini | Chat, Streaming, Functions |
| **Anthropic** | Claude Opus 4.1, Sonnet 4, Haiku 3.5 | Chat, Streaming, System msgs |
| **Gemini** | Gemini 2.5 Pro/Flash, 1.5 Pro/Flash | Chat, Streaming |
| **X.AI** | Grok-4.1, Grok-4, Grok-Code | Chat, Streaming, 2M context |
| **Ollama** | Llama 3, Mistral, Qwen2.5 | Local inference |
| **Bedrock*** | Claude, Titan models | AWS-native |

*Available as external module

---

# Built-in vs External Providers

**Built-in providers**: Zero additional dependencies

```go
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   os.Getenv("OPENAI_API_KEY"),
})
```

**External providers**: Optional modules for heavy dependencies

```go
import "github.com/agentplexus/omnillm-bedrock"

bedrockProvider, _ := bedrock.NewProvider("us-east-1")
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    CustomProvider: bedrockProvider,
})
```

---

<!-- _class: section-divider -->

# Core Capabilities

---

# Streaming Responses

Real-time token streaming across all providers:

```go
stream, _ := client.CreateChatCompletionStream(ctx, &omnillm.ChatCompletionRequest{
    Model:    omnillm.ModelGPT4o,
    Messages: []omnillm.Message{{Role: omnillm.RoleUser, Content: "Tell a story"}},
})
defer stream.Close()

for {
    chunk, err := stream.Recv()
    if err == io.EOF { break }
    fmt.Print(chunk.Choices[0].Delta.Content)
}
```

---

# Conversation Memory

Persistent conversation history using any KVS backend:

```go
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider:     omnillm.ProviderNameOpenAI,
    APIKey:       "your-key",
    Memory:       redisKVS,  // Redis, DynamoDB, or custom
    MemoryConfig: &omnillm.MemoryConfig{
        MaxMessages: 50,
        TTL:         24 * time.Hour,
    },
})

// Automatically loads/saves conversation history
response, _ := client.CreateChatCompletionWithMemory(ctx, "session-123", req)
```

---

# Observability Hooks

Add tracing, logging, and metrics without modifying core code:

```go
type OTelHook struct { tracer trace.Tracer }

func (h *OTelHook) BeforeRequest(ctx context.Context, info LLMCallInfo,
    req *ChatCompletionRequest) context.Context {
    ctx, _ = h.tracer.Start(ctx, "llm.chat_completion",
        trace.WithAttributes(
            attribute.String("llm.provider", info.ProviderName),
            attribute.String("llm.model", req.Model),
        ))
    return ctx
}

func (h *OTelHook) AfterResponse(ctx context.Context, info LLMCallInfo,
    req *ChatCompletionRequest, resp *ChatCompletionResponse, err error) {
    span := trace.SpanFromContext(ctx)
    defer span.End()
}
```

---

# Retry with Backoff

Automatic retries for rate limits and transient errors:

```go
rt := retryhttp.NewWithOptions(
    retryhttp.WithMaxRetries(5),
    retryhttp.WithInitialBackoff(500 * time.Millisecond),
    retryhttp.WithMaxBackoff(30 * time.Second),
)

client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider:   omnillm.ProviderNameOpenAI,
    APIKey:     "...",
    HTTPClient: &http.Client{Transport: rt},
})
```

**Retries**: 429, 500, 502, 503, 504 | **Respects** `Retry-After` headers

---

<!-- _class: section-divider -->

# Extensibility

---

# Creating Custom Providers

Implement the `provider.Provider` interface:

```go
package myprovider

import "github.com/agentplexus/omnillm/provider"

type MyProvider struct { /* ... */ }

func (p *MyProvider) Name() string { return "myprovider" }

func (p *MyProvider) CreateChatCompletion(ctx context.Context,
    req *provider.ChatCompletionRequest) (*provider.ChatCompletionResponse, error) {
    // Your implementation
}

func (p *MyProvider) CreateChatCompletionStream(ctx context.Context,
    req *provider.ChatCompletionRequest) (provider.ChatCompletionStream, error) {
    // Your streaming implementation
}
```

---

# Injecting Custom Providers

Use your provider without modifying the core library:

```go
import (
    "github.com/agentplexus/omnillm"
    "github.com/yourname/omnillm-myprovider"
)

func main() {
    customProvider := myprovider.NewProvider("config")

    client, _ := omnillm.NewClient(omnillm.ClientConfig{
        CustomProvider: customProvider,
    })

    // Use the same unified API
    response, _ := client.CreateChatCompletion(ctx, req)
}
```

---

<!-- _class: section-divider -->

# Roadmap

---

# Feature Roadmap

| Priority | Feature | Status |
|----------|---------|--------|
| **High** | Retry with Backoff | Completed |
| **High** | Request Timeouts | Planned |
| **High** | Fallback Providers | Planned |
| **Medium** | Rate Limiting | Planned |
| **Medium** | Token Counting | Planned |
| **Medium** | Response Caching | Planned |
| **Medium** | Circuit Breaker | Planned |
| Nice to Have | Embeddings API | Planned |
| Nice to Have | Structured Output Validation | Planned |
| Nice to Have | Batch Processing | Planned |

---

# Fallback Providers (Planned)

Automatic failover when primary provider fails:

```go
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameOpenAI,
    APIKey:   openaiKey,
    FallbackProviders: []ProviderConfig{
        {Provider: omnillm.ProviderNameAnthropic, APIKey: anthropicKey},
        {Provider: omnillm.ProviderNameGemini, APIKey: geminiKey},
    },
})
```

**Use cases**: High availability, cost optimization, load balancing

---

<!-- _class: section-divider -->

# Getting Started

---

# Installation

```bash
go get github.com/agentplexus/omnillm
```

**Optional external providers:**

```bash
go get github.com/agentplexus/omnillm-bedrock
```

---

# Quick Start

```go
package main

import (
    "context"
    "fmt"
    "github.com/agentplexus/omnillm"
)

func main() {
    client, _ := omnillm.NewClient(omnillm.ClientConfig{
        Provider: omnillm.ProviderNameOpenAI,
        APIKey:   "your-api-key",
    })
    defer client.Close()

    response, _ := client.CreateChatCompletion(context.Background(),
        &omnillm.ChatCompletionRequest{
            Model:    omnillm.ModelGPT4o,
            Messages: []omnillm.Message{
                {Role: omnillm.RoleUser, Content: "Hello!"},
            },
        })

    fmt.Println(response.Choices[0].Message.Content)
}
```

---

# Running Examples

```bash
# Basic usage
go run examples/basic/main.go

# Streaming
go run examples/streaming/main.go
go run examples/anthropic_streaming/main.go

# Conversation memory
go run examples/memory_demo/main.go

# Provider-specific
go run examples/xai/main.go
go run examples/ollama/main.go
go run examples/gemini/main.go

# Custom providers
go run examples/custom_provider/main.go
```

---

<!-- _class: lead -->

# OmniLLM

**One SDK. All Providers. Zero Lock-in.**

```bash
go get github.com/agentplexus/omnillm
```

github.com/agentplexus/omnillm
