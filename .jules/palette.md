## 2024-05-15 - Missing ARIA labels and button types in interactive elements
**Learning:** Icon-only buttons and form elements (`<input>`, `<select>`) lacking visible `<label>` text often omit `aria-label` attributes, reducing screen reader accessibility. Additionally, buttons used strictly for JS event handling frequently lack `type="button"`, which can cause unintended form submissions.
**Action:** Always include explicit `aria-label` attributes on inputs, selects, and icon-only buttons that lack visible text. Always explicitly add `type="button"` to non-submit buttons.
