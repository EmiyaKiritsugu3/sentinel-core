## 2024-05-24 - Accessibility pattern for icon-only buttons
**Learning:** Icon-only buttons (like 'X' for close or clear) are inaccessible to screen readers without proper labeling. While 'title' provides a tooltip on hover, 'aria-label' is explicitly needed for screen readers.
**Action:** Always add 'aria-label' to buttons that lack descriptive text content to ensure accessibility.
