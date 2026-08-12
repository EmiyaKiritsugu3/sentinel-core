## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2024-05-24 - React rendering optimization for event logs
**Learning:** In the Sentinel EventLog, new events are appended frequently. Without `React.memo`, appending a single new event causes every existing event row to unnecessarily re-render, leading to O(n²) rendering complexity as the log grows.
**Action:** Always apply `React.memo()` to list items when the list is frequently appended to and the item props are immutable.
