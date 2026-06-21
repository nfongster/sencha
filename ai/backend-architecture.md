# Backend Architecture

Stack: Go (Gin), PostgreSQL (pgx + golang-migrate + sqlc), in-memory fallback.

## System Overview

```
┌──────────┐     ┌───────────────────┐     ┌──────────────────────┐
│ console/ │────>│ backend API (Gin) │────>│ Repository interface │
│ (REPL)   │     │                   │     ├──────────────────────┤
└──────────┘     │ POST /api/sessions│     │ in-memory (maps)     │
                 │ GET  /api/levels  │     │ PostgreSQL (pgx+sqlc)│
┌──────────┐     │ PATCH /api/phases │     └──────────────────────┘
│ frontend │────>│ …16 routes total  │
│ (SPA)    │     └───────────────────┘
└──────────┘
```

- Sessions are **ephemeral** (global in-memory store, never persisted).
- Sentence generation is **LLM-based** (two-pass) with a hardcoded fallback.
- Config loaded from `config.json`; `database_url` selects repo backend.

---

## API Surface

### Health

| Method | Route | Response |
|--------|-------|----------|
| `GET` | `/api/health` | `{"status": "ok"}` |

### Prompt

| Method | Route | Request | Response |
|--------|-------|---------|----------|
| `GET` | `/api/prompt` | — | `{"text": string}` |
| `PUT` | `/api/prompt` | `{"text": string}` | `{"message": "prompt updated"}` |

**`PUT` validation:** text must be non-empty; returns `INVALID_PROMPT` (400) otherwise. Writes to `prompt.tmpl` on disk and calls `sengen.ReloadPrompt()`. Returns `PROMPT_SAVE_ERROR` (500) on write failure.

### Phases

| Method | Route | Request | Response |
|--------|-------|---------|----------|
| `GET` | `/api/phases` | — | `{"phases": [{number, name}]}` |
| `POST` | `/api/phases` | `{"number": int, "name": string}` | `{"message": "phase created"}` |
| `PATCH` | `/api/phases/:number` | `{"name": string}` | `{"message": "phase updated"}` |
| `DELETE` | `/api/phases/:number` | — | `{"message": "phase deleted"}` |
| `GET` | `/api/phases/:number/levels` | — | `{"levels": [{number, phase_number, grammar_md}]}` |

**Phase creation rule:** `number` must be between 1 and `max_phase + 1` (sequential).

### Levels

| Method | Route | Request | Response |
|--------|-------|---------|----------|
| `POST` | `/api/levels` | `{"phase_number": int, "grammar_markdown": string, "vocabulary": [{korean, english}]}` | `{"message": "level created", "level_number": int}` |
| `POST` | `/api/levels/extract-from-url` | `{"url": string}` | `{"grammar_markdown": string, "vocabulary": [{korean, english, category}]}` |
| `GET` | `/api/levels/max` | — | `{"max": int}` |
| `GET` | `/api/levels/:number` | — | `{"level": {number, phase_number, grammar_md}, "vocabulary": [{korean, english}], "level_vocabulary": [{korean, english}]}` |
| `PATCH` | `/api/levels/:number` | `{"grammar_markdown": string}` | `{"message": "level rules updated"}` |
| `PUT` | `/api/levels/:number/vocabulary` | `{"vocabulary": [{korean, english}]}` | `{"message": "vocabulary updated"}` |
| `DELETE` | `/api/levels/:number` | — | `{"message": "level deleted"}` |
| `GET` | `/api/levels/:number/sentences` | — | `{"sentences": [{korean, english}]}` |
| `GET` | `/api/levels/:number/sentences/count` | — | `{"count": int}` |
| `DELETE` | `/api/levels/:number/sentences` | — | `{"message": "sentences deleted"}` |
| `POST` | `/api/levels/:number/sentences/generate` | `{"count": int}` | `{"sentences": [{korean, english}]}` |

**Level creation rule:** auto-assigns `max_level + 1`. Grammar is required. Phase must exist.

**`PATCH /api/levels/:number`:** accepts `grammar_markdown` (required).

**`GET /api/levels/:number`:** returns three fields:
- `level` — the level metadata (grammar_md)
- `vocabulary` — cumulative vocab from levels 1..N
- `level_vocabulary` — vocab specific to this level only

### Sessions

