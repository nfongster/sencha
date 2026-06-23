# Manual Test Plan

## Prerequisites

- Server running at `http://localhost:8080`
- Open a browser to `http://localhost:8080`
- Works with both PostgreSQL and in-memory backends

---

## 1. Categories API

```bash
curl http://localhost:8080/api/levels/categories
```

**Expect:** `{"categories":["determiner","noun","pronoun"]}`

---

## 2. LoadLevelData returns category on vocab

- Journey → click any level dot → check the modal's vocabulary table
- **Expect:** each row shows Korean, English, **and Category** (noun/pronoun/determiner)

---

## 3. Adding a level with categories

- Journey → "Add Level"
- Pick a phase, fill in grammar, add 2–3 vocab rows
- **For each row, verify the category `<select>` dropdown** lists `determiner`, `noun`, `pronoun`
- Select a mix of categories
- Submit

**Verify:**
```bash
curl http://localhost:8080/api/levels/categories
```
Category set should still be `["determiner","noun","pronoun"]`.

---

## 4. Editing existing vocab categories

- Journey → click a level → "Edit Vocab"
- **Expect:** each row's `<select>` shows the pre-set category as selected
- Change a category on one row, add a new row with a different category
- Save
- Re-open the modal — **expect** the changed categories persisted

---

## 5. Per-category sampling

`LoadLevelData` partitions 50 across categories.

Add extra vocab entries if needed (seed data has 30 — need 50+ total).

- Start a session at any level
- The LLM prompt will contain a mix proportionally distributed across categories
- With 3 categories: perCat = 50/3 = 16, remainder = 2 → first category (alphabetically "determiner") gets 18, others get 16 each

---

## 6. Empty categories fallback

- If no categories exist in the DB, `LoadLevelData` falls back to flat random sampling of 50 entries
- `/api/levels/categories` returns `{"categories":[]}`

---

## 7. Frontend edge cases

- **Modal with no vocab:** Edit Vocab on an empty level — form shows one empty row with populated category `<select>`
- **Keyboard navigation:** `ESC` closes modal
- **Remove row:** clicking `×` on a vocab row removes it before submit
- **Multiple modals:** open level detail → "Edit Vocab" → close → open another level detail — no stale data

---

## 8. Regression checks (Phase 1)

- [ ] Creating a session (all 3 directions)
- [ ] Reveal / grade cycle
- [ ] Journey tree view, phase add/edit/delete
- [ ] Level delete
- [ ] Grammar edit
- [ ] All backend tests pass: `cd backend && go test ./... -count=1`

---

## 9. Session Persistence (feature/2)

### 9a. Refresh during active session

1. Start a session (any level, any direction)
2. Reveal the first card
3. Grade it (e.g., "pass")
4. **Refresh the page** (F5 / Cmd+R)
5. **Expect:** same card is displayed (front shows, back shows if revealed), progress graph shows the graded card in green, grade counts at bottom match
6. Continue grading remaining cards
7. **Expect:** session completes normally → summary screen

### 9b. Full state restoration

1. Start a session, reveal the card
2. **Refresh** — expect the card is still in "revealed" state (back visible, grade buttons shown)
3. Grade "hard", refresh again — expect progress graph shows yellow for that card, counts reflect 1 hard

### 9c. State cleared after completion

1. Complete a full session
2. On summary screen, verify grade counts match what you chose
3. Refresh the page — expect home screen (no stale session state)

### 9d. Session expired (server restart)

1. Start a session, reveal a card
2. **Kill the backend server** (`Ctrl+C`)
3. **Restart the server**
4. In the browser, grade the card (press 1/2/3)
5. **Expect:** redirected to `#home?expired=1` with an orange warning banner: "Session expired. The server was restarted — please start a new session."
6. Click Start → verify you can start a new session normally

### 9e. Start New / Quit clears state

1. Start a session, complete it
2. On summary, press **S** (Start New) — verify taken to setup, no stale session data
3. Go to summary via back-navigation trick (or start another session, complete it)
4. Press **Q** (Quit) — verify taken to home, session state cleared

---

## 10. Prompt Editing (feature/3)

### 10a. Open prompt editor

1. Navigate to **How It Works** (`#how-it-works`)
2. **Expect:** "Edit Generator Prompt" button is visible (below the description text)
3. Click it — a modal opens with a large monospace textarea
4. **Expect:** textarea contains the current prompt template text (starts with "Create N Korean-English sentence pairs...")

### 10b. Edit and save prompt

1. Open the prompt editor
2. Append a comment to the end: `# TEST: modified prompt`
3. Click **Save**
4. **Expect:** success message, modal closes
5. Re-open the editor — **expect** the text includes your appended line

### 10c. New prompt takes effect

1. After saving a modified prompt, start a session
2. **Expect:** generated sentences reflect the modification (e.g., if you added a constraint like "only use nouns", verify sentences use only nouns)
3. Revert to the original prompt (copy original text back, save)

### 10d. Validation — empty prompt rejected

1. Open prompt editor
2. Clear the textarea entirely
3. Click **Save**
4. **Expect:** error message (red banner at top) indicating empty prompt is invalid
5. **Expect:** modal stays open so you can fix it

### 10e. Persistence across server restart

1. Edit the prompt and save
2. **Restart the backend server**
3. Navigate to How It Works → open prompt editor
4. **Expect:** the edited prompt is still there (saved to disk)

