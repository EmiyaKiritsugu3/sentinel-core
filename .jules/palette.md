## 2024-05-18 - Missing ARIA labels on Icon Buttons and Forms
**Learning:** Found multiple instances where icon-only buttons (like ✕) and core form controls (like search inputs and dropdowns) in standard utility components lack accessible names, making them difficult for screen reader users to identify.
**Action:** Always verify that buttons lacking visible text and standalone inputs/selects have explicit `aria-label` attributes to provide necessary context for assistive technologies.

## 2024-05-18 - Missing type attribute on buttons causing SonarCloud failures
**Learning:** Found that when modifying existing `<button>` elements to improve accessibility, if the button does not have an explicit `type` attribute, SonarCloud flags the modified line as "New Code" and fails the maintainability Quality Gate.
**Action:** Always verify that `<button>` elements have an explicit `type` attribute (e.g., `type="button"`) when making UI modifications to prevent unexpected CI failures related to maintainability rules.
