## 2024-05-18 - Improve Accessibility for Form Inputs and Icon-only Buttons
**Learning:** Inputs (`<input>`, `<select>`) without visible labels and icon-only buttons rely heavily on explicit `aria-label` and `type="button"` attributes to be screen-reader friendly and prevent unintended form submissions in this application's custom components.
**Action:** Always verify that interactive elements like inputs, selects, and icon buttons include appropriate accessibility attributes (like `aria-label` and `type="button"`) when designing or modifying UI components.
