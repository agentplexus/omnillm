# Release Notes - MetaLLM v0.8.0

**Release Date:** 2025-12-22
**Base Version:** v0.7.1

## Overview

Version 0.8.0 renames the module from `fluxllm` to `metallm` for consistency with the `metaserp` naming convention and adds project direction documentation.

**Summary:**

- **Module Rename**: `github.com/grokify/fluxllm` → `github.com/grokify/metallm`
- **Documentation**: Added ROADMAP.md with project direction

---

## Breaking Changes

### Module Rename

The module has been renamed from `fluxllm` to `metallm` for consistency with `metaserp`:

**Before:**
```go
import "github.com/grokify/fluxllm"

client, err := fluxllm.NewClient(fluxllm.ClientConfig{...})
```

**After:**
```go
import "github.com/grokify/metallm"

client, err := metallm.NewClient(metallm.ClientConfig{...})
```

**Migration:**

1. Update import paths: `github.com/grokify/fluxllm` → `github.com/grokify/metallm`
2. Update type prefixes: `fluxllm.` → `metallm.`
3. Update go.mod: `go get github.com/grokify/metallm@v0.8.0`

---

## New Features

### ROADMAP.md

Added project roadmap documentation outlining the direction and planned features for the unified LLM SDK.

---

## Upgrade Guide

### From v0.7.1

1. Update import paths from `fluxllm` to `metallm`
2. Update go.mod dependency
3. Run `go mod tidy`

```bash
go get github.com/grokify/metallm@v0.8.0
go mod tidy
```
