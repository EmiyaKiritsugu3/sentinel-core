## 2024-05-23 - Added missing ARIA labels to icon-only buttons
**Learning:** Found multiple instances where we rely on icon-only buttons (`✕` and `X`) for closing panels and clearing logs. These used `title` attributes but lacked `aria-label` for proper screen reader announcement.
**Action:** Always verify icon-only buttons have an `aria-label` attribute alongside `title` for improved accessibility.
