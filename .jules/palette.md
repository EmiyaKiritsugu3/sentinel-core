## 2024-08-20 - Adding ARIA labels to Icon-only Buttons
**Learning:** Icon-only buttons (like the close/clear buttons) lack screen reader context when relying solely on title attributes, and modifying them without explicit type attributes triggers maintainability warnings.
**Action:** Added explicit aria-label and type="button" attributes to icon-only buttons to ensure aural context for screen readers and semantic maintainability.
