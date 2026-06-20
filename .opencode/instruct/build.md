## Test verification

After making code changes, run `go test ./backend/...` to verify. If you
prefer to delegate, use the `tester` subagent via the `task` tool:

> "Tester, run the backend tests and report failures."

The tester will run tests and report failures. Fix any issues the tester finds,
then re-run tests before committing.