| Method | Route | Request | Response |
|--------|-------|---------|----------|
| `POST` | `/api/sessions` | `{"direction": string, "level_number": int}` | `{"session_id": string, "direction": string, "total_cards": int, "cards_remaining": int, "session_complete": bool}` |
| `GET` | `/api/sessions/:id` | — | same shape |
| `POST` | `/api/sessions/:id/reveal` | — | `{"card_id": int, "front": string, "back": string}` |
| `POST` | `/api/sessions/:id/grade` | `{"grade": string}` | `{"cards_remaining": int, "session_complete": bool, "grade_counts": {"pass": int, "hard": int, "fail": int}}` |

**Direction** values: `korean-to-english` (default), `english-to-korean`, `mixed`.

**Grade** values: `pass`, `hard`, `fail`.

**Session creation flow:**
1. Fetch `LevelData` from repository for the given level (defaults to 1).
2. Try to get 10 random sentences from the DB for the level.
3. If fewer than 10 exist, generate the shortfall via `sengen.Generate(needed, levelData)` and save.
4. Create ephemeral session (in-memory store) with shuffled cards.
- If 10 sentences exist → no LLM call (fast path). If 0 → generates all 10 (current behavior).

### Error response format

All errors return `{"error": string, "code": string}` with appropriate HTTP status.

| Code | When |
|------|------|
| `INVALID_DIRECTION` | Bad direction value |
| `INVALID_LEVEL` | Level number not found |
| `GENERATION_FAILED` | LLM call failed |
| `SESSION_COMPLETE` | Already complete (409) |
| `CARD_NOT_REVEALED` | Grade called before reveal (409) |
| `INVALID_GRADE` | Bad grade value (400) |
| `NOT_FOUND` | Session ID not found |
| `PHASE_NOT_FOUND` | Phase number not found |
| `LEVEL_NOT_FOUND` | Level number not found |
| `INVALID_PHASE_NUMBER` | Phase number out of range |
| `INVALID_LEVEL_NUMBER` | Level number not parseable |
| `PHASE_CREATE_ERROR` | Phase creation conflict (duplicate) |
| `LEVEL_CREATE_ERROR` | Level creation conflict |
| `LEVEL_UPDATE_ERROR` | Level update failure |
| `VOCAB_ADD_ERROR` | Vocab insert failure after level creation |
| `VOCAB_UPDATE_ERROR` | Vocab replace failure |
| `VOCAB_ERROR` | Vocab fetch failure |
| `MISSING_NAME` | Name field required |
| `MISSING_GRAMMAR` | Grammar field required |

| `INVALID_REQUEST` | JSON parse failure |
| `INTERNAL_ERROR` | Unexpected server error |

---

## Repository

File: `backend/internal/repository/interface.go`

```go
type Repository interface {
    // Phases
    ListPhases() ([]Phase, error)
    CreatePhase(p Phase) error
    UpdatePhase(number int, name string) error
    DeletePhase(number int) error
    MaxPhaseNumber() (int, error)

    // Levels
    LevelsInPhase(phaseNumber int) ([]Level, error)
    Level(number int) (*Level, error)
    CreateLevel(l Level) error
    UpdateLevel(number int, grammarMD string) error
    DeleteLevel(number int) error
    MaxLevelNumber() (int, error)
    LevelsUpTo(number int) ([]Level, error)

    // Vocabulary
    VocabularyUpTo(levelNumber int) ([]VocabEntry, error)
    VocabularyForLevel(levelNumber int) ([]VocabEntry, error)
    AddVocabulary(levelNumber int, entries []VocabEntry) error
    SetVocabulary(levelNumber int, entries []VocabEntry) error

    // Sentences
    SaveSentences(sentences []Sentence) error
    SentencesForLevel(levelNumber int) ([]Sentence, error)
    CountSentencesForLevel(levelNumber int) (int, error)
    DeleteSentencesForLevel(levelNumber int) error
    RandomSentencesForLevel(levelNumber int, count int) ([]Sentence, error)

    // Convenience
    LoadLevelData(levelNumber int) (*LevelData, error)
}
```

### Two implementations

- **In-memory** (`memory.go`): map-based, thread-safe via `sync.RWMutex`. Auto-seeded via `Seed()` when no `database_url` is configured.
- **PostgreSQL** (`postgres.go`): wraps sqlc-generated `*db.Queries`, uses pgx pool. Selected when `database_url` is set.

