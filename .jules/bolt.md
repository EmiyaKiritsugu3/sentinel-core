## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-07-01 - Replace nested loops with O(1) hash map lookup for intent keywords
**Learning:** The heuristic intent classification was using O(n²) nested loops to check keywords against a static dictionary. By replacing this with a pre-initialized O(1) package-level hash map lookup, we reduced processing time per word by 10x.
**Action:** Use package-level hash maps initialized in `init()` for fast static dictionary lookups instead of nested iterations.
