# Get/Set Rules Feature — Manual Test Plan

Branch: `feat/rules`
Pushed: yes (upstream set)

## Prerequisites
- Backend server running (`cd backend && go run ./cmd/api/`)
- Console built (`cd console && go build -o console .`)

---

### Test 1: `rules get` — default seeded level

```
cd console && ./console
```

At the REPL prompt:

```
> rules
Rules subcommands: get <level_number>, set <level_number> <grammar.md> [exceptions.md]
> get 1
```

**Expect:** Prints Level 1's grammar markdown and `(none)` for exceptions.

---

### Test 2: `rules set` — update grammar from file

```bash
cat > /tmp/my-rules.md << 'EOF'
# My Custom Rules
- Always use formal speech
- Never use object markers with 이다
EOF
```

Back in the console:

```
> rules
> set 1 /tmp/my-rules.md
```

**Expect:** Prints `Grammar loaded from /tmp/my-rules.md (XX bytes)` and `Level 1 rules updated successfully.`

---

### Test 3: Verify with `rules get`

```
> rules
> get 1
```

**Expect:** Grammar now shows "My Custom Rules" instead of the original.

---

### Test 4: `rules set` with exceptions

```bash
cat > /tmp/my-exceptions.md << 'EOF'
- Watch out for irregular verb conjugations
- Remember to use proper particle agreement
EOF
```

```
> rules
> set 1 /tmp/my-rules.md /tmp/my-exceptions.md
```

**Expect:** Both grammar and exceptions are reported as loaded. Then `get 1` should show both.

---

### Test 5: Smoke test study session still works

```
> start
Direction [default: korean-to-english]: (ENTER)
```

**Expect:** Session starts normally, cards reveal and grade as before.

---

### Test 6: Invalid level number

```
> rules
> get 99
```

**Expect:** `Error fetching level 99: level not found (LEVEL_NOT_FOUND)`

---

### Test 7: Missing file path

```
> rules
> set 1 /tmp/nonexistent.md
```

**Expect:** `Error reading grammar file "/tmp/nonexistent.md": ...`

---

### Test 8: API direct test (curl)

```bash
curl -s http://localhost:8080/api/levels/1 | jq .level.grammar_md | head -c 100

curl -s -X PATCH http://localhost:8080/api/levels/1 \
  -H 'Content-Type: application/json' \
  -d '{"grammar_markdown": "# Test\nShort rules", "exceptions_markdown": "Beware of X"}' | jq .

curl -s http://localhost:8080/api/levels/1 | jq .level
```

**Expect:** PATCH returns `{"message": "level rules updated"}`, GET reflects the new content.
