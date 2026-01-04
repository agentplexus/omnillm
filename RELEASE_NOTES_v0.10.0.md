# Release Notes - OmniLLM v0.10.0

**Release Date:** 2026-01-04
**Base Version:** v0.9.0

## Overview

Version 0.10.0 adds configurable HTTP client timeout support and Claude 4.5 model constants. The timeout configuration is essential for reasoning models that may take longer to respond.

**Summary:**

- **Configurable Timeout**: New `ClientConfig.Timeout` field for setting HTTP client timeout
- **Claude 4.5 Models**: Added model constants for Claude Opus 4.5, Sonnet 4.5, and Haiku 4.5
- **HTTP Client Refactor**: Unified HTTP client creation with `getHTTPClient()` helper

---

## New Features

### 1. Configurable HTTP Client Timeout

Added `Timeout` field to `ClientConfig` for configuring HTTP client timeout. This is particularly important for reasoning models (like xAI Grok reasoning models) that may need more time to complete.

**Usage:**
```go
import (
    "time"
    "github.com/agentplexus/omnillm"
)

client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameXAI,
    APIKey:   os.Getenv("XAI_API_KEY"),
    Timeout:  300 * time.Second, // 5 minutes for reasoning models
})
```

**Behavior:**

- If `Timeout` is set and `HTTPClient` is nil, a new HTTP client is created with the specified timeout
- If `HTTPClient` is provided, it takes precedence (for custom transports with retry logic, etc.)
- If neither is set, providers use their default timeouts

**Recommendation:** Use `300 * time.Second` (5 minutes) for reasoning models.

### 2. Claude 4.5 Model Constants

Added model constants for the Claude 4.5 family:

```go
import "github.com/agentplexus/omnillm/models"

// Available Claude 4.5 models
models.ClaudeOpus4_5    // "claude-opus-4-5-20251101"
models.ClaudeSonnet4_5  // "claude-sonnet-4-5-20250929"
models.ClaudeHaiku4_5   // "claude-haiku-4-5-20251001"
```

### 3. Documentation

Added Marp presentation and GitHub Pages HTML documentation for the project.

---

## Improvements

### HTTP Client Handling Refactor

The HTTP client creation logic has been refactored with a new `getHTTPClient()` helper function that:

1. Returns the custom `HTTPClient` if provided
2. Creates a new client with `Timeout` if specified
3. Returns nil to let providers use their defaults

This provides a unified and consistent HTTP client handling across all providers (OpenAI, Anthropic, xAI, Ollama).

---

## Upgrade Guide

### From v0.9.0

No breaking changes. To use the new timeout feature:

```go
// Before (provider default timeout)
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameXAI,
    APIKey:   apiKey,
})

// After (custom timeout for reasoning models)
client, err := omnillm.NewClient(omnillm.ClientConfig{
    Provider: omnillm.ProviderNameXAI,
    APIKey:   apiKey,
    Timeout:  300 * time.Second,
})
```

```bash
go get github.com/agentplexus/omnillm@v0.10.0
go mod tidy
```

---

## Provider Default Timeouts

For reference, the default timeouts when no `Timeout` is configured:

| Provider   | Default Timeout |
|------------|----------------|
| OpenAI     | 30 seconds     |
| Anthropic  | 30 seconds     |
| xAI        | 60 seconds     |
| Ollama     | 60 seconds     |
| Gemini     | (SDK default)  |

**Note:** These defaults may be too short for reasoning models. Set `Timeout: 300 * time.Second` for longer-running inference tasks.
