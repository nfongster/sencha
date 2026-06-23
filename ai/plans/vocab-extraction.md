# URL Auto-fill for Level Grammar & Vocabulary

When creating or editing a level, the client can supply a URL. The server fetches the webpage, sends the HTML to the LLM for extraction, and returns grammar rules + vocabulary that auto-populate the form for review/editing before saving.

---

## Problem

Manually typing grammar rules and vocabulary entries for each new level is tedious. Many Korean learning resources exist online — the teacher should be able to point the app at a URL and have the LLM extract structured content.

---

## Solution

### 1. Extraction prompt template

**File:** `backend/internal/sengen/extract.tmpl` (new, embedded via `//go:embed`)

The prompt instructs the LLM to extract grammar rules (Markdown) and vocabulary entries from the HTML content. Returns JSON only.

```
You are a Korean language teaching assistant. Given the following HTML content from a webpage, extract:

1. Korean grammar rules found in the content. Write them in Markdown format.
2. Korean vocabulary words found in the content, with their English translations and categories (noun, verb, adjective, adverb, pronoun, determiner, etc.).

Return your response as a JSON object with this exact structure:
{
  "grammar_markdown": "...",
  "vocabulary": [
    {"korean": "...", "english": "...", "category": "..."}
  ]
}

Only return valid JSON. Do not include any other text.

Webpage HTML:
{{.HTML}}
```

### 2. New Go types and function in sengen

**File:** `backend/internal/sengen/sengen.go`

```go
type ExtractResult struct {
    GrammarMD string                   `json:"grammar_markdown"`
    Vocabulary []repository.VocabEntry `json:"vocabulary"`
}

// testExtractFunc, when set, replaces real extraction logic for tests
type ExtractFunc func(string) (*ExtractResult, error)

var testExtractFunc ExtractFunc

func SetExtractFunc(fn ExtractFunc) {
    testExtractFunc = fn
}

func ExtractFromHTML(html string) (*ExtractResult, error) {
    if testExtractFunc != nil {
        return testExtractFunc(html)
    }
    // build prompt from extract.tmpl template
    // call callLLM(prompt)
    // parse JSON response into ExtractResult
    // validate: grammar_markdown non-empty (vocab can be empty)
    // return result or error
}
```

Template rendering helper (same pattern as `buildPrompt`):
- Parse `extract.tmpl` (embedded)
- Execute with `html` as data
- Call `callLLM(prompt)`
- `json.Unmarshal` the response into `ExtractResult`

### 3. New API endpoint

**Path:** `POST /api/levels/extract-from-url`

**Request body:**
```json
{"url": "https://example.com/korean-lesson"}
```

**Handler logic:**

1. Validate URL scheme is `http` or `https` (reject others)
2. Fetch HTML with:
   - Timeout: 10 seconds
   - Max response size: 100 KB
   - User-Agent: set to `sencha/1.0`
3. Call `sengen.ExtractFromHTML(html)`
4. Return `{"grammar_markdown": "...", "vocabulary": [...]}`

**Error responses:**

| Status | Condition |
|--------|-----------|
| 400 | Invalid URL scheme, empty URL |
| 502 | URL fetch failed (unreachable, timeout, too large) |
| 502 | LLM returned unparseable JSON |
| 500 | LLM call failed |

**File:** `backend/internal/handler/levels.go` — add `ExtractFromUrlHandler`

**File:** `backend/internal/handler/routes.go` — register:
```go
r.POST("/api/levels/extract-from-url", ExtractFromUrlHandler)
```

### 4. Frontend: API client

**File:** `frontend/app.js`

```js
extractFromUrl(url) {
    return this.post('/api/levels/extract-from-url', { url });
}
```

### 5. Frontend: Level detail modal

**File:** `frontend/app.js` — `showLevelDetail()`

Add a new button in the action bar:
```html
<button class="btn btn-sm" onclick="showExtractUrlModal(${level.number})">Auto-fill from URL</button>
```

Insert between "Edit Rules" and "Edit Vocab".

### 6. Frontend: Add Level modal

**File:** `frontend/app.js` — `showAddLevelModal()`

Add an "Auto-fill from URL" link/button below the vocab rows. Same behavior as the detail modal but targets the create form.

### 7. Frontend: URL prompt + extraction flow

**New function:** `showExtractUrlModal(levelNumber?)`

Where `levelNumber` is present when called from the level detail modal, absent for Add Level.

The modal:
1. Shows a simple overlay with a URL input field and Submit/Cancel buttons
2. On submit:
   - Disables the button, shows "Extracting..." text
   - Calls `API.extractFromUrl(url)`
3. On success:
   - **Add Level mode**: Populates the grammar textarea (`#add-level-grammar`) and replaces vocab rows with extracted entries
   - **Edit mode**: Calls `API.getLevel()` first, then:
     - Populates grammar in the level detail view (by re-rendering the detail modal)
     - Stores extracted vocab temporarily; when the user opens "Edit Vocab", it pre-populates the vocab form
4. On error:
   - Shows error message in the overlay
   - User can retry or cancel

### 8. Styling

**File:** `frontend/style.css`

- `.extract-url-modal` — small centered overlay for URL input
- No major new styles needed; reuse existing modal patterns

---

## Files touched

| File | Change |
|------|--------|
| `backend/internal/sengen/extract.tmpl` | New — extraction prompt template |
| `backend/internal/sengen/sengen.go` | Add `ExtractResult`, `ExtractFromHTML()`, `SetExtractFunc()`, extract prompt builder |
| `backend/internal/handler/levels.go` | Add `ExtractFromUrlHandler` |
| `backend/internal/handler/routes.go` | Register `POST /api/levels/extract-from-url` |
| `frontend/app.js` | Add `API.extractFromUrl()`, `showExtractUrlModal()`, update `showLevelDetail()` and `showAddLevelModal()` |
| `frontend/style.css` | Add `.extract-url-modal` class if needed |

---

## Testing

1. **Unit test:** `TestExtractFromHTML_ParsesJSON` — mock LLM response with valid JSON, verify `ExtractResult` is correctly parsed
2. **Unit test:** `TestExtractFromHTML_InvalidJSON` — mock LLM response with malformed JSON, verify error
3. **Handler test:** `TestExtractFromUrl_Success` — set mock `ExtractFunc`, POST valid URL, verify 200 + grammar + vocab in response
4. **Handler test:** `TestExtractFromUrl_InvalidURL` — POST empty/invalid URL, verify 400
5. **Handler test:** `TestExtractFromUrl_FetchFails` — POST unreachable URL (no mock), verify 502
6. **Frontend:** Add Level — click "Auto-fill from URL", enter URL, verify grammar textarea and vocab rows are populated
7. **Frontend:** Level detail — click "Auto-fill from URL", enter URL, verify detail view updates with extracted content
