# Checkpoint — feat/rules

**Status:** ✅ Merged to `main`.

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
