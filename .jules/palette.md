## 2025-01-20 - Missing Accessible Names on Inputs & Buttons
**Learning:** Found a systemic pattern where icon-only buttons and label-less form inputs (like search/select) are missing explicit `aria-label`s. Furthermore, React `<button>` elements used strictly for JS interactions lack `type="button"`, which can cause unintended form submissions and a11y issues.
**Action:** Always verify that every `<button>` has an explicit `type="button"` and that any interactive elements without visible labels have an `aria-label`.
