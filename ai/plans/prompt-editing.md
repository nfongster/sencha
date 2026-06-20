# Prompt Editing

Allow the client to view and edit the LLM sentence-generator prompt template at runtime.

---

## Problem

The prompt template (`backend/internal/sengen/prompt.tmpl`) is compiled into the binary via `//go:embed`. Changing it requires editing the file and rebuilding. The user (teacher/curator) should be able to tune the prompt from the UI without touching the server.

---

## Solution

### 1. Un-embed the prompt template

**File:** `backend/internal/sengen/sengen.go`

Replace `//go:embed prompt.tmpl` with a file-path-based loader:

- At startup, read `prompt.tmpl` from the **working directory** (the `backend/` dir where the server runs). Fall back to the embedded copy if the file doesn't exist (for production builds where only the binary is deployed).
- Add a `ReloadPrompt()` function that re-reads the file at runtime.
- The `buildPrompt` function uses the mutable variable (not the embedded const).

```go
var promptTmplSrc string  // mutable; loaded from file at startup, reloadable via ReloadPrompt()

func Init(cfg *config.LLMConfig) {
    globalConfig = cfg
    loadPrompt()
}

func loadPrompt() {
    data, err := os.ReadFile("prompt.tmpl")
    if err != nil {
        // fall back to embedded default
        data, _ = embeddedPromptTmpl.ReadFile("prompt.tmpl")
    }
    promptTmplSrc = string(data)
}

func ReloadPrompt() error {
    loadPrompt()
    return nil
}
```

### 2. Add API endpoints

**File:** `backend/internal/handler/routes.go` and new handler file (or add to an existing one)

| Method | Path | Purpose |
|--------|------|---------|
| `GET` | `/api/prompt` | Return current prompt text as `{"text": "..."}` |
| `PUT` | `/api/prompt` | Update prompt text. Body: `{"text": "..."}`. Writes to file, calls `ReloadPrompt()`. |

The `PUT` handler writes the text to `prompt.tmpl` on disk, then calls `ReloadPrompt()` so subsequent generations use the new template.

**Error responses:**
- `PUT` with empty text → `400 {"error": "prompt text is required", "code": "INVALID_PROMPT"}`
- File write failure → `500 {"error": "failed to save prompt", "code": "PROMPT_SAVE_ERROR"}`

### 3. Update "How it Works" page

**File:** `frontend/app.js` — `renderHowItWorks()`

Add a new section at the bottom of the page:

```
[Edit Generator Prompt]
```

Clicking it opens a modal with:
- `<textarea>` pre-filled with the current prompt (fetched via `GET /api/prompt`)
- Save button → `PUT /api/prompt` with the textarea content
- Cancel button → closes modal

Add a corresponding API method to the client:
```js
getPrompt()    // GET /api/prompt → {text: "..."}
updatePrompt(text)  // PUT /api/prompt → {message: "..."}
```

### 4. Styling

**File:** `frontend/style.css`

- The prompt editor textarea should be tall (min 400px) and monospace.
- Add a `.prompt-editor` class for the textarea styling.

---

## Files touched

| File | Change |
|------|--------|
| `backend/internal/sengen/sengen.go` | Replace embed with file-based loader; add `ReloadPrompt()` |
| `backend/internal/sengen/prompt.tmpl` | Keep as file; update `loadPrompt()` to read from disk |
| `backend/internal/handler/routes.go` | Register `GET /api/prompt` and `PUT /api/prompt` |
| `backend/internal/handler/prompt.go` | New file — `GetPromptHandler`, `UpdatePromptHandler` |
| `frontend/app.js` | Add `getPrompt()`/`updatePrompt()` to API client; update `renderHowItWorks()` |
| `frontend/style.css` | Add `.prompt-editor` textarea styles |

---

## Testing

1. **Backend:** `curl GET /api/prompt` returns the current prompt template.
2. **Backend:** `curl -X PUT /api/prompt -d '{"text":"Create {{.Count}} sentences..."}'` updates the prompt.
3. **Backend:** After update, start a session → generated sentences follow the new prompt.
4. **Frontend:** "How it Works" page shows the "Edit Generator Prompt" link.
5. **Frontend:** Modal opens, textarea is pre-filled with current prompt.
6. **Frontend:** Save button updates the prompt; subsequent sessions use the new template.
7. **Edge case:** Empty prompt is rejected with 400.
8. **Edge case:** Server restart preserves the edited prompt (file persisted on disk).
