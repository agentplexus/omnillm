# Release Notes - OmniLLM v0.9.0

**Release Date:** 2025-12-27
**Base Version:** v0.8.0

## Overview

Version 0.9.0 moves the module to the agentplexus organization and renames it to `omnillm`, establishing the project's permanent home.

**Summary:**

- **Module Rename**: `github.com/grokify/metallm` → `github.com/agentplexus/omnillm`
- **Organization Move**: Project now maintained under the agentplexus organization
- **Dependency Updates**: All dependencies updated via `go mod tidy`

---

## Breaking Changes

### Module Rename and Organization Move

The module has been renamed and moved to a new organization:

**Before:**
```go
import "github.com/grokify/metallm"

client, err := metallm.NewClient(metallm.ClientConfig{...})
```

**After:**
```go
import "github.com/agentplexus/omnillm"

client, err := omnillm.NewClient(omnillm.ClientConfig{...})
```

**Migration:**

1. Update import paths: `github.com/grokify/metallm` → `github.com/agentplexus/omnillm`
2. Update type prefixes: `metallm.` → `omnillm.`
3. Update go.mod: `go get github.com/agentplexus/omnillm@v0.9.0`

---

## Improvements

### Dependency Updates

All Go module dependencies have been updated to their latest stable versions.

---

## Upgrade Guide

### From v0.8.0

1. Update import paths from `github.com/grokify/metallm` to `github.com/agentplexus/omnillm`
2. Update type prefixes from `metallm.` to `omnillm.`
3. Update go.mod dependency

```bash
go get github.com/agentplexus/omnillm@v0.9.0
go mod tidy
```

### Notes

- This is the final module rename. The `github.com/agentplexus/omnillm` path is the permanent home for this project.
- All previous module names (`gollm`, `fluxllm`, `metallm`) are deprecated.
