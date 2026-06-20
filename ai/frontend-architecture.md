# Frontend Architecture

Stack: Vanilla JS SPA, HTML, CSS. No build tools, no frameworks, no TypeScript.

Served by the Gin backend: `r.Static("/static", "../frontend")` and `r.GET("/", c.File("../frontend/index.html"))`.

---

## File Layout

```
frontend/
├── index.html    # Shell: containers for views, CDN scripts, stylesheet link
├── style.css     # All styles (dark theme, black background)
└── app.js        # Application logic: routing, state, API client, views, keyboard (1067 lines)
```

**CDN dependency:** `marked.js` (loaded via `<script>` in `index.html`) for rendering grammar markdown as HTML.

---

## Hash Routing

8 hash-routed views — the switch in `router()` maps hash to render function:

| Hash | View | Render Function | Description |
|------|------|-----------------|-------------|
| `#home` | Home | `renderHome()` | Title + 3 navigation buttons |
| `#level-select` | Level Select | `renderLevelSelect()` | Start with latest level or pick from tree |
| `#level-picker` | Level Picker | `renderLevelPicker()` | Full phase/level tree, pick a level to start |
| `#setup` | Session Setup | `renderSetup()` | Direction picker (KR→EN, EN→KR, Mixed) |
| `#session` | Active Session | `renderSession()` | Progress graph, card, reveal/grade buttons |
| `#summary` | Session Summary | `renderSummary()` | Stats (Pass/Hard/Fail + percentage) |
| `#rules` | Rules Browser | `renderRules()` | Phase/level tree with full CRUD |
| `#how-it-works` | How It Works | `renderHowItWorks()` | Static info page about generation/grading/shortcuts |

Default route (unknown hash) → `#home`.

---

## App State

```js
const App = {
  sessionId: null,
  direction: null,
  totalCards: 0,
  cardsRemaining: 0,
  currentIndex: 0,
  cardStates: [],        // per-card color: #4b5563 | #6b7280 | #4ade80 | #fbbf24 | #f87171
  gradeCounts: { pass: 0, hard: 0, fail: 0 },
  sessionComplete: false,
  currentCard: null,     // { card_id, front, back }
  backRevealed: false,
  rulesData: null,       // { phases, levelsByPhase }
  selectedLevel: null,
};
```

State is **not persisted** across page reloads. Navigation to `#session` without a `sessionId` redirects to `#home`.

---

## API Client

```js
const API = {
  base: '',
  async request(method, path, body) { ... },
  get(path), post(path, body), patch(path, body), put(path, body), del(path),
  // Typed wrappers:
  createSession(direction, levelNumber),
  maxLevelNumber(),
  revealCard(sessionId),
  gradeCard(sessionId, grade),
  listPhases(),
  createPhase(number, name),
  updatePhase(number, name),
  deletePhase(number),
  deleteLevel(number),
  levelsInPhase(phaseNumber),
  getLevel(levelNumber),
  createLevel(data),
  updateLevel(number, grammarMD, exceptionsMD),
  setVocabulary(number, entries),
};
```

All errors thrown as `Error(message + " (" + code + ")")`.

---

## Views Detail

### Home (`#home`)

```
[Sencha title]
[1] Start         → #level-select
[2] Journey       → #rules
[3] How it Works  → #how-it-works
```

Keyboard: `1`/`2`/`3` navigate to respective views.

### Level Select (`#level-select`)

```
[Select Level]
[1] Latest Level (N)  → #setup with selectedLevel = max
[2] Choose a level    → #level-picker
← Back
```

Fetches `GET /api/levels/max` on render. Falls back to level 1 if fetch fails.
`startWithLatest()` sets `App.selectedLevel` and navigates to `#setup`.

### Level Picker (`#level-picker`)

Fetches all phases and their levels. Renders a visual tree:

```
Phase 1 ●
  ○ 1  ○ 2
Phase 2 ●
  ○ 3
```

Clicking a level node opens a **picker modal** showing:
- Grammar (rendered markdown) + Exceptions
- Vocabulary list (Korean — English)
- **[Start]** button (sets `App.selectedLevel` and navigates to `#setup`)
- **[Cancel]** closes modal

### Session Setup (`#setup`)

```
← Home
Select Direction
[1] Korean → English    (pre-selected)
[2] English → Korean
[3] Mixed
[Start Session]
```

Keyboard: `1`/`2`/`3` to select, `ENTER` to start.

`startSession()`:
1. Calls `API.createSession(direction, selectedLevel)`.
2. Initializes `App` state (sessionId, totalCards, cardStates, etc.).
3. Calls `revealCard()` to pre-fetch first card.
4. Navigates to `#session`.

### Active Session (`#session`)

Layout:
```
[Progress Graph SVG]    ← 10 connected circles
[Stats Bar]             ← Completed: 0/10 | Pass: 0 Hard: 0 Fail: 0
[Card Front]            ← Korean text (large, white, centered)
[Divider]
[Card Back]             ← Hidden until revealed
[Actions]               ← Reveal button OR grade buttons
```

