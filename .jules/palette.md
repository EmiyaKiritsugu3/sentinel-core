## 2024-08-23 - Accessible Icon Buttons and Types
**Learning:** Icon-only buttons (like "X" or "✕" for close/clear) require explicit `aria-label`s for screen reader accessibility. Additionally, all `<button>` elements in React should have an explicit `type="button"` attribute to prevent unintended form submissions and meet SonarCloud maintainability checks.
**Action:** Always include `aria-label` on buttons that lack descriptive text content, and consistently apply `type="button"` to non-submit buttons across all React components.
