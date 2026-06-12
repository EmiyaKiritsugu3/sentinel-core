## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-23 - Optimize case-insensitive overlap for short string slices
**Learning:** When comparing short string slices (like tags) for overlap, allocating a map and using `strings.ToLower` on each element incurs significant heap allocation overhead. Even though a nested loop with `strings.EqualFold` is theoretically O(N*M), it avoids these heap allocations and is substantially faster in practice for small N and M.
**Action:** Use nested loops with `strings.EqualFold` instead of map allocations and `strings.ToLower` for case-insensitive overlap or matching checks when dealing with very short string slices.
