## 2024-08-26 - [Accessible Icon Buttons]
**Learning:** Icon-only buttons (like the Close "X" in panels or Clear "✕" in logs) need explicit `aria-label` attributes for screen readers, and modifying them triggers maintainability checks if `type="button"` is missing.
**Action:** Always include both `aria-label` and `type="button"` when implementing or updating icon-only buttons.
