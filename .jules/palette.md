## 2024-08-21 - Icon-Only Buttons Accessibility
**Learning:** Icon-only buttons (like 'X' or '✕' for close/clear actions) need explicit `aria-label`s for screen readers, and `type="button"` to avoid default form submission behavior and maintainability issues.
**Action:** Always add `aria-label` and `type="button"` when creating icon-only buttons in React components.
