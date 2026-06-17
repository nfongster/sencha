# Phase 1 — Sencha: Console App + Go Server

## Status: ✅ Complete

All 30 tests pass. All 15 manual smoke tests passed.

## Project Structure

```
sencha/
├── ai/
│   └── phase1-plan.md
├── backend/
│   ├── cmd/
│   │   └── api/
│   │       └── main.go             # Entry point
│   ├── internal/
│   │   ├── handler/
│   │   │   ├── health.go
│   │   │   ├── routes.go
│   │   │   ├── sessions.go
│   │   │   └── handler_test.go     # 15 tests
│   │   ├── session/
│   │   │   ├── session.go
│   │   │   ├── cards.go
│   │   │   └── session_test.go     # 15 tests
│   │   └── store/
│   │       └── memory.go           # In-memory session store
│   ├── go.mod
│   └── go.sum
├── console/
│   ├── main.go                     # Entry point + REPL loop
│   ├── client.go                   # HTTP client for backend API
│   ├── session.go                  # Session runner logic
│   ├── go.mod
│   └── .gitignore
├── .gitignore
└── README.md
```

## What's Implemented

### REST API
| Method | Endpoint | Purpose |
|--------|----------|---------|
| `GET` | `/api/health` | Health check |
| `POST` | `/api/sessions` | Create session (optional `direction` param) |
| `GET` | `/api/sessions/:id` | Get session status |
| `POST` | `/api/sessions/:id/reveal` | Reveal current card |
| `POST` | `/api/sessions/:id/grade` | Grade card (`pass`/`hard`/`fail`) |

### Session Logic
- New session shuffles all 10 cards randomly
- Direction modes: `korean-to-english` (default), `english-to-korean`, `mixed`
- Guards: grade-before-reveal (409), stale actions after completion (409), invalid grade (400)
- Session tracks grade counts, returns summary on completion

### Console REPL
- Main menu: `start` / `quit`
- `start` prompts for direction, creates session via API
- Session loop: show front → ENTER to reveal → grade `p`/`h`/`f`
- Post-session summary with option to start new or quit

## Developer Notes

- **TDD approach used** throughout: write test → red → implement → green
- **To run tests:** `cd backend && go test ./... -v -count=1`
- **To run server:** `cd backend && go run ./cmd/api/`
- **To run console:** `cd console && go run .`

## Next Steps (Paused)

### Sub-Phase A: CI/CD
GitHub Actions workflow triggering unit tests on push/PR to `main`.

### Sub-Phase B: Docker
Containerize the backend application.

