# Future Plans

Features and ideas not yet implemented.

---

## Per-category vocab sampling

Add a `category` property to vocabulary entries (noun, verb, adjective, adverb, etc.) and change the sampling rule from "sample 50 random words across the entire vocab set (or all if ≤ 50)" to "sample N random words from each category (or all words for that category if ≤ N)". This ensures an even distribution of word types in every session rather than letting the draw be lopsided.

---

## Multiple journeys (language-agnostic)

The current model assumes Korean only. Add a top-level "Journey" container (one per language) so the app can support multiple languages. The phase/level/sentence model would be nested under a journey.

Phase 3's original plan noted this as a deferred item — the database schema and repository interface currently have no journey concept.

---

## Sentence reuse / regenerate strategy

Every session currently makes a fresh LLM call to generate 10 sentences. Generated sentences are saved to the `sentences` table but never read back. A future strategy could:
- Check for existing unsentenced sentences for the current level before generating.
- Implement a TTL-based regeneration policy.
- Allow manual regeneration of specific sentences.

---

## Grammar accumulation for 100+ levels

The current approach packs all grammar rules from levels 1..N into the LLM prompt. For 100+ levels, this prompt could exceed context windows. A strategy is needed for grammar accumulation at scale (e.g., summarization, sliding window, retrieval-augmented generation).

---

## User accounts

Multi-user support with persistent progress tracking across sessions. This would require:
- Authentication (session-based or JWT).
- Per-user session history and grade tracking.
- Per-user progress data (e.g., spaced repetition state per card).

---

## Frontend tests

No frontend testing infrastructure exists. The vanilla JS SPA has no tests. Options:
- Manual test plan (existing, in `ai/` history).
- Headless browser tests (Playwright/Cypress) for session flow and CRUD operations.

---

## Mobile responsiveness

The current CSS is desktop-only. The SPA is not responsive or mobile-friendly.

---

## Persistent session state

Sessions are currently ephemeral — refreshing the browser returns to `#home` with no session state. Could store active session in `sessionStorage`/`localStorage` to survive reloads.

---

## Session expiry handling

The original UI plan described redirecting to `#home` with a "Session expired" message when a session API returns 404. Currently, a generic error banner covers all API failures, but specific session-expired handling was not implemented.
