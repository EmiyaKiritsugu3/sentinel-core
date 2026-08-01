## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-08-01 - Optimizing O(n*m*k) string manipulation
**Learning:** Found a nested loop where an expensive string manipulation (`strings.Trim`) was performed on the same word for every intent category and every keyword within it, resulting in redundant allocations and CPU cycles.
**Action:** Move invariant operations (like trimming the current word in an outer loop) out of inner loops.
