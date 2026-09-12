## 2024-09-12 - Missing explicit labels and button types on interactive elements
**Learning:** The application uses icon-only buttons and visually-unlabeled form inputs that lack descriptive accessible names, making them difficult for screen readers to interpret. Furthermore, buttons lacking explicit type="button" attributes could lead to unintended form submissions.
**Action:** Consistently apply aria-label attributes to icon-only buttons and form controls without visible labels, and strictly enforce the addition of type="button" on all non-submit buttons.
