# Phase 2 — LLM-Powered Sentence Generation

## Status: ✅ Complete

All 38 tests pass. All 15 manual smoke tests pass.

## Progress

- [x] Sub-phase 1: Config — config.go + config.json + main.go changes
- [x] Sub-phase 2: Session — SentencePair type + WithPairs option
- [x] Sub-phase 3: Sengen — vocab.go + grammar.md + prompt.tmpl + sengen.go (+ 8 tests)
- [x] Sub-phase 4: Handler — integrate sengen into CreateSessionHandler (+ mock generator)
- [x] Sub-phase 5: Wrap-up — run full test suite, final review

## Overview

Replace the hard-coded sentence list with a sentence generator ("sen-gen") that uses an LLM to dynamically produce practice sentences from a vocabulary list, grammar rules, and a prompt template.

---

## Project Structure (Additions)

```
backend/
├── config.json                        # LLM configuration
├── internal/
│   ├── config/
│   │   └── config.go                  # Config struct + JSON loader
│   ├── handler/
│   │   ├── sessions.go                # [modified] calls sengen.Generate
│   │   └── handler_test.go            # [modified] uses mock generator
│   ├── session/
│   │   └── session.go                 # [modified] SentencePair + WithPairs
│   └── sengen/                        # NEW: sentence generator
│       ├── sengen.go                  # Generator orchestrator + LLM client
│       ├── sengen_test.go             # 8 tests: prompt building + parsing
│       ├── vocab.go                   # 48 hard-coded vocabulary entries
│       ├── grammar.md                 # Grammar rules (embedded)
│       └── prompt.tmpl                # LLM prompt template (embedded)
├── go.mod
└── go.sum
```

## New Files

| File | Purpose |
|------|---------|
| `backend/config.json` | LLM config: `base_url`, `model`, `api_key` |
| `internal/config/config.go` | Config struct + JSON loader |
| `internal/sengen/sengen.go` | Orchestrator: `Generate(count int) ([]session.SentencePair, error)` |
| `internal/sengen/sengen_test.go` | 8 tests for prompt building + response parsing |
| `internal/sengen/vocab.go` | Hard-coded Korean/English vocab list (Go `var`) |
| `internal/sengen/grammar.md` | Grammar rules, embedded via `//go:embed` |
| `internal/sengen/prompt.tmpl` | Go `text/template` prompt, embedded via `//go:embed` |

## Modified Files

| File | Change |
|------|--------|
| `cmd/api/main.go` | Load `config.json` → `config.Config` → inject into `sengen` |
| `internal/session/session.go` | Add `SentencePair` type + `WithPairs()` option (maps to `Card` + shuffles internally) |
| `internal/handler/sessions.go` | `CreateSessionHandler` calls `sengen.Generate(10)`; error → 503; passes `WithPairs` |
| `internal/handler/routes.go` | `Initialize` now calls `sengen.Init` |
| `internal/handler/handler_test.go` | Uses `sengen.SetGenerateFunc` mock for testing |

## Dependency Graph

```
cmd/api/main.go
  → config (load config.json)
  → handler
      → session (SentencePair, WithPairs)
      → sengen (Generate)
          → session (SentencePair type)
          → config (LLM settings)
          → (embed) grammar.md + prompt.tmpl
          → (http) LLM API
```

## `config.json`

```json
{
  "llm": {
    "base_url": "http://localhost:11434/v1",
    "model": "qwen3:8b",
    "api_key": ""
  }
}
```

### Fields

| Field | Required | Notes |
|-------|----------|-------|
| `base_url` | Yes | Ollama default. Change to `https://api.openai.com/v1` for OpenAI. |
| `model` | Yes | `qwen3:8b` default. Change to swap models. |
| `api_key` | No | Leave empty for Ollama. Set for OpenAI. |

## `session.SentencePair` + `WithPairs`

```go
type SentencePair struct {
    Korean  string
    English string
}

func WithPairs(pairs []SentencePair) SessionOption
```

When `WithPairs` is provided, `NewSession` skips `HardCodedCards()` and:

1. Maps each pair → `Card{Front, Back}` based on `Direction`
2. Shuffles with `rand.Shuffle`
3. Sets `CardsRemaining = TotalCards = len(pairs)`

Existing tests without `WithPairs` still hit `HardCodedCards()` — no changes needed.

## Prompt Template (`prompt.tmpl`)

```
Create {{.Count}} Korean-English sentence pairs using the following constraints.
Respond with ONLY the sentence pairs, one per line, in this exact format without numbering or commentary:

"Korean sentence"
"English sentence"

# Grammatical Constraints
Use only the following grammatical rules. Do not add complexity.
{{.Grammar}}

# Vocab Constraints
Use only the vocabulary in this list.
{{.Vocab}}
```

### Template Data Struct

```go
type promptData struct {
    Count   int
    Grammar string
    Vocab   string
}
```

## `sengen.Generate` Flow

1. Read `grammar.md` and `prompt.tmpl` (compile-time embedded via `//go:embed`)
2. Read vocab from `vocab.go` constant
3. Render `prompt.tmpl` with grammar + vocab
4. POST to `/v1/chat/completions` at `config.LLM.BaseURL`
5. Parse response text into `[]session.SentencePair`
6. Return pairs (unshuffled, direction-agnostic)

## Error Handling

`sengen.Generate` fails → `CreateSessionHandler` returns HTTP 503:

```json
{"error": "sentence generation failed", "code": "GENERATION_FAILED"}
```

No retries, no fallback to hard-coded sentences.

## Usage

```bash
# Start Ollama (if not already running)
ollama serve

# Pull the model (first time only)
ollama pull qwen3:8b

# Edit config to pick model (optional — defaults to qwen3:8b)
vim backend/config.json

# Start server
cd backend && go run ./cmd/api/
```

### Swap Models

Edit `config.json`, change `model` field, restart the server.

### Switch to OpenAI

Edit `config.json`:

```json
{
  "llm": {
    "base_url": "https://api.openai.com/v1",
    "model": "gpt-4o-mini",
    "api_key": "sk-..."
  }
}
```

## Test Strategy

- `sengen_test.go`: unit-test prompt building + response parsing (mock the LLM call via interface or test helper)
- Existing session tests still pass (no `WithPairs` → `HardCodedCards` fallback)
- Existing handler tests still pass
