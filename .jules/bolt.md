## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-07-12 - Replace O(n^2) keyword match with O(n) map lookup
**Learning:** In Go, to optimize intent classification or keyword matching against a static dictionary, replace O(n²) nested loop iterations over slices with O(n) lookups using a package-level hash map pre-initialized in the `init()` function.
**Action:** If a keyword can map to multiple categories, use a slice value (e.g., `map[string][]Intent`) rather than a single value (`map[string]Intent`) to preserve original semantics and avoid non-deterministic mapping regressions.
