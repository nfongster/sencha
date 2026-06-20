# Checkpoint — feat/rules

**Status:** ✅ Merged to `main`.

> **⚠️ MAINTENANCE NOTE:** The "How Sencha Works" popup in `frontend/app.js` (`showHowItWorks`) describes the sentence generation rules, vocabulary selection, grading system, and keyboard shortcuts. If any of these behaviors change, the popup text must be updated to match.

What was done:
- Added `UpdateLevel` to the Repository interface (memory + postgres implementations)
- Added `PATCH /api/levels/:number` endpoint for updating grammar/exceptions
- Added `rules get` and `rules set` subcommands to the console REPL
- All backend tests pass
- Manual test plan executed: all 8 tests passed
- Two bugs found and fixed during testing:
  1. Missing JSON tags on `Level`/`VocabEntry` structs causing zero-value deserialization
  2. Index-out-of-range panic in `rules set` when exceptions path omitted

Next up: pick from the deferred Phase 3 items, or start a new feature.
