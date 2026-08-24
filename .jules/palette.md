## 2024-08-24 - [Accessible Icon Buttons]
**Learning:** Icon-only buttons with just a `title` attribute are inconsistently announced by screen readers. Explicit `aria-label` attributes provide reliable accessibility. Additionally, implicitly typed `<button>` elements default to `type="submit"` within forms, which can cause unintended side effects, and failing to set an explicit `type="button"` triggers maintainability violations.
**Action:** Always provide both an explicit `aria-label` and `type="button"` for icon-only buttons to ensure robust screen reader support and maintainability.
