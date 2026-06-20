# UI Phase 1 — Manual Test Plan

**Branch:** `feat/ui-phase1`

## Prerequisites

1. Backend running: `cd backend && go run ./cmd/api/`
2. Open `http://localhost:8080` in a browser
3. Use in-memory mode (remove or comment out `database_url` in `config.json`)

---

## Test 1: Homepage loads

1. Navigate to `http://localhost:8080`
2. **Expect:**
   - Black background
   - "Sencha" title centered in white
   - Two buttons: **[1] Start** and **[2] Rules**
3. Press `1` → should navigate to session setup (`#setup`)
4. Press `2` → should navigate to rules page (`#rules`)
5. Navigate back to `#home`

---

## Test 2: Session setup — direction selection

1. From home, click **Start** or press `1`
2. **Expect:**
   - Title "Select Direction"
   - Three options:
     - [1] Korean → English (pre-selected)
     - [2] English → Korean
     - [3] Mixed
   - **Start Session** button (green)
3. Click each option → border highlight should move
4. Press `1`, `2`, `3` → each selects the corresponding option
5. Press Enter → should trigger **Start Session**

---

## Test 3: Session — happy path

### 3a: Session creation
1. On setup page, select "Korean → English"
2. Click **Start Session**
3. **Expect:**
   - SVG progress graph at top: 10 gray circles connected by lines
   - Statistics: "Completed: 0 / 10" + "Pass: 0 Hard: 0 Fail: 0"
   - Card area: "Press Reveal to see the card" in gray
   - Green **Reveal** button

### 3b: Reveal a card
1. Click **Reveal** (or press ENTER or SPACE)
2. **Expect:**
   - Front text (Korean) displayed large, white, centered
   - Thin gray divider
   - Back text (English) below in gray
   - Reveal button replaced by three grade buttons: **[1] Pass** (green), **[2] Hard** (gold), **[3] Fail** (red)
   - First progress circle turns gray with white highlight ring

### 3c - 3e: Grade cards
1. Press `1` (Pass) → circle turns green, stats update: "Completed: 1/10, Pass: 1"
2. Reveal next → press `2` (Hard) → circle turns gold, "Hard: 1"
3. Reveal next → press `3` (Fail) → circle turns red, "Fail: 1"
4. Grade remaining cards with any mix
5. After 10th card → auto-navigate to summary

---

## Test 4: Session summary

1. After completing all 10 cards:
2. **Expect:**
   - "Session Complete!" title
   - Three stat columns: Pass (green), Hard (gold), Fail (red) with counts
   - Percentage (e.g., "60%")
   - **[S] Start New** (green) and **[Q] Quit** buttons
3. Press `S` → goes to setup; press `Q` → goes to home
4. Run a second session to verify clean state

---

## Test 5: Rules — browse levels

1. From home, click **Rules** or press `2`
2. **Expect:**
   - Left sidebar: "Phase 1" with a dot, level "1" circle below
   - "+ Add Phase" and "+ Add Level" buttons
   - Right panel: "Select a level to view rules"
3. Click level "1" node
4. **Expect:**
   - Modal with grammar markdown (rendered) and vocabulary list
   - Close button (×) works; pressing ESC closes; clicking background closes

---

## Test 6: Add a phase

1. Click **+ Add Phase**
2. **Expect:**
   - Modal with auto-filled number (2) and name ("Phase 2")
   - Cancel and Create buttons
3. Change name to "Custom Phase", click **Create**
4. **Expect:** Modal closes, rules page shows both phases
5. Click the new phase → "No levels"

---

## Test 7: Add a level

1. Click **+ Add Level**
2. **Expect:**
   - Phase dropdown, grammar textarea, exceptions textarea, vocabulary row
   - "+ Add another word" button
3. Fill in:
   - Phase: "Custom Phase"
   - Grammar: `# My Level\n- Rule 1\n- Rule 2`
   - Exceptions: `- Watch out for X`
   - Vocab row 1: Korean `테스트` / English `test`
   - Click "+ Add another word", row 2: Korean `예시` / English `example`
4. Click **Create**
5. **Expect:** New level appears under "Custom Phase" with next sequential number
6. Click the new level node → verify grammar + vocab in pop-up

---

## Test 8: Keyboard shortcuts

1. Start a session
2. ENTER or SPACE triggers reveal (when reveal button visible)
3. `1`/`2`/`3` triggers Pass/Hard/Fail (when grade buttons visible)
4. Test all keyboard inputs across all 10 cards

---

## Test 9: Errors

1. Stop the backend server
2. Try to start a session
3. **Expect:** Red error banner at top: "Failed to create session: ..."
4. Banner auto-dismisses after ~5 seconds
5. Restart server, verify session works again

---

## Test 10: Visual and edge cases

1. Black background everywhere (`#000`)
2. Buttons have hover effect (opacity change)
3. Progress graph colors:
   - `#4b5563` (gray) = unrevealed
   - `#6b7280` (lighter gray) = revealed
   - `#4ade80` (green) = pass
   - `#fbbf24` (gold) = hard
   - `#f87171` (red) = fail
4. Rules page with no phases shows "No phases yet"
5. Level pop-up with no vocab shows "No vocabulary for this level"
6. Browser refresh during session → returns to home
7. Modal closes on ESC and on background click
