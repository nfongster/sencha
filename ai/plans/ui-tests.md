# UI Tests (Playwright)

Add automated UI tests for the vanilla JS frontend SPA using Playwright with API mocking.

---

## Problem

The frontend has zero automated tests. All testing is manual (documented in `ai/manual-test-plan.md`). This makes regression detection slow and unreliable as the app grows.

---

## Approach

**Playwright** with API route interception — no Go backend needed during tests.

- Tests run in a real Chromium browser (headless in CI).
- API calls are mocked at the Playwright route level (`page.route()`).
- Frontend files served via Python's built-in HTTP server (`python3 -m http.server`).
- No build tools, no bundler, no transpiler needed.

---

## Setup

### 1. Initialize npm + install Playwright

```bash
cd frontend
npm init -y
npm install -D @playwright/test
npx playwright install chromium
```

### 2. Create Playwright config

**File:** `frontend/playwright.config.js`

```js
const { defineConfig } = require('@playwright/test');

module.exports = defineConfig({
  testDir: './tests',
  fullyParallel: true,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 1 : 0,
  workers: process.env.CI ? 2 : undefined,
  use: {
    baseURL: 'http://localhost:3000',
    headless: true,
  },
  webServer: {
    command: 'python3 -m http.server 3000 --directory ..',
    port: 3000,
    reuseExistingServer: !process.env.CI,
    cwd: 'frontend',
  },
});
```

Note: `--directory ..` so the server serves the repo root as `/`, making
`/static/app.js` and `/static/style.css` resolve correctly (the HTML loads
these as `/static/...` paths).

---

## API Mocking Pattern

```js
test.beforeEach(async ({ page }) => {
  await page.route('**/api/levels/max', route =>
    route.fulfill({ contentType: 'application/json', body: JSON.stringify({ max: 5 }) })
  );
  await page.route('**/api/phases', route =>
    route.fulfill({ contentType: 'application/json', body: JSON.stringify({
      phases: [
        { number: 1, name: 'Phase 1' },
        { number: 2, name: 'Phase 2' },
      ]
    })})
  );
  // … per-file overrides
});
```

Each test file sets up its own mocks in `beforeEach`. Tests that need specific
responses can override individual routes within a `test()` block.

---

## Test File Structure

```
frontend/tests/
├── home.spec.js
├── level-select.spec.js
├── setup.spec.js
├── session.spec.js
├── summary.spec.js
├── how-it-works.spec.js
├── rules.spec.js
└── keyboard.spec.js
```

---

## Test Specifications

### `home.spec.js`

| Test | Assertions |
|------|------------|
| Renders title | `Sencha` heading visible |
| Three navigation buttons | Text: `[1] Start`, `[2] Journey`, `[3] How it Works` |
| Click Start | Navigates to `#level-select` |
| Click Journey | Navigates to `#rules` |
| Click How it Works | Navigates to `#how-it-works` |

### `level-select.spec.js`

| Test | Assertions |
|------|------------|
| Renders latest level | Button shows `Latest Level (5)` |
| API failure fallback | Button shows `Latest Level` (no number) |
| Click latest level | Navigates to `#setup`, sets `selectedLevel = 5` |
| Click choose level | Navigates to `#level-picker` |
| Picker renders phases | Phase names visible, level dots visible |
| Click level dot | Modal opens with grammar + vocab |
| Start button in modal | Closes modal, navigates to `#setup` |
| No levels message | Phase shows `No levels` |
| API failure | Error banner shown |

### `setup.spec.js`

| Test | Assertions |
|------|------------|
| Three direction options | KR→EN, EN→KR, Mixed visible and clickable |
| Default selection | `korean-to-english` has `.selected` class |
| Click changes selection | Clicked option gains `.selected`, others lose it |
| Keyboard 1/2/3 | Selects corresponding direction |
| Enter without selection | Uses default (korean-to-english) |
| Start session success | Calls API with correct direction + level, navigates to `#session` |
| Start session API failure | Error banner shown, stays on `#setup` |

### `session.spec.js`

| Test | Assertions |
|------|------------|
| Redirect without sessionId | Navigates to `#home` |
| Progress graph renders | 10 circles in SVG, first has white outline |
| Stats bar | Shows `Completed: 0 / 10`, all zeros |
| Card front | Shows Korean text (from mock reveal response) |
| Reveal button visible | Green `Reveal` button present |
| Click reveal | Back text appears, grade buttons replace reveal |
| Enter/Space reveal | Same as click reveal |
| Grade buttons after reveal | `[1] Pass` (green), `[2] Hard` (gold), `[3] Fail` (red) |
| Grade pass | Graph updates circle to green, stats increment, next card loaded |
| Grade hard | Graph updates to gold |
| Grade fail | Graph updates to red |
| Keyboard 1/2/3 grade | Same as button click |
| All 10 cards graded | Auto-navigates to `#summary` |
| Grade API failure | Error banner shown |
| Reveal API failure | Error banner shown |

### `summary.spec.js`

