## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2025-03-08 - Hoist Regex Compilation to Package Level
**Learning:** Found `regexp.MustCompile` being called inside frequently executed string manipulation functions (`Slugify` and context extraction). In Go, this causes unnecessary CPU and memory overhead by recompiling the regex on every single call. Benchmark showed a 2-3x speedup (~9000 ns to ~5000 ns, and ~6000 ns to ~2000 ns) just by hoisting.
**Action:** Always hoist `regexp.MustCompile` to package-level variables so they are compiled only once during initialization.
