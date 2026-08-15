## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2025-03-09 - Optimize regex compilation in tight loops
**Learning:** Compiling regular expressions using `regexp.MustCompile` inside functions called frequently creates unnecessary overhead, as the regex is recompiled on every invocation.
**Action:** Always extract `regexp.MustCompile` to package-level variables so they are compiled only once during application startup.
