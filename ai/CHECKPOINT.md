# Checkpoint — `main`

**Status:** All major phases merged.

## Merged Work

| # | PR | What |
|---|----|------|
| 1 | — | Phase 1: Console REPL + Go server with hardcoded 10-sentence sessions |
| 2 | #3 | Phase 2: LLM sentence generation via sengen (OpenAI-compatible API) |
| 3 | #5/#6 | Grammar check phase after generation (two-pass: generate then validate) |
| 4 | #4 | Config cleanup: remove defaults, hard-stop on missing/invalid config |
| 5 | #7 | Phase 3: Structured curriculum (levels/phases) with PostgreSQL persistence + sqlc + migrations + in-memory fallback. Repository pattern with two backends. |
| 6 | #8 | Console `rules get`/`rules set` subcommands, `PATCH /api/levels/:number` |
| 7 | #9 | UI Phase 1: Vanilla JS SPA (hash-routing, session flow, rules browser, add phase/level) |
| 8 | #10 | Edit vocabulary: `PUT /api/levels/:number/vocabulary` + in-modal vocab editor in frontend |
| 9 | #11 | Level grammar-only mode, cleanup of old embedded globals |

## Current API Surface

| Method | Route | Purpose |
|--------|-------|---------|
| `GET` | `/api/health` | Health check |
| `GET` | `/api/phases` | List all phases |
| `POST` | `/api/phases` | Create a phase |
| `PATCH` | `/api/phases/:number` | Update phase name |
| `DELETE` | `/api/phases/:number` | Delete phase (cascades) |
| `GET` | `/api/phases/:number/levels` | List levels in a phase |
| `POST` | `/api/levels` | Create level (auto-assigns sequential number) |
| `GET` | `/api/levels/max` | Get highest level number |
| `GET` | `/api/levels/:number` | Get level detail (grammar, exceptions, vocab) |
| `PATCH` | `/api/levels/:number` | Update grammar/exceptions |
| `PUT` | `/api/levels/:number/vocabulary` | Replace level vocabulary |
| `DELETE` | `/api/levels/:number` | Delete level (cascades, renumbers) |
| `POST` | `/api/sessions` | Create session (accepts `direction` + `level_number`) |
| `GET` | `/api/sessions/:id` | Get session status |
| `POST` | `/api/sessions/:id/reveal` | Reveal current card |
| `POST` | `/api/sessions/:id/grade` | Grade card (`pass`/`hard`/`fail`) |

## Test Status

- `cd backend && go test ./... -v -count=1` — all pass
- CI: GitHub Actions runs tests on push/PR to `main`
- No frontend tests (vanilla SPA, no framework)

## Architecture

```
console/  →  backend API (Gin)  →  Repository interface
                                    ├── in-memory (maps, seeded)
                                    └── PostgreSQL (pgx + sqlc + migrate)

frontend/  (vanilla JS SPA, hash-routed)
```

- Sessions are ephemeral (in-memory store), not persisted.
- Sentence generation is LLM-based with a hardcoded fallback.
- Config loaded from `config.json`; `database_url` selects repo backend.

## Known maintenance notes

- If sentence generation, grading, or level behavior changes, the "How Sencha Works" popup text in `frontend/app.js` (`renderHowItWorks`) must be updated.
- The hardcoded fallback cards (`backend/internal/session/cards.go`) are used only when no `SentencePairs` are provided to `NewSession`.
