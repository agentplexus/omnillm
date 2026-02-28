# OmniLLM Models Catalog

This package provides a comprehensive catalog of LLM model identifiers and documentation references for all supported providers.

## Purpose

- **Centralized Model Constants**: All model IDs in one organized location
- **Documentation URLs**: Official reference pages for each provider
- **Easy Updates**: When providers release new models, update in one place
- **Type Safety**: Use constants instead of hardcoded strings

## Usage

### Import the Package

```go
import "github.com/plexusone/omnillm/models"
```

### Use Model Constants

```go
// Anthropic Claude
client, _ := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameAnthropic,
    APIKey:   apiKey,
})
response, _ := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: models.ClaudeOpus4_1,
    Messages: messages,
})

// OpenAI
response, _ := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: models.GPT4o,
    Messages: messages,
})

// X.AI Grok
response, _ := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: models.Grok4_1FastReasoning,
    Messages: messages,
})

// Google Gemini
response, _ := client.CreateChatCompletion(ctx, &omnillm.ChatCompletionRequest{
    Model: models.Gemini2_5Pro,
    Messages: messages,
})
```

### Access Documentation URLs

```go
// Print documentation URLs for reference
fmt.Println("Anthropic Models:", models.AnthropicModelsURL)
fmt.Println("OpenAI Models:", models.OpenAIModelsURL)
fmt.Println("X.AI Models:", models.XAIModelsURL)
fmt.Println("Gemini Models:", models.GeminiModelsURL)
fmt.Println("Bedrock Models:", models.BedrockModelsURL)
fmt.Println("Ollama Models:", models.OllamaModelsURL)
```

## Package Structure

```
models/
├── doc.go          # Package documentation
├── README.md       # This file
├── anthropic.go    # Claude models + docs URL
├── openai.go       # OpenAI models + docs URL
├── xai.go          # X.AI Grok models + docs URL
├── gemini.go       # Google Gemini models + docs URL
├── bedrock.go      # AWS Bedrock models + docs URL
├── ollama.go       # Ollama models + docs URL
└── vertex.go       # Google Vertex AI models + docs URL
```

## Updating Models

When providers release new models or deprecate existing ones:

1. **Check Documentation**: Use the provider's `ModelsURL` constant to visit their docs
2. **Update Constants**: Add new models or mark deprecated ones
3. **Update Constants Package**: Update root `constants.go` if needed for backwards compatibility
4. **Update Tests**: Update integration tests to use latest models
5. **Update Examples**: Update example code to showcase new models

### Example Update Workflow

```bash
# 1. Check X.AI documentation
open https://docs.x.ai/docs/models

# 2. Add new model constants to models/xai.go
# 3. Update constants.go for backwards compatibility
# 4. Update tests and examples
# 5. Update RELEASE_NOTES.md
```

## Model Categories by Provider

### Anthropic Claude

- **Latest**: Claude Opus 4.1, Claude Opus 4, Claude Sonnet 4
- **Current**: Claude 3.7 Sonnet, Claude 3.5 Haiku
- **Legacy**: Claude 3 Opus, Sonnet, Haiku
- **Documentation**: https://docs.anthropic.com/en/docs/about-claude/models

### OpenAI

- **Latest**: GPT-5, GPT-4.1
- **Current**: GPT-4o, GPT-4o Mini
- **Legacy**: GPT-4 Turbo, GPT-3.5 Turbo
- **Documentation**: https://platform.openai.com/docs/models

### X.AI Grok

- **Latest**: Grok 4.1 Fast (Reasoning/Non-Reasoning)
- **Current**: Grok 4 (0709, Fast variants), Grok Code Fast
- **Previous**: Grok 3, Grok 3 Mini
- **Legacy**: Grok 2, Grok 2 Vision
- **Documentation**: https://docs.x.ai/docs/models

### Google Gemini

- **Latest**: Gemini 2.5 Pro, Gemini 2.5 Flash
- **Current**: Gemini 1.5 Pro, Gemini 1.5 Flash
- **Legacy**: Gemini Pro
- **Documentation**: https://ai.google.dev/gemini-api/docs/models/gemini

### AWS Bedrock

- **Claude Models**: Opus 4, Claude 3 Opus, Claude 3 Sonnet
- **Amazon Models**: Titan Text Express
- **Documentation**: https://docs.aws.amazon.com/bedrock/latest/userguide/models-supported.html

### Ollama (Local Models)

- **Llama**: Llama 3 8B, Llama 3 70B
- **Mistral**: Mistral 7B, Mixtral 8x7B
- **Code**: CodeLlama, DeepSeek Coder
- **Other**: Gemma, Qwen 2.5
- **Documentation**: https://ollama.com/library

## Benefits

### 1. Type Safety
```go
// ✅ Good: Using constants (typo protection, autocomplete)
model := models.ClaudeOpus4_1

// ❌ Bad: Hardcoded strings (prone to typos)
model := "claude-opus-4-1-20250805"
```

### 2. Documentation
```go
// Every model has inline comments explaining features
// Hover over constant in IDE to see documentation
models.Grok4_1FastReasoning  // Shows: "Best tool-calling model with 2M context"
```

### 3. Centralized Updates
```go
// When a model ID changes or is deprecated, update in one place
// All code using the constant automatically uses new value
```

### 4. Easy Discovery
```go
// IDE autocomplete shows all available models
models.Claude  // → Claude3Haiku, Claude3Opus, ClaudeOpus4, ClaudeOpus4_1, etc.
models.GPT     // → GPT35Turbo, GPT4Turbo, GPT4o, GPT4oMini, GPT5, etc.
models.Grok    // → Grok2_1212, Grok3, Grok3Mini, Grok4_0709, etc.
```

## Compatibility

The root `constants.go` file maintains backwards compatibility by re-exporting models from this package:

```go
// constants.go
const (
    // For backwards compatibility, these re-export from models package
    ModelClaudeOpus4 = models.ClaudeOpus4
    ModelGPT4o = models.GPT4o
    ModelGrok4_1FastReasoning = models.Grok4_1FastReasoning
)
```

Existing code continues to work, but new code should import `models` package directly for better organization.

## Contributing

When adding support for a new provider:

1. Create `models/<provider>.go` file
2. Add `<Provider>ModelsURL` constant
3. Add `<Provider>APIURL` constant
4. Add model constants with documentation comments
5. Update this README.md
6. Update root `constants.go` for backwards compatibility
7. Add tests if needed

## License

MIT - See root LICENSE file
