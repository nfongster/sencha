# Session Persistence & Expiry Handling

Improve the frontend session UX by surviving browser refreshes and gracefully handling server-restarted sessions.

---

## Problem

1. **No persistence:** Refreshing the browser during a session resets all `App` state to defaults. `sessionId` is null, so `renderSession()` redirects to `#home`. Progress is lost.

2. **No expiry UX:** Sessions live in the backend's in-memory store (`backend/internal/store/memory.go`). If the server restarts, all sessions are wiped. Currently, reveal/grade API calls return a 404 with `{"error":"session not found","code":"NOT_FOUND"}`. The frontend shows a generic red banner ("Failed to reveal card: ...") that auto-dismisses in 5 seconds, leaving the user stranded on the session screen.

---

## Changes

### 1. Persist App state to `sessionStorage`

**File:** `frontend/app.js`

Add a save/restore mechanism around the `App` object:

- **After every state mutation** (`startSession`, `revealCard`, `gradeCard`, `backReveal`), call a `saveSessionState()` function that writes a subset of `App` to `sessionStorage`.

```js
function saveSessionState() {
  const state = {
    sessionId: App.sessionId,
    direction: App.direction,
    totalCards: App.totalCards,
    cardsRemaining: App.cardsRemaining,
    currentIndex: App.currentIndex,
    cardStates: App.cardStates,
    gradeCounts: App.gradeCounts,
    sessionComplete: App.sessionComplete,
    currentCard: App.currentCard,
    backRevealed: App.backRevealed,
    selectedLevel: App.selectedLevel,
  };
  sessionStorage.setItem('sencha-session', JSON.stringify(state));
}
```

- **On page load** (inside the `load` event listener, before `router()` is called):

```js
function restoreSessionState() {
  const saved = sessionStorage.getItem('sencha-session');
  if (!saved) return;
  try {
    const state = JSON.parse(saved);
    Object.assign(App, state);
  } catch (_) { /* ignore corrupt data */ }
}
```

- **On session completion or explicit quit**, clear `sessionStorage`:

```js
sessionStorage.removeItem('sencha-session');
```

### 2. Verify session on restore

After restoring state from `sessionStorage`, verify the session is still alive on the backend by calling `GET /api/sessions/:id`. This catches the case where the server restarted between refreshes.

```js
async function verifyAndRender() {
  if (!App.sessionId) return;
  try {
    await API.get(App.sessionId);
    // Session exists — re-render the current view
    router();
  } catch (err) {
    // Session expired — clean up and show message
    clearSession();
    sessionStorage.removeItem('sencha-session');
    location.hash = '#home';
    // Show a non-auto-dismissing message on home
  }
}
```

This should be called from the `load` event listener after `restoreSessionState()`.

### 3. Session expiry handling in API calls

**File:** `frontend/app.js`

In `revealCard()` and `gradeCard()`, detect `NOT_FOUND` errors and handle them specifically:

```js
async function revealCard() {
  try {
    const data = await API.revealCard(App.sessionId);
    // ... existing logic ...
  } catch (err) {
    if (err.message.includes('NOT_FOUND')) {
      clearSession();
      sessionStorage.removeItem('sencha-session');
      location.hash = '#home';
      // Show session expired message on home
      return;
    }
    showError('Failed to reveal card: ' + err.message);
  }
}
```

Same pattern for `gradeCard()`.

### 4. "Session expired" message

On home, if the user was redirected due to session expiry, show a distinct message:

**Option A:** Pass a query param: `location.hash = '#home?expired=1'`. The `renderHome()` function checks for it and renders a yellow warning banner.

**Option B:** Use a different message style — a yellow/orange banner that doesn't auto-dismiss (unlike the red error banner). Add a `showPersistentMessage(msg)` function.

### 5. Clear on completion/quit

In `renderSummary()`, when the user clicks **Start New** or **Quit**, clear `sessionStorage`. Also clear when explicitly navigating away from `#home` to start a new session.

---

## Files touched

| File | Change |
|------|--------|
| `frontend/app.js` | Add `saveSessionState()`, `restoreSessionState()`, `clearSession()`. Add `verifyAndRender()`. Update `startSession`, `revealCard`, `gradeCard`, `backReveal` to call save/restore. Update `revealCard`/`gradeCard` to handle NOT_FOUND. Clear storage on session end. |
| `frontend/style.css` | Add style for persistent message banner (yellow/orange) if Option B is chosen. |

---

## Testing

1. Start a session, grade a few cards, refresh the browser → should restore to the same card.
2. Restart the backend server, then try to reveal/grade → should redirect to home with "Session expired" message.
3. Complete a session, click "Start New" → old session state should be cleared.
4. Complete a session, click "Quit" → old session state should be cleared.
