## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2025-02-27 - Hoist regexp.MustCompile
**Learning:** Compiling regexes using `regexp.MustCompile` inside frequently called functions adds unnecessary overhead. Hoisting them to package-level variables ensures they are compiled only once, reducing CPU time significantly (e.g., from ~5250 ns/op to ~1428 ns/op).
**Action:** Always declare `regexp.MustCompile` as a package-level variable to optimize performance in loops or frequently executed functions.