| Test | Assertions |
|------|------------|
| Renders stats | Pass/Hard/Fail counts visible |
| Percentage calculated | e.g., 7/10 = 70% |
| Start New button | Navigates to `#setup` |
| Quit button | Navigates to `#home` |
| Keyboard S | Navigates to `#setup` |
| Keyboard Q | Navigates to `#home` |

### `how-it-works.spec.js`

| Test | Assertions |
|------|------------|
| All sections present | Phases, Generation, Grading, Shortcuts headings visible |
| Back button | Navigates to `#home` |

### `rules.spec.js`

| Test | Assertions |
|------|------------|
| Tree renders | Phase names visible with level dots |
| Click level dot | Detail modal with grammar + vocab + action buttons |
| Edit Rules opens form | Modal with pre-filled grammar textarea |
| Save rules | Calls updateLevel API, modal closes |
| Edit Vocab opens form | Dynamic vocab rows with inputs and category dropdowns |
| Add vocab row | New empty row appended |
| Remove vocab row | Row deleted from DOM |
| Save vocab | Calls setVocabulary API with row data |
| Delete Level confirmation | Modal shows warning + renumbering info |
| Confirm delete level | Calls deleteLevel API, tree re-renders |
| Cancel delete | Modal closes, no API call |
| Add Phase modal | Number auto-filled, name input |
| Submit add phase | Calls createPhase API, tree re-renders |
| Phase edit pencil | Modal with pre-filled name |
| Save phase edit | Calls updatePhase API, tree re-renders |
| Phase delete icon | Confirmation modal with level count warning |
| Confirm delete phase | Calls deletePhase API, tree re-renders |
| Add Level modal | Phase dropdown, grammar textarea, vocab rows |
| Submit add level | Calls createLevel API with all fields |
| Empty phases state | Shows `No phases yet` |
| Empty levels state | Shows `No levels` |
| Empty vocab state | Shows `No vocabulary for this level` |
| API error on CRUD | Error banner shown, modal stays open |

### `keyboard.spec.js`

| Test | Assertions |
|------|------------|
| Escape closes modal | Open modal → Escape → modal removed |
| Home 1 | Navigates to `#level-select` |
| Home 2 | Navigates to `#rules` |
| Home 3 | Navigates to `#how-it-works` |
| Level select 1 | Triggers startWithLatest |
| Level select 2 | Navigates to `#level-picker` |
| Setup 1/2/3 | Changes selected direction |
| Setup Enter | Clicks the Start button |
| Session Enter/Space | Clicks reveal button (when visible) |
| Session 1/2/3 | Clicks grade buttons (when visible) |
| Summary S | Navigates to `#setup` |
| Summary Q | Navigates to `#home` |
| Default prevented | `e.preventDefault()` called for handled keys |

---

## Edge Cases

| Scenario | Coverage |
|----------|----------|
| API failure: load levels | level-select.spec.js |
| API failure: create session | setup.spec.js |
| API failure: reveal card | session.spec.js |
| API failure: grade card | session.spec.js |
| API failure: rules CRUD | rules.spec.js |
| Session without sessionId | session.spec.js |
| Empty level picker | level-select.spec.js |
| Empty vocab | rules.spec.js |
| Empty phases | rules.spec.js |
| Level delete with renumbering | rules.spec.js |
| Phase delete with levels | rules.spec.js |
| Phase delete without levels | rules.spec.js |
| Keyboard while modal open | keyboard.spec.js |

---

## CI Integration

Add a `ui-tests` job to `.github/workflows/ci.yml`:

```yaml
ui-tests:
  name: UI Tests
  runs-on: ubuntu-latest
  steps:
    - uses: actions/checkout@v4
    - uses: actions/setup-node@v4
      with:
        node-version: '18'
    - run: npm ci
      working-directory: frontend
    - run: npx playwright install --with-deps chromium
      working-directory: frontend
    - run: npx playwright test
      working-directory: frontend
```

---

## Files Touched

| File | Change |
|------|--------|
| `frontend/package.json` | New — npm init, add `@playwright/test` |
| `frontend/playwright.config.js` | New — Playwright configuration |
| `frontend/tests/home.spec.js` | New |
| `frontend/tests/level-select.spec.js` | New |
| `frontend/tests/setup.spec.js` | New |
| `frontend/tests/session.spec.js` | New |
| `frontend/tests/summary.spec.js` | New |
| `frontend/tests/how-it-works.spec.js` | New |
| `frontend/tests/rules.spec.js` | New |
| `frontend/tests/keyboard.spec.js` | New |
| `.github/workflows/ci.yml` | Add `ui-tests` job |

---

## Notes

- The `webServer` uses Python's built-in HTTP server to avoid Node.js server
  dependencies. Python is available on Ubuntu (CI) and most dev machines.
- The server serves from the repo root (`--directory ..`) so that
  `/static/app.js` and `/static/style.css` resolve from `/static/` paths
  (matching the Go backend's URL structure).
- `marked.js` loads from CDN via `index.html`. Tests should wait for the
  `marked` global before interacting with views that parse markdown.
- First-time setup requires downloading Chromium (~200 MB via
  `npx playwright install chromium`).
- State resets naturally between tests via `page.goto()` (full page reload
  reinitialises the `App` object).
