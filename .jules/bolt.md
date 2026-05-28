## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-24 - React List Rendering with Prepended Items
**Learning:** Using array indices (or timestamp-index combinations) as React `key`s for lists where items are prepended causes O(n) re-renders, as all existing items get new keys and are torn down/recreated.
**Action:** Always assign a stable, unique identifier (e.g. an incrementing counter) to list items on creation and use `React.memo` to ensure O(1) DOM operations when prepending new items.
