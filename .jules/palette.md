## 2024-05-31 - [Added ARIA attributes to UI Components]
**Learning:** React components (EventLog, StatusHUD, InfoPanel, FilterToolbar) were missing critical ARIA attributes for screen reader accessibility, particularly around dynamic state updates and icon-only buttons.
**Action:** Always verify keyboard accessibility and ensure `aria-label`, `role="status"`, and `aria-live` attributes are appropriately set for interactive elements and dynamic updates.
