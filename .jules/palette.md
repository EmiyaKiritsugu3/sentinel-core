## 2025-03-01 - Interactive Element Accessibility
**Learning:** React components using `<button>` exclusively for JavaScript event handling (like "Clear" or "Close" icon buttons) require explicit `type="button"` to prevent unintended form submissions, and icon-only buttons need `aria-label` for screen reader compatibility. Inputs also need `aria-label` if they lack `<label>` wrappers.
**Action:** Always verify that interactive icon elements use semantic `<button type="button">` tags with descriptive `aria-label`s, and stand-alone inputs have accessibility labels.
