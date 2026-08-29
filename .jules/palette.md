## 2025-03-04 - Adding ARIA Labels and Type to Icon-Only Buttons
**Learning:** Icon-only buttons (like the 'X' for close or clear) without explicit `aria-label`s are opaque to screen readers, making the interface inaccessible for visually impaired users. Furthermore, missing `type="button"` attributes can cause unintended form submissions or validation failures depending on component context.
**Action:** Always provide descriptive `aria-label`s and explicit `type="button"` attributes on icon-only buttons to ensure full accessibility and prevent unintended default behaviors.