**Progress Graph SVG** (`buildProgressGraph`):
- Horizontal line of 10 connected circles (600px viewBox, responsive width).
- Circle colors:
  - `#4b5563` (gray) = unrevealed
  - `#6b7280` (lighter gray) = current/revealed (not yet graded)
  - `#4ade80` (green) = pass
  - `#fbbf24` (gold) = hard
  - `#f87171` (red) = fail
- Current unrevealed card: white outline ring.
- Connector lines: `#374151`.

**Before reveal:** `[Reveal]` button (green). Keyboard: `ENTER` or `SPACE`.

**After reveal:** Front stays, back appears below divider. Three grade buttons replace Reveal button:
- `[1] Pass` (green, `#4ade80`)
- `[2] Hard` (gold, `#fbbf24`)
- `[3] Fail` (red, `#f87171`)

Keyboard: `1`/`2`/`3` for grading.

**After grading:** Card state updates in progress graph. Next card revealed immediately via `revealCard()` API call. If `session_complete: true` → auto-navigate to `#summary`.

### Session Summary (`#summary`)

```
Session Complete!
[Pass: X]  [Hard: Y]  [Fail: Z]
[XX%]
[S] Start New   → #setup
[Q] Quit        → #home
```

Keyboard: `S` → setup, `Q` → home.

### Rules Browser (`#rules`)

Two-column layout:

```
← Home
Phase 1 ●  [✏] [🗑]        │ Select a level to view rules
  ○ 1  ○ 2                 │
Phase 2 ●  [✏] [🗑]        │
  ○ 3                      │
[+ Add Phase]              │
[+ Add Level]              │
```

Click a level node → **level detail modal**:
- Grammar (rendered via marked.js) + Exceptions
- Vocabulary list
- **[Edit Rules]** button → inline form (grammar textarea + exceptions textarea + Save/Cancel)
- **[Edit Vocab]** button → inline form (dynamic rows: Korean + English inputs, add/remove, Save/Cancel)
- **[Delete Level]** button → confirmation modal with renumbering warning

**Phase edit/delete:** Each phase has edit (✏) and delete (🗑) buttons inline:
- Edit → modal with name input → calls `PATCH /api/phases/:number`.
- Delete → confirmation modal that warns about cascading level deletion.

**Add Phase** modal: number (auto-filled next sequential) + name → `POST /api/phases`.

**Add Level** modal: phase dropdown + grammar textarea + exceptions textarea + dynamic vocab rows → `POST /api/levels`.

### How It Works (`#how-it-works`)

Static info page describing:
- Phases & Levels curriculum structure
- Sentence generation (grammar + 50 random vocab words)
- Grading system (Pass/Hard/Fail)
- Keyboard shortcuts guide

Must be kept in sync if generation, grading, or level behavior changes.

---

## Modal System

Shared modal overlay with:
- `showModal(html)` — creates overlay, appends to body, wires close button + background click.
- `closeModal()` — removes overlay.
- Global `Escape` key listener closes active modal.

Modal pattern:
```html
<div class="modal">
  <button class="modal-close">&times;</button>
  <div class="modal-title">...</div>
  <div class="modal-body">...</div>
  <div class="modal-form">...inputs...</div>
  <div class="form-actions">
    <button class="btn btn-sm" onclick="closeModal()">Cancel</button>
    <button class="btn btn-sm btn-green" onclick="submitX()">Save</button>
  </div>
</div>
```

---

## Keyboard Handling

Single global `keydown` listener (at `document` level) dispatches per view:

| View | Keys | Action |
|------|------|--------|
| `#home` | `1`, `2`, `3` | Navigate to Start/Journey/How It Works |
| `#level-select` | `1`, `2` | Latest level / Choose a level |
| `#setup` | `1`, `2`, `3`, `ENTER` | Select direction / Start |
| `#session` (reveal visible) | `ENTER`, `SPACE` | Reveal card |
| `#session` (grades visible) | `1`, `2`, `3` | Pass / Hard / Fail |
| `#summary` | `S`, `Q` | Start New / Quit |
| Any (modal open) | `ESC` | Close modal |

---

## Error / Edge-Case Handling

- **API errors** — Red banner at top (`#error-banner`), auto-dismiss after 5 seconds. Shows `error.message` with code.
- **Loading states** — Each view shows "Loading..." during async operations.
- **Empty states** — "No phases yet", "No vocabulary for this level", "No levels" messages.
- **Session expired** — If session ID is null on `#session` render, redirect to `#home`.
- **Start without session** — Attempting to render `#session` without `App.sessionId` redirects home.

---

## Styling (`style.css`)

- Black background (`#000`) throughout.
- White text for headings and card fronts.
- Gray (`#6b7280` / `#9ca3af`) for secondary text.
- Green (`#4ade80`), gold (`#fbbf24`), red (`#f87171`) for grades and accents.
- Dark gray (`#1f2937`) for button backgrounds and modal surfaces.
- Buttons have hover opacity transition.
- Monospace font for code snippets in How It Works.
- Layout is flexbox-based, not mobile-responsive.
- Modal overlay: semi-transparent black backdrop, centered content.
