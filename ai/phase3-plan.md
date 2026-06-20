# Phase 3 — Levels, Phases, and Persistence

## Status: ✅ Complete (merged in PR #7)

> **Note:** Additional endpoints were implemented beyond the original plan:
> - `PATCH /api/phases/:number` — update phase name
> - `DELETE /api/phases/:number` — delete phase (cascades)
> - `GET /api/levels/max` — get highest level number
> - `DELETE /api/levels/:number` — delete level (cascades, renumbers following levels)
> - `PUT /api/levels/:number/vocabulary` — replace level vocabulary
>
> Items still deferred: user accounts, multi-language "journeys", sentence reuse strategy, grammar accumulation for 100+ levels.

---

## Overview

Introduce a structured curriculum (levels grouped into phases) backed by a database, replacing the current hardcoded grammar rules and vocabulary. The LLM generates sentences constrained by the accumulating rules of the user's current level.

---

## Terminology

- **Journey** — Top-level container (one per language). Not implemented this phase; only noted.
- **Phase** — Group of levels (analogous to a Unit). Enumerated starting at 1.
- **Level** — A single lesson's grammatical rules + vocabulary. Globally sequential across phases. Enumerated starting at 1.

Levels are **globally sequential** — they never reset per phase:

| Phase | Levels |
|-------|--------|
| 1     | 1, 2   |
| 2     | 3, 4, 5 |
| 3     | 6, 7, 8 |

To create a new level in a new phase, the phase number must be ≤ max phase + 1.
To create a new level, its number must be ≤ max level + 1.

---

## Database Tooling Stack

