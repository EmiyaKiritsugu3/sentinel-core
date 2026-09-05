## 2025-03-05 - Accessibility in React Components
**Learning:** Icon-only buttons and form inputs without visible labels require descriptive `aria-label` attributes for screen reader compatibility, and all JS-only buttons need explicit `type="button"` to prevent unintended form submissions.
**Action:** Always verify `aria-label` and `type="button"` attributes are present on these components when auditing UI elements.
