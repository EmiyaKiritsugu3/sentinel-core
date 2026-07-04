## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-07-04 - Optimize static dictionary lookups
**Learning:** O(n²) nested slice iterations for keyword classification are a performance bottleneck that can be optimized to O(n) by using a pre-initialized package-level hash map.
**Action:** Use package-level hash maps initialized in an init() function for static dictionary lookups instead of nested loops.
