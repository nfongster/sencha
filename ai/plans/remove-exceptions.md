# Remove Grammatical Exceptions

Remove the `exceptions_md` concept from the entire stack — database, backend Go code, frontend JS, and console CLI.

---

## Problem

The `exceptions_md` field on levels is unnecessary complexity. It stores grammar exceptions that the sentence grammar checker uses to watch for common mistakes. But a small amount of difficulty leakage between levels is acceptable, making this feature not worth maintaining.

---

## Solution

Remove `exceptions_md` from every layer:

### 1. Database migration

Create `000004_remove_exceptions.up.sql`:
```sql
ALTER TABLE levels DROP COLUMN exceptions_md;
```

Create `000004_remove_exceptions.down.sql`:
```sql
ALTER TABLE levels ADD COLUMN exceptions_md TEXT;
```

### 2. SQL queries

**File:** `backend/internal/repository/queries/levels.sql`

| Query | Change |
|-------|--------|
| `LevelsInPhase` | `SELECT number, phase_number, grammar_md` (remove COALESCE + alias) |
| `GetLevel` | Same |
| `CreateLevel` | `INSERT INTO levels (number, phase_number, grammar_md) VALUES ($1, $2, $3)` (remove 4th param) |
| `UpdateLevel` | `UPDATE levels SET grammar_md = $2 WHERE number = $1` (remove `exceptions_md = $3`) |
| `LevelsUpTo` | Same as LevelsInPhase |

Regenerate with `sqlc generate` (or hand-update `db/levels.sql.go` and `db/models.go`).

### 3. Go models

**File:** `backend/internal/repository/models.go`

Remove `ExceptionsMD string` from both `Level` and `LevelData` structs.

**File:** `backend/internal/repository/db/models.go`

Remove `ExceptionsMd pgtype.Text` from generated model.

### 4. Repository interface

**File:** `backend/internal/repository/interface.go`

```go
UpdateLevel(number int, grammarMD string) error
```

### 5. Postgres repository

**File:** `backend/internal/repository/postgres.go`

- Remove all `ExceptionsMD` mapping from query scan results (lines that reference `row.ExceptionsMd`)
- Simplify `CreateLevel` / `UpdateLevel` — no exceptions param, no `pgtype.Text` wrapping
- Remove `var exceptions pgtype.Text` and related logic

### 6. Memory repository

**File:** `backend/internal/repository/memory.go`

- Remove `ExceptionsMD` from `UpdateLevel` param and body
- Remove `ExceptionsMD: l.ExceptionsMD` from `LoadLevelData`

### 7. Seed data

**File:** `backend/internal/repository/seed.go`

Remove `ExceptionsMD: ""` from the `CreateLevel` call.

### 8. Sentence generator

**File:** `backend/internal/sengen/sengen.go`

- Remove `Exceptions` field from `gramCheckData` struct
- Change `grammarCheck(pairs []session.SentencePair, exceptions string)` → `grammarCheck(pairs []session.SentencePair)`
- Remove exceptions handling in `grammarCheck` body
- Remove exceptions from `buildGrammarCheckPrompt` signature, template data, and template execution
- Update `Generate()` call to `grammarCheck(pairs)` (no exceptions arg)
- Remove `data.ExceptionsMD` usage

**File:** `backend/internal/sengen/gramchecker.tmpl`

Remove the entire `{{if .Exceptions}}...{{end}}` block.

### 9. Handler

**File:** `backend/internal/handler/levels.go`

- Remove `Exceptions` from `createLevelRequest` struct
- Remove `ExceptionsMD: req.Exceptions` from CreateLevel body
- Remove `Exceptions` from `updateLevelRulesRequest` struct
- Change `UpdateLevelRulesHandler`: accept only `grammar_markdown`. Remove the `if req.GrammarMD == "" && req.Exceptions == ""` validation (grammar_md is now required and sufficient)

### 10. Frontend

**File:** `frontend/app.js`

- `API.updateLevel(number, grammarMD)` — drop `exceptionsMD` param
- `showAddLevelModal()` — remove exceptions textarea and its label
- `showEditLevelForm()` — remove exceptions textarea and its label
- `showLevelDetail()` — remove `if (level.exceptions_md) { grammarHtml += '<h3>Exceptions</h3>' + ... }` blocks (two occurrences: lines ~305 and ~636)
- `submitAddLevel()` — remove `exceptions` from form gathering and API call
- `submitEditLevel()` — remove `exceptions` from form gathering and API call

### 11. Console

**File:** `console/client.go`

- Remove `ExceptionsMD` and `Exceptions` fields from response/request structs
- Change `UpdateLevel(number int, grammarMD string)` — drop `exceptions` param

**File:** `console/rules.go`

- Remove exceptions from get display (`fmt.Println("Exceptions:") / if l.ExceptionsMD == "" / else blocks`)
- Remove exceptions from set subcommand (no more optional exceptions file arg)

### 12. Tests

**File:** `backend/internal/sengen/sengen_test.go`

Remove `TestBuildGrammarCheckPrompt_IncludesExceptions`.

Remove `exceptions` param from any remaining `buildGrammarCheckPrompt` calls in tests.

---

## Files touched

| File | Change |
|------|--------|
| `backend/internal/repository/migrations/000004_remove_exceptions.up.sql` | New |
| `backend/internal/repository/migrations/000004_remove_exceptions.down.sql` | New |
| `backend/internal/repository/queries/levels.sql` | Remove exceptions_md column refs |
| `backend/internal/repository/db/levels.sql.go` | Regenerate/update |
| `backend/internal/repository/db/models.go` | Remove ExceptionsMd |
| `backend/internal/repository/models.go` | Remove ExceptionsMD from Level, LevelData |
| `backend/internal/repository/interface.go` | UpdateLevel sig change |
| `backend/internal/repository/postgres.go` | Remove exceptions handling |
| `backend/internal/repository/memory.go` | Remove exceptions handling |
| `backend/internal/repository/seed.go` | Remove ExceptionsMD |
| `backend/internal/sengen/sengen.go` | Remove exceptions from gramCheckData, grammarCheck, buildGrammarCheckPrompt, Generate |
| `backend/internal/sengen/gramchecker.tmpl` | Remove {{if .Exceptions}} block |
| `backend/internal/sengen/sengen_test.go` | Remove exceptions test |
| `backend/internal/handler/levels.go` | Remove exceptions from request structs, validation, create/update calls |
| `frontend/app.js` | Remove exceptions textarea, rendering, form fields |
| `console/client.go` | Remove exceptions fields |
| `console/rules.go` | Remove exceptions subcommands |

---

## Testing

1. Run all backend tests: `go test ./backend/...`
2. Start server, create/update levels — verify grammar_md alone works
3. Start a session — verify generated sentences still pass grammar check
4. Level detail modal has no "Exceptions" heading
5. Add Level modal has no exceptions textarea
6. Edit Level form has no exceptions textarea
