## 2026-05-24 - Add type="button" and aria-labels to buttons
**Learning:** Icon-only buttons used strictly for JS event handling must explicitly have `type="button"` to prevent unintended form submissions, and `aria-label` for screen reader accessibility.
**Action:** Always verify that `<button>` tags specify `type="button"` when not explicitly intended as form submit buttons, and include `aria-label`s for icons.
