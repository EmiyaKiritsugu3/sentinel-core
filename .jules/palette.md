## 2023-10-24 - Missing ARIA Labels and Button Types on Interactive Elements
**Learning:** Found a pattern across components where icon-only buttons, text inputs, and select dropdowns lacked explicit `aria-label` attributes, rendering them inaccessible to screen readers. Action buttons also lacked `type="button"`, risking unintended form submissions.
**Action:** Always verify that interactive elements without visible text labels have descriptive `aria-label`s, and ensure `<button>` elements used strictly for JS actions explicitly declare `type="button"`.