### Seed data

`backend/internal/repository/seed.go` creates:
- Phase 1 ("Phase 1")
- Level 1 (in Phase 1, with ~40 lines of default grammar markdown covering basic sentence structure, the copula 이다, topic/object particles, and determiners/pronouns)
- 28 vocabulary entries (common nouns: 한국, 사람, 책, 학생, etc.)

---

## Data Model

### SQL tables

```sql
CREATE TABLE phases (
    number INTEGER PRIMARY KEY,
    name   TEXT NOT NULL
);

CREATE TABLE levels (
    number        INTEGER PRIMARY KEY,
    phase_number  INTEGER NOT NULL REFERENCES phases(number),
    grammar_md    TEXT NOT NULL
);

CREATE TABLE vocabulary (
    id            SERIAL PRIMARY KEY,
    level_number  INTEGER NOT NULL REFERENCES levels(number),
    korean        TEXT NOT NULL,
    english       TEXT NOT NULL
);

CREATE TABLE sentences (
    id            SERIAL PRIMARY KEY,
    level_number  INTEGER NOT NULL REFERENCES levels(number),
    korean        TEXT NOT NULL,
    english       TEXT NOT NULL,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
```

### Go structs (with JSON tags)

```go
type Phase struct {
    Number int    `json:"number"`
    Name   string `json:"name"`
}

type Level struct {
    Number      int    `json:"number"`
    PhaseNumber int    `json:"phase_number"`
    GrammarMD   string `json:"grammar_md"`
}

type VocabEntry struct {
    Korean  string `json:"korean"`
    English string `json:"english"`
}

type Sentence struct {
    LevelNumber int
    Korean      string
    English     string
}

type LevelData struct {
    GrammarMD string
    Vocab     []VocabEntry
}
```

**Levels are globally sequential** — they never reset per phase:
| Phase | Levels |
|-------|--------|
| 1     | 1, 2   |
| 2     | 3, 4, 5 |
| 3     | 6, 7, 8 |

Creation rules:
- New phase number must be ≤ `max_phase + 1`.
- New level number auto-assigns `max_level + 1`.
- `DeleteLevel` renumbers all following levels down by 1.
- `DeletePhase` cascades to levels, vocabulary, and sentences.

---

## Config

File: `backend/config.json`

```json
{
  "llm": {
    "base_url": "http://localhost:11434/v1",
    "model": "llama3.2"
  },
  "database_url": "postgres://postgres:sencha@localhost:5432/sencha?sslmode=disable"
}
```

Fields:
- `llm.base_url` — required, OpenAI-compatible chat completion endpoint (e.g., Ollama, OpenAI, vLLM).
- `llm.model` — required, model name.
- `llm.api_key` — optional, sent as `Authorization: Bearer <key>`.
- `database_url` — optional. When set, uses PostgreSQL; when omitted, uses in-memory repository.

The `Repository` field is injected at runtime by `main()` (tagged `json:"-"`).

---

## Session Model

File: `backend/internal/session/session.go`

```
Direction: korean-to-english | english-to-korean | mixed
Grade:     pass | hard | fail
```

**State machine per card:**
1. `Reveal()` — returns the current card (front/back). Sets `Revealed = true`.
2. `Grade(grade)` — records grade, advances to next card. Returns `GradeSummary`.

**Guards:**
- Grade before reveal → `ErrCardNotRevealed` (409).
- Action after completion → `ErrSessionComplete` (409).
- Invalid grade string → `ErrInvalidGrade` (400).

**Session options:**
- `WithDirection(dir)` — sets direction.
- `WithPairs(pairs)` — uses LLM-generated pairs (shuffled). Without this, falls back to `HardCodedCards()`.

Cards are shuffled on creation via `rand.Shuffle`.

**`mixed` direction:** randomly assigns each card as Korean→English or English→Korean (coin flip per card).

### Hardcoded fallback

File: `backend/internal/session/cards.go`

10 hardcoded Korean-English sentence pairs used when no `SentencePairs` are provided to `NewSession`.

---

## LLM Integration (sengen)

File: `backend/internal/sengen/sengen.go`

### Two-pass generation

