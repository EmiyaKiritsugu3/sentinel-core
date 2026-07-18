## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-07-18 - Optimizing Heuristic Classifier Search Algorithm
**Learning:** O(n²) nested loop iterations over slices inside string field loops can severely bottleneck performance, particularly for frequently-called classification algorithms.
**Action:** Replace nested loops iterating through static arrays with O(n) pre-computed hash map lookups. If a keyword maps to multiple categories, use `map[string][]Intent` to preserve semantics instead of `map[string]Intent` which would overwrite values and cause regressions.
