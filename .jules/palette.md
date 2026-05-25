## 2024-05-25 - ARIA Labels for Icon Buttons and Forms
**Learning:** React web apps frequently miss `aria-label`s on icon-only buttons (like `EventLog` and `InfoPanel` close/clear buttons) and form controls without explicit text labels (like the `FilterToolbar` search and dropdowns). This breaks accessibility for screen reader users.
**Action:** Always verify that interactive elements lacking visible, semantic text descriptions include an appropriate `aria-label`.
