## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-09-12 - Hoisting regexp.MustCompile
**Learning:** Compiling regular expressions inside frequently called functions using `regexp.MustCompile` causes unnecessary recompilation and memory overhead.
**Action:** Always hoist `regexp.MustCompile` calls to package-level variables for single compilation during initialization.
