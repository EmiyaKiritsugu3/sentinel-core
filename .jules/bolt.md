## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-08-29 - Hoist regexp.MustCompile to package level
**Learning:** In Go, calling `regexp.MustCompile` inside a frequently called function is an anti-pattern as the regular expression gets re-compiled every time the function is called, which consumes CPU and memory.
**Action:** Always hoist regular expressions to package-level variables to ensure they are compiled only once during initialization, preventing unnecessary CPU and memory overhead.
