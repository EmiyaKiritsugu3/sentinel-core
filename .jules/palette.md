## 2024-05-24 - Accessibility of form controls and icon buttons
**Learning:** Inputs, selects, and icon-only buttons without visible labels require explicit `aria-label` attributes to ensure screen readers can announce their purpose correctly. In addition, all buttons used for JavaScript actions should have `type="button"` to prevent unintended form submissions.
**Action:** Always verify that `<input>`, `<select>`, and icon-only `<button>` elements have descriptive `aria-label`s, and ensure all JS-handled buttons explicitly set `type="button"`.