---

## 11. Client-Controlled Sentence Generation (feature/4)

### 11a. Sentence count display

1. Navigate to **Journey** (`#rules`)
2. Click any level dot
3. **Expect:** the modal shows a sentence count line: "N sentences generated" with a "view" link
4. For a fresh level with no sentences, expect "0 sentences generated"

### 11b. Generate sentences

1. Open a level detail modal (any level)
2. Click **"Generate"** button
3. **Expect:** a modal with a number input (default 10) and a Generate button
4. Enter "5" and click Generate (or press Enter)
5. **Expect:** status message "Generating..." then success — 5 sentences appear
6. **Expect:** sentence count updates to "5 sentences generated"

### 11c. View sentences

1. Open a level that has generated sentences
2. Click the **"view"** link next to the sentence count (or the "View Sentences" button if present)
3. **Expect:** a scrollable table with Korean | English columns, showing all sentences for the level
4. Close the modal

### 11d. Clear sentences

1. Open a level that has sentences
2. Click **"Clear"** button
3. **Expect:** confirmation dialog: "Are you sure? This will delete all N sentences for this level."
4. Confirm
5. **Expect:** sentence count resets to 0, sentences are deleted

### 11e. Fast path — session with existing sentences

1. Ensure a level has at least 10 sentences (generate them if needed)
2. Start a session at that level
3. **Expect:** session starts immediately with no perceptible LLM latency (sentences loaded from DB)
4. Verify all 10 cards come from the pre-generated set

### 11f. Shortfall path — mix of existing and generated

1. Ensure a level has exactly 4 sentences
2. Start a session at that level
3. **Expect:** the 4 stored sentences appear as cards, plus 6 new LLM-generated ones (10 total)
4. Complete the session normally

### 11g. Generate count bounds

1. Open Generate modal
2. Try count "0" — **expect** validation error
3. Try count "101" — **expect** validation error
4. Try count "1" — should work (generates 1 sentence)
5. Try count "100" — should work (generates 100 sentences)

### 11h. Clear then re-generate

1. Generate 5 sentences, verify count shows 5
2. Clear all sentences
3. Generate 3 sentences — **expect** count shows 3 (not cumulative)
4. View sentences — **expect** only the 3 new ones

---

## 12. Vocab Extraction (feature/5)

### 12a. Auto-fill from URL — level detail (edit mode)

1. Navigate to **Journey** (`#rules`)
2. Click any level dot
3. **Expect:** "Auto-fill from URL" button visible (between "Edit Vocab" and "Delete Level")
4. Click it — a modal opens with a URL text input and an "Extract" button

### 12b. Successful extraction

1. Open the extract modal from a level detail
2. Enter a valid Korean language learning URL (e.g., `https://en.wiktionary.org/wiki/%ED%95%98%EB%8B%A4` or a Korean study blog)
3. Click **Extract**
4. **Expect:** button shows "Extracting..." with status text
5. **Expect:** on success, modal closes and level detail re-renders with new grammar and/or vocabulary
6. Verify the grammar markdown and vocabulary entries were updated via the API

### 12c. Auto-fill from URL — Add Level modal

1. Navigate to **Journey** → **"+ Add Level"**
2. **Expect:** "Auto-fill from URL" button visible (alongside "+ Add another word")
3. Click it — URL input modal opens
4. Enter a valid URL → Extract
5. **Expect:** grammar textarea is populated with the extracted markdown
6. **Expect:** vocabulary rows are populated with the extracted entries (including categories)
7. **Note:** nothing is saved yet — this is form pre-population
8. Submit the Add Level form — **expect** the level is created with the pre-filled data

### 12d. Invalid URL

1. Open the extract modal
2. Enter an empty string — click Extract
3. **Expect:** error message (invalid URL)
4. Enter `ftp://not-supported.com` — click Extract
5. **Expect:** error message (only http/https allowed)

### 12e. Unreachable URL

1. Open the extract modal
2. Enter `http://localhost:19999/nonexistent`
3. **Expect:** error message indicating the URL could not be fetched (timeout or connection refused)

### 12f. LLM extraction failure (if possible to simulate)

- If the LLM returns unparseable JSON or missing grammar, **expect** an `EXTRACTION_FAILED` error
- (Hard to force; rely on the backend's error handling)

### 12g. Modal hygiene

1. Open extract modal from level detail → **ESC** closes it
2. Open extract modal from Add Level → **ESC** closes it
3. Open extract modal, then open another modal (e.g., Edit Vocab) — no overlap/stale state

---

## 13. Full regression

Before calling the feature complete, run through all of the above plus:

- [ ] Creating sessions in all 3 directions (Korean→English, English→Korean, Mixed)
- [ ] Reveal / grade cycle (pass, hard, fail)
- [ ] Progress graph colors update correctly
- [ ] Session summary shows correct counts and percentage
- [ ] Keyboard shortcuts work in every view
- [ ] Journey tree: phase add/edit/delete + level add/edit/delete
- [ ] Grammar edit via level detail modal
- [ ] Vocab edit via level detail modal
- [ ] Refresh during any view — no crashes
- [ ] Modal close via ESC and background click
- [ ] Error banner auto-dismisses after 5 seconds
- [ ] All backend tests pass:
  ```bash
  cd backend && go test ./... -count=1
  ```
