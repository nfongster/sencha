# UI Phase 1 — Web Frontend SPA

## Status: ✅ Complete (merged in PR #9)

> **Differences from this plan in the final implementation:**
> - Home page has 3 buttons (Start, Journey, How it Works) instead of 2 (Start, Rules).
> - 7 hash-routed views instead of 5: added `#level-select`, `#level-picker`, `#how-it-works`.
> - Rules page has full CRUD (edit grammar/exceptions/vocabulary, delete phases/levels) rather than read-only browse.
> - 16 API endpoints consumed instead of the 10 originally listed.
> - Session expiry 404 handling was not implemented; generic error banner covers all cases.

## Overview

Replace the console REPL with a browser-based single-page application (SPA) served by the existing Gin backend. All existing API endpoints, session logic, and repository layer remain unchanged. Zero new build tooling — vanilla HTML, CSS, and JavaScript.

## New directory structure

```
sencha/
├── frontend/
│   ├── index.html      # Single HTML shell — all views rendered in JS
│   ├── style.css       # All styles
│   └── app.js          # Application logic (routing, state, API, rendering)
├── backend/            # unchanged
├── console/            # unchanged
└── ai/
    └── ui-phase1-plan.md
```

## Backend changes (minimal)

**`backend/cmd/api/main.go`** — Add static file serving:

```go
r.Static("/static", "./frontend")
r.GET("/", func(c *gin.Context) {
    c.File("./frontend/index.html")
})
```

The API routes remain under `/api/*` with no changes.

No changes to handlers, repository, session logic, config, or tests.

## Frontend architecture

### Routing (hash-based)

| Hash | View | Description |
|------|------|-------------|
| `#home` | Home | Two buttons: Start, Rules |
| `#setup` | Session setup | Direction picker (KR→EN, EN→KR, Mixed) |
| `#session` | Active session | Progress graph + card + grading |
| `#summary` | Session summary | Stats + start new / quit |
| `#rules` | Rules browser | Phase/level tree + detail pop-ups |

### Session state (local)

```js
session = {
  id: "session-N",
  direction: "korean-to-english",
  totalCards: 10,
  cardsRemaining: 10,
  complete: false,
  cardStates: [], // per-card: "unrevealed" | "revealed" | "pass" | "hard" | "fail"
  gradeCounts: { pass: 0, hard: 0, fail: 0 }
}
```

### Page-by-page detail

#### Home (`#home`)
- Full black background (`#000`)
- "Sencha" title in white, centered
- Two buttons centered below:
  - **[1] Start** — navigates to `#setup`
  - **[2] Rules** — navigates to `#rules`
- Keyboard: `1` → Start, `2` → Rules

#### Session setup (`#setup`)
- Title: "Select Direction"
- Three radio-style buttons:
  - **[1] Korean → English**
  - **[2] English → Korean**
  - **[3] Mixed**
- "Start Session" button
- Keyboard: `1`/`2`/`3` to select, ENTER to confirm

#### Active session (`#session`)

**Progress graph (SVG):**
- Horizontal line of 10 connected circles (one per card)
- Colors: green (`#4ade80`) = pass, gold (`#fbbf24`) = hard, red (`#f87171`) = fail, gray (`#4b5563`) = unrevealed
- Current unrevealed card: white outline ring
- Connector lines between circles

**Statistics bar (below progress graph):**
- "Completed: A / B"
- "Pass: X" (green), "Hard: Y" (gold), "Fail: Z" (red)

**Card area (center):**
- Front text: large, centered, white
- Below front: horizontal divider

**Before reveal:**
- **[Reveal]** button — green, large
- Keyboard: ENTER or SPACE

**After reveal:**
- Front text stays in place
- Back text rendered below divider
- **[1] Pass** (green), **[2] Hard** (gold), **[3] Fail** (red) replace Reveal button
- Keyboard: `1` → Pass, `2` → Hard, `3` → Fail
- After grading: card state updates in progress graph, next card appears immediately

**Session complete:**
- If `session_complete: true` from grade response → auto-navigate to `#summary`

#### Session summary (`#summary`)
- "Session Complete!" header
- Final stats:
  - Pass: X (green)
  - Hard: Y (gold)
  - Fail: Z (red)
  - Percentage: X/10 = XX%
- Two buttons:
  - **[S] Start New** → `#setup`
  - **[Q] Quit** → `#home`

#### Rules browser (`#rules`)
- **Left sidebar:**
  - Vertical list of phase nodes (circles with phase name)
  - Each phase node expands horizontally to show its level nodes (connected dots)
  - Phase node colors: white text on dark gray
  - Level node colors: white text, numbered
- **Clicking a level node:**
  - Opens a centered modal pop-up
  - **Left column:** Grammar markdown rendered as HTML (via CDN-loaded marked.js)
  - **Right column:** Vocabulary list (Korean — English)
  - Close button (X) or press ESC
- **"Add Phase" button** — at bottom of phase list
  - Pop-up opens with inputs: Phase number (auto-filled), Phase name
  - Calls `POST /api/phases`
- **"Add Level" button** — at bottom of level list
  - Pop-up opens with:
    - Phase number (dropdown of existing phases)
    - Grammar rules textarea (write markdown directly — no file picker)
    - Exceptions textarea (optional)
    - Vocabulary section: inputs for Korean + English word pairs, "Add another" button
  - Calls `POST /api/levels`
- **Back button** — returns to `#home`

### Error / edge-case handling
- **API errors** — Red banner at top, auto-dismiss after 5 seconds. Shows error message + code.
- **Loading states** — "Loading..." overlay during API calls.
- **Session expired** — If session API returns 404, redirect to `#home` with "Session expired" message.
- **Empty states** — "No phases yet" / "No vocabulary for this level" messages.
- **Markdown rendering** — Use `marked.js` loaded from CDN. Sanitize via DOMPurify or marked's built-in escaping.

## CDN dependencies
- `marked.js` — Markdown to HTML rendering (for grammar rules display)
- Loaded via `<script>` tags in `index.html`

## No API changes required
The frontend consumes these existing endpoints:

| Method | Path | Used by |
|--------|------|---------|
| `GET` | `/api/health` | (optional, for startup check) |
| `POST` | `/api/sessions` | Session setup |
| `GET` | `/api/sessions/:id` | Session status |
| `POST` | `/api/sessions/:id/reveal` | Card reveal |
| `POST` | `/api/sessions/:id/grade` | Card grading |
| `GET` | `/api/phases` | Rules page |
| `POST` | `/api/phases` | Add phase |
| `GET` | `/api/phases/:number/levels` | Rules page |
| `GET` | `/api/levels/:number` | Level detail (grammar + vocab) |
| `POST` | `/api/levels` | Add level |
| `PATCH` | `/api/levels/:number` | (future use) |

## Out of scope
- React or any JS framework
- Build tooling / bundlers
- TypeScript
- Mobile responsiveness
- Animations beyond simple transitions
- Persistent state across page reloads
- User accounts
