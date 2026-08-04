## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2024-08-04 - Precompile regexp dynamically inside functions
**Learning:** In Go, avoiding compiling regular expressions dynamically inside functions using `regexp.MustCompile` avoids CPU overhead, as they can be reused across calls.
**Action:** When a regular expression is static, always pre-compile it at the package level as a global variable. Compiled `*regexp.Regexp` objects are goroutine-safe.
