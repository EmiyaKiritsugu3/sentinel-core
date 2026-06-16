## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-23 - Optimize Tag Overlap Check
**Learning:** `strings.ToLower` and allocating maps for small slices introduces a heavy performance penalty in Go due to garbage collection overhead. For string matching between small slices, an `O(N^2)` loop using `strings.EqualFold` is faster than the `O(N)` map strategy.
**Action:** When comparing case-insensitive string overlaps between small arrays (length <= 10), implement a fast path using nested loops and `strings.EqualFold` to avoid heap allocations entirely, keeping the map as a slow-path fallback.