| Concern | Tool | Purpose |
|---------|------|---------|
| Schema migrations | [golang-migrate/migrate](https://github.com/golang-migrate/migrate) | Versioned `.up.sql` / `.down.sql` files, run on startup |
| Query generation | [sqlc](https://sqlc.dev) | Generate type-safe Go structs + query functions from raw SQL |
| Driver | [pgx](https://github.com/jackc/pgx) | PostgreSQL driver (sqlc generates pgx-compatible code by default) |

**Workflow:**

1. Write migration SQL in `internal/repository/migrations/`
2. Write query SQL in `internal/repository/queries/` — sqlc reads these and generates `internal/repository/db/` (models + query methods)
3. Repository implementation wraps the generated `*db.Queries` for PostgreSQL, or uses maps for the in-memory variant
4. On startup, if `database_url` is set, run migrations and create a `*db.Queries` from a `pgx` pool

**sqlc config** (`backend/sqlc.yaml`):

```yaml
version: "2"
sql:
  - engine: "postgresql"
    schema: "internal/repository/migrations/"
    queries: "internal/repository/queries/"
    gen:
      go:
        package: "db"
        out: "internal/repository/db"
        sql_package: "pgx/v5"
```

---

## Data Model

### Tables

```sql
CREATE TABLE phases (
    number INTEGER PRIMARY KEY,
    name   TEXT NOT NULL
);

CREATE TABLE levels (
    number        INTEGER PRIMARY KEY,  -- globally sequential: 1, 2, 3…
    phase_number  INTEGER NOT NULL REFERENCES phases(number),
    grammar_md    TEXT NOT NULL,
    exceptions_md TEXT                   -- nullable
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

---

## Repository Interface

Defined in `internal/repository/`:

```go
type Phase struct {
    Number int
    Name   string
}

type Level struct {
    Number       int
    PhaseNumber  int
    GrammarMD    string
    ExceptionsMD string
}

type VocabEntry struct {
    Korean  string
    English string
}

type LevelData struct {
    GrammarMD    string       // accumulated from levels 1..N
    Vocab        []VocabEntry // accumulated from levels 1..N
    ExceptionsMD string       // current level only (not cumulative)
}

type Repository interface {
    // Phases
    ListPhases() ([]Phase, error)
    CreatePhase(p Phase) error
    MaxPhaseNumber() (int, error)

    // Levels
    LevelsInPhase(phaseNumber int) ([]Level, error)
    Level(number int) (*Level, error)
    CreateLevel(l Level) error
    MaxLevelNumber() (int, error)
    LevelsUpTo(number int) ([]Level, error)

    // Vocabulary
    VocabularyUpTo(levelNumber int) ([]VocabEntry, error)
    AddVocabulary(entries []VocabEntry) error

    // Sentences
    SaveSentences(sentences []Sentence) error

    // Convenience
    LoadLevelData(levelNumber int) (*LevelData, error)
}
```

---

## Config Changes

Add a non-JSON field to `Config` for runtime dependency injection:

```go
type Config struct {
    LLM         LLMConfig    `json:"llm"`
    DatabaseURL string       `json:"database_url"`
    Repository  Repository   `json:"-"` // wired by main()
}
```

---

## API Endpoints

| Method | Path | Description |
|--------|------|-------------|
| `GET` | `/api/phases` | List all phases |
| `POST` | `/api/phases` | Create phase `{"number": N, "name": "..."}` |
| `GET` | `/api/phases/:number/levels` | List levels in phase |
| `POST` | `/api/levels` | Create level |
| `GET` | `/api/levels/:number` | Get level detail |

**POST /api/levels body:**

```json
{
  "phase_number": 2,
  "grammar_markdown": "## New rules\n...",
  "exceptions_markdown": "- Watch out for X\n- Watch out for Y",
  "vocabulary": [
    {"korean": "학교", "english": "school"}
  ]
}
```

The server auto-assigns the next sequential level number and validates constraints.

---

## Session creation becomes level-aware

**POST /api/sessions** now accepts an optional `level_number` (defaults to 1):

```json
{"level_number": 3, "direction": "korean-to-english"}
```

Handler flow:
1. Fetch `LevelData` from repository for the given level
2. Call `sengen.Generate(10, levelData)`
3. Save generated sentences to repository
4. Create session as before (ephemeral, in-memory)

---

## Sengen refactor

### Remove embedded globals

- Delete `grammar.md`
- Delete `vocab.go`
- Remove `//go:embed grammar.md`

### `Generate` now takes `LevelData`

```go
func Generate(count int, data LevelData) ([]session.SentencePair, error)
```

- `buildPrompt` receives grammar + vocab from `LevelData`
- `grammarCheck` receives `ExceptionsMD` and includes it in the prompt if non-empty

### Grammar checker prompt updated

```markdown
Verify and correct each Korean-English sentence pair:
1. Is the Korean sentence grammatically correct? Fix it if not.
2. Is the English translation correct for the (corrected) Korean? Fix it if not.
{{if .Exceptions}}
Pay special attention to these common mistakes:
{{.Exceptions}}
{{end}}
Return ONLY the corrected list, one pair per line, in this format:
"Korean sentence"
"English translation"
...
Do not add, remove, or reorder pairs.

Pairs:
{{.Pairs}}
```

---

## Execution order

| # | Step | Files |
|---|------|-------|
| 1 | Define interfaces + models in `internal/repository/` | `internal/repository/models.go`, `internal/repository/interface.go` |
| 2 | In-memory implementation (maps, seed with current content) | `internal/repository/memory.go`, `internal/repository/seed.go` |
| 3 | Refactor sengen — remove `grammar.md`, `vocab.go`, make `Generate` accept `LevelData`, update tests | `internal/sengen/sengen.go`, `internal/sengen/sengen_test.go` |
| 4 | Add exceptions to grammar checker — update template, pass through `GrammarCheck` | `internal/sengen/gramchecker.tmpl`, `internal/sengen/sengen.go` |
| 5 | New API handlers — phases + levels CRUD | `internal/handler/phases.go`, `internal/handler/levels.go`, `internal/handler/routes.go` |
| 6 | Update session creation — accept `level_number`, fetch level data, pass to `Generate`, save sentences | `internal/handler/sessions.go` |
| 7 | Wire repository in `main.go` — seed defaults, inject into config | `cmd/api/main.go`, `internal/config/config.go` |
| 8 | Update handler tests — mock repository | `internal/handler/handler_test.go` |
| 9 | Update integration tests | `cmd/api/main_test.go` |
| 10 | Write DB migrations | `internal/repository/migrations/*.sql` |
| 11 | Write sqlc queries | `internal/repository/queries/*.sql`, `backend/sqlc.yaml` |
| 12 | Generate sqlc code | `internal/repository/db/` (auto-generated) |
| 13 | PostgreSQL implementation of Repository | `internal/repository/postgres.go` |
| 14 | Final wiring — `database_url` drives backend selection | `cmd/api/main.go`, `internal/config/config.go` |

---

## Deferred to future phases

- Multiple journeys (language-agnostic support)
- Sentence reuse/regenerate strategy (pulling from `sentences` table vs fresh LLM call)
- Grammar accumulation strategy for 100+ levels (prompt size management)
- Phase/level editing (update, delete)
- User accounts

---

## Backward compatibility

- The server starts with a seeded in-memory repo containing Phase 1, Level 1, and all current vocabulary
- If no `database_url` is set, the server works identically to today
- Existing handler tests mock the repository
- Existing integration tests seed the in-memory repo explicitly
