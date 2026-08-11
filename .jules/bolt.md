## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-05-23 - Memoize list items in React
**Learning:** In React, appending to an array used to render a list without memoizing the individual list items causes all items to unnecessarily re-render on every new addition.
**Action:** When rendering lists of immutable objects, especially logs or event streams, always wrap the item component in `React.memo()` to skip rendering for items whose props have not changed.
