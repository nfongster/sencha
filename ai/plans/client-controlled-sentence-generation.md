# Client-Controlled Sentence Generation

Replace the always-generate session creation with a DB-first sentence retrieval strategy, and expose sentence management in the UI (count, generate, clear, view).

---

## Problem

1. Every session makes an LLM call to generate 10 sentences, even if usable sentences already exist in the DB. This wastes tokens and adds latency.
2. The user has no visibility into what sentences exist for a level and no way to manage them (batch generate, clear, review).

---

## Solution

### 1. Add sqlc queries

**File:** `backend/internal/repository/queries/sentences.sql`

```sql
-- name: ListSentencesByLevel :many
SELECT level_number, korean, english FROM sentences WHERE level_number = $1;

-- name: CountSentencesByLevel :one
SELECT COUNT(*) FROM sentences WHERE level_number = $1;

-- name: DeleteSentencesByLevel :exec
DELETE FROM sentences WHERE level_number = $1;

-- name: GetRandomSentencesByLevel :many
SELECT level_number, korean, english FROM sentences WHERE level_number = $1 ORDER BY RANDOM() LIMIT $2;
```

**Action:** Regenerate sqlc code (`cd backend && sqlc generate`).

### 2. Update Repository interface

**File:** `backend/internal/repository/interface.go`

Add to the `Repository` interface:

```go
SentencesForLevel(levelNumber int) ([]Sentence, error)
CountSentencesForLevel(levelNumber int) (int, error)
DeleteSentencesForLevel(levelNumber int) error
RandomSentencesForLevel(levelNumber int, count int) ([]Sentence, error)
```

### 3. Implement in both repos

**File:** `backend/internal/repository/postgres.go`

```go
func (r *PostgresRepository) SentencesForLevel(levelNumber int) ([]Sentence, error) {
    rows, err := r.queries.ListSentencesByLevel(r.ctx, int32(levelNumber))
    // ... map db.Sentence → repository.Sentence
}

func (r *PostgresRepository) CountSentencesForLevel(levelNumber int) (int, error) {
    count, err := r.queries.CountSentencesByLevel(r.ctx, int32(levelNumber))
    return int(count), err
}

func (r *PostgresRepository) DeleteSentencesForLevel(levelNumber int) error {
    return r.queries.DeleteSentencesByLevel(r.ctx, int32(levelNumber))
}

func (r *PostgresRepository) RandomSentencesForLevel(levelNumber int, count int) ([]Sentence, error) {
    rows, err := r.queries.GetRandomSentencesByLevel(r.ctx, db.GetRandomSentencesByLevelParams{
        LevelNumber: int32(levelNumber),
        Limit:       int32(count),
    })
    // ... map
}
```

**File:** `backend/internal/repository/memory.go`

Add an in-memory `sentencesByLevel` map. Implement the four new methods using `math/rand` shuffle + slice for `RandomSentencesForLevel`.

### 4. New API endpoints

**File:** `backend/internal/handler/routes.go` + new handler file

| Method | Path | Handler | Purpose |
|--------|------|---------|---------|
| `GET` | `/api/levels/:number/sentences/count` | `CountSentencesHandler` | Returns `{"count": N}` |
| `GET` | `/api/levels/:number/sentences` | `ListSentencesHandler` | Returns `{"sentences": [...]}` |
| `DELETE` | `/api/levels/:number/sentences` | `DeleteSentencesHandler` | Deletes all sentences for the level |
| `POST` | `/api/levels/:number/sentences/generate` | `GenerateSentencesHandler` | Body: `{"count": N}` — generates N sentences via LLM, saves them, returns them |

**GenerateSentencesHandler** — receives a count (1–100), calls `sengen.Generate(count, levelData)`, saves via `SaveSentences()`, returns the generated sentences.

**Sentence JSON shape:**
```json
{"korean": "...", "english": "..."}
```

### 5. Change session creation to pull from DB first

**File:** `backend/internal/handler/sessions.go` — `CreateSessionHandler`

Replace the direct `sengen.Generate(10, levelData)` call:

```go
func CreateSessionHandler(c *gin.Context) {
    // ... existing validation ...

    // 1. Try to get 10 random sentences from DB
    sentences, err := appConfig.Repository.RandomSentencesForLevel(levelNum, 10)
    if err != nil {
        // fall back to generation
    }

    if len(sentences) < 10 {
        // 2. Generate the shortfall
        needed := 10 - len(sentences)
        pairs, err := sengen.Generate(needed, levelData)
        if err != nil {
            c.JSON(http.StatusServiceUnavailable, ...)
            return
        }
        // Convert to Sentence and save
        newSentences := sessionsToSentences(pairs, levelNum)
        if err := appConfig.Repository.SaveSentences(newSentences); err != nil {
            log.Printf("[handler] failed to save sentences: %v", err)
        }
        // Append to existing sentences
        for _, s := range newSentences {
            sentences = append(sentences, s)
        }
    }

    // 3. Convert Sentence → SentencePair and create session
    pairs := sentencesToPairs(sentences)
    var sess *session.Session
    // ... rest of existing logic ...
}
```

