## 2025-01-20 - Adding explicit type and ARIA labels to buttons
**Learning:** React buttons inside components can unintentionally trigger form submissions if they lack an explicit `type="button"`, and icon-only buttons require `aria-label` for screen reader accessibility.
**Action:** Always include `type="button"` for JS-only actions and `aria-label` on icon-only buttons like "Close" or "Clear".
