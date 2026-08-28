## 2025-03-05 - Add ARIA Labels to Icon-Only Close Buttons
**Learning:** Icon-only buttons using text characters like 'X' or '✕' for visual close/clear actions lack semantic meaning for screen readers, as the title attribute alone is insufficient for robust accessibility across all assistive technologies.
**Action:** Always add an explicit aria-label and type='button' to icon-only interactive elements to ensure clear intent for assistive tech and prevent unintended form submissions.