**Important edge cases:**
- If 10 sentences exist → no LLM call at all (fast path)
- If 5 exist → generate 5 more (hybrid)
- If 0 exist → generate all 10 (current behavior)

### 6. Frontend: Show sentence count on Journey page

**File:** `frontend/app.js` — `renderRules()` / `renderLevelPicker()`

When rendering level dots/nodes, fetch the sentence count for each level and display it beneath the level number:

```js
// After loading levels, fetch counts in parallel
const countPromises = levels.map(l =>
    API.countSentences(l.number).then(d => d.count).catch(() => 0)
);
const counts = await Promise.all(countPromises);
```

Display in the level dot:
```html
<div class="level-node" data-level="${level.number}">
    ${level.number}
    <span class="sentence-count">${count} sentences</span>
</div>
```

### 7. Frontend: Sentence management buttons in level modal

**File:** `frontend/app.js` — `showLevelDetail()`

Add three buttons to the modal's `form-actions` div:

```
[Generate Sentences] [View Sentences] [Clear Sentences] [Close]
```

- **Generate Sentences** → calls `showGenerateSentencesModal(levelNumber)`. Modal has a number input (min=1, max=100) and Confirm/Cancel. On confirm → `POST /api/levels/:number/sentences/generate` with the count. On success, show a success message and refresh the count display.
- **Clear Sentences** → confirm dialog ("Delete all N sentences for this level?"). On confirm → `DELETE /api/levels/:number/sentences`. Refresh the count to 0.
- **View Sentences** → `showViewSentencesModal(levelNumber)`. Fetches `GET /api/levels/:number/sentences` and displays in a scrollable table (Korean | English). Close button to dismiss.

### 8. Frontend: API client additions

**File:** `frontend/app.js`

```js
countSentences(levelNumber)    // GET /api/levels/:number/sentences/count
listSentences(levelNumber)     // GET /api/levels/:number/sentences
deleteSentences(levelNumber)   // DELETE /api/levels/:number/sentences
generateSentences(levelNumber, count)  // POST /api/levels/:number/sentences/generate
```

### 9. Frontend: Styling

**File:** `frontend/style.css`

- `.sentence-count` — small text under level dot (font-size 10px, color #6b7280)
- `.sentences-table` — table styling for the View Sentences modal
- `.count-input` — number input styling for the Generate modal

---

## Files touched

| File | Change |
|------|--------|
| `backend/internal/repository/queries/sentences.sql` | Add 4 new queries |
| `backend/internal/repository/db/sentences.sql.go` | Regenerated by sqlc |
| `backend/internal/repository/db/models.go` | May be regenerated |
| `backend/internal/repository/interface.go` | Add 4 new sentence methods |
| `backend/internal/repository/postgres.go` | Implement 4 new methods + fix `SaveSentences` if needed |
| `backend/internal/repository/memory.go` | Add 4 new methods + in-memory sentence store |
| `backend/internal/handler/sessions.go` | Rewrite `CreateSessionHandler` to pull from DB first |
| `backend/internal/handler/sentences.go` | New file: count, list, delete, generate handlers |
| `backend/internal/handler/routes.go` | Register 4 new sentence routes |
| `frontend/app.js` | API client methods; update `renderRules`/`showLevelDetail`; new modals for generate/view/clear |
| `frontend/style.css` | Sentence count, sentences table, count input styles |

---

## Testing

1. **Backend:** `GET /api/levels/1/sentences/count` returns 0 initially.
2. **Backend:** `POST /api/levels/1/sentences/generate` with `{"count": 5}` saves and returns 5 sentences.
3. **Backend:** `GET /api/levels/1/sentences/count` now returns 5.
4. **Backend:** `GET /api/levels/1/sentences` returns the 5 sentences.
5. **Backend:** Start a session for level 1 → no LLM call (sentences pulled from DB).
6. **Backend:** `DELETE /api/levels/1/sentences` clears them; `GET /api/levels/1/sentences/count` returns 0.
7. **Backend:** Start a session for level 1 with 0 sentences → generates all 10 (fallback).
8. **Frontend:** Journey page shows "5 sentences" under each level dot.
9. **Frontend:** Level modal has 3 new buttons.
10. **Frontend:** Generate modal creates N sentences; count updates.
11. **Frontend:** View modal shows all sentences.
12. **Frontend:** Clear button deletes sentences; count drops to 0.
13. **Edge case:** Generate with count < 1 or > 100 → 400 error.
14. **Edge case:** Start session with exactly 10 sentences → no LLM call (fast path).
15. **Edge case:** Start session with 3 existing sentences → generates 7 more, saves them.