```
Generate(count, levelData)
  │
  ├── 1. buildPrompt(grammar + 50 random vocab words)
  │     └── callLLM(prompt) → parseResponse(reply)
  │
  └── 2. grammarCheck(pairs)
        └── callLLM(checkPrompt) → parseResponse(reply)
        └── returns corrected pairs
```

### Prompt templates

The generation prompt (`prompt.tmpl`) is loaded from disk at startup via `os.ReadFile`, falling back to an embedded copy if the file is absent. It can be reloaded at runtime via `sengen.ReloadPrompt()` (triggered by `PUT /api/prompt`). The grammar checker prompt (`gramchecker.tmpl`) remains embedded and is not user-editable.

**`prompt.tmpl`** — Generation prompt:
```
Create N Korean-English sentence pairs using the following constraints.
Respond with ONLY the sentence pairs, one per line:
"Korean sentence 1"
"English sentence 1"
...
# Grammatical Constraints
{grammar_md}
# Vocab Constraints
{vocab list}
```

**`gramchecker.tmpl`** — Grammar check prompt:
```
Verify and correct each Korean-English sentence pair:
1. Is the Korean sentence grammatically correct? Fix it if not.
2. Is the English translation correct for the (corrected) Korean? Fix it if not.

Return ONLY the corrected list...
Pairs:
{pairs}
```

### LLM call details
- Uses OpenAI-compatible `/chat/completions` endpoint.
- System message: "You are a Korean language teacher creating practice sentences for students."
- HTTP client with 60s timeout.
- `api_key` sent as Bearer token when non-empty.
- Logs latency for each step.

### Test hook

```go
type GenerateFunc func(int, repository.LevelData) ([]session.SentencePair, error)
func SetGenerateFunc(fn GenerateFunc)
```

When `testGenerateFunc` is set, `Generate()` delegates to it instead of calling the real LLM. Used by handler tests.

### Vocab sampling

- Samples 50 words distributed across vocabulary categories (noun, pronoun, determiner, etc.).
- Partitioning: `perCat = 50 / numCategories`, remainder goes to the first category.
- If no categories exist, falls back to flat random sampling of 50 words.
- Postgres implementation: `UNION ALL` of per-category `(SELECT ... LIMIT $N)` subqueries, with outer `ORDER BY RANDOM()` when multiple categories.
- In-memory implementation: groups entries by category, shuffles and slices each group, then shuffles the combined result.

---

## Session Store

File: `backend/internal/store/memory.go`

Global in-memory `map[string]*session.Session` with `sync.RWMutex`:
- `Get(id)` — read lock.
- `Set(id, sess)` — write lock.
- `Reset()` — replaces entire store (used in tests).

---

## Console REPL

Files in `console/`:
- `main.go` — REPL loop, commands: `start`, `rules`, `quit`.
- `client.go` — HTTP client wrapping all API endpoints.
- `session.go` — session runner (front→ENTER→grade loop in terminal).
- `rules.go` — `rules get <N>` and `rules set <N> <grammar.md>` subcommands.

---

## Database Workflow

### Migrations (golang-migrate)

Files in `backend/internal/repository/migrations/`:
- `000001_init.up.sql` / `.down.sql` — creates `phases`, `levels`, `vocabulary`, `sentences` tables.
- `000002_seed.up.sql` / `.down.sql` — inserts Phase 1, Level 1, and 28 vocabulary entries.

Migrations run automatically on startup when `database_url` is set.

### Query generation (sqlc)

Files in `backend/internal/repository/queries/`:
- `phases.sql` — list, create, update, delete, max number.
- `levels.sql` — list in phase, get, create, update, delete, max number, levels up to N.
- `vocabulary.sql` — up to level N, for level, add, set (delete+insert).
- `sentences.sql` — insert batch.

Generated Go code in `backend/internal/repository/db/` (auto-generated, not hand-edited).

---

## Testing

- `cd backend && go test ./... -v -count=1` — all pass.
- CI: GitHub Actions runs tests on push/PR to `main`.
- Handler tests use the in-memory repository and `SetGenerateFunc` to avoid LLM calls.
- Session tests cover state machine, guards, direction modes, and boundary cases.

---

## Maintenance Notes

- If sentence generation, grading, or level behavior changes, update the "How Sencha Works" popup text in `frontend/app.js` (`renderHowItWorks`).
- The hardcoded fallback cards (`backend/internal/session/cards.go`) are used only when no `SentencePairs` are provided to `NewSession`.
