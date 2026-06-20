# Manual Test Plan — Per-Category Vocab Sampling

## Prerequisites

- Server running in `--dev` mode (in-memory repo, seed data loads)
- Open `http://localhost:8080` in a browser

---

## 1. Categories API

```bash
curl http://localhost:8080/api/levels/categories
```

**Expect:** `{"categories":["determiner","noun","pronoun"]}`

---

## 2. LoadLevelData returns category on vocab

- Journey → click any level dot → check the modal's vocabulary table
- **Expect:** each row shows Korean, English, **and Category** (noun/pronoun/determiner)

---

## 3. Adding a level with categories

- Journey → "Add Level"
- Pick a phase, fill in grammar, add 2–3 vocab rows
- **For each row, verify the category `<select>` dropdown** lists `determiner`, `noun`, `pronoun`
- Select a mix of categories
- Submit

**Verify:**
```bash
curl http://localhost:8080/api/levels/categories
```
Category set should still be `["determiner","noun","pronoun"]`.

---

## 4. Editing existing vocab categories

- Journey → click a level → "Edit Vocab"
- **Expect:** each row's `<select>` shows the pre-set category as selected
- Change a category on one row, add a new row with a different category
- Save
- Re-open the modal — **expect** the changed categories persisted

---

## 5. Per-category sampling (in-memory)

In `--dev` mode `LoadLevelData` partitions 50 across categories.

Add extra vocab entries if needed (seed data has 30 — need 50+ total).

- Start a session at any level
- The LLM prompt will contain a mix proportionally distributed across categories
- With 3 categories: perCat = 50/3 = 16, remainder = 2 → first category (alphabetically "determiner") gets 18, others get 16 each

---

## 6. Empty categories fallback

- If no categories exist in the DB, `LoadLevelData` falls back to flat random sampling of 50 entries
- `/api/levels/categories` returns `{"categories":[]}`

---

## 7. Frontend edge cases

- **Modal with no vocab:** Edit Vocab on an empty level — form shows one empty row with populated category `<select>`
- **Keyboard navigation:** `ESC` closes modal
- **Remove row:** clicking `×` on a vocab row removes it before submit
- **Multiple modals:** open level detail → "Edit Vocab" → close → open another level detail — no stale data

---

## 8. Regression checks

- [ ] Creating a session (all 3 directions)
- [ ] Reveal / grade cycle
- [ ] Journey tree view, phase add/edit/delete
- [ ] Level delete
- [ ] Grammar edit
- [ ] All backend tests pass: `cd backend && go test ./... -count=1`
