## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-07-17 - Optimize repeated keyword matching with static map
**Learning:** In Go, checking if a string matches against a static 2D dictionary (e.g., iterating through a map of categories to string slices) using nested `for` loops results in O(n²) time complexity. This becomes a bottleneck if the dictionary is large or the check is called frequently (like in intent classification).
**Action:** For static keyword dictionaries, precompute a reverse-lookup hash map (mapping strings to categories) in the package's `init()` function. This reduces runtime complexity from O(n²) to an O(1) map lookup. If a keyword can map to multiple categories, use a slice value (`map[string][]Category`) to preserve all semantics and avoid regressions.
