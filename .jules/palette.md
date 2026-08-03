## 2024-05-18 - Missing ARIA labels on Icon Buttons and Forms
**Learning:** Found multiple instances where icon-only buttons (like ✕) and core form controls (like search inputs and dropdowns) in standard utility components lack accessible names, making them difficult for screen reader users to identify.
**Action:** Always verify that buttons lacking visible text and standalone inputs/selects have explicit `aria-label` attributes to provide necessary context for assistive technologies.
