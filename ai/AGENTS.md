# Agent Instructions

## Plan workflow
- Active plans are stored in `ai/plans/` as `.md` files.
- When starting work on a plan, read it first, then implement.

## On completing a feature
1. Update `ai/backend-architecture.md` and/or `ai/frontend-architecture.md`
   if the feature changes any documented behavior, API surface, or
   architectural patterns.
2. Delete the corresponding plan file from `ai/plans/`.

## Tester subagent

A `tester` subagent is registered in `opencode.json` (file at
`.opencode/agents/tester.md`). Use it to delegate test runs:

> "Tester, run the backend tests and report failures."

The tester has `edit: deny` — it reads and runs tests only, does not fix code.
