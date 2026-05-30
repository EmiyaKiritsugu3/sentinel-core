## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-30 - Avoid recompiling regexes inside frequently called functions
**Learning:** Functions like `extractDocuments` and `extractConcepts` in `internal/context/service.go` were compiling the regex on each function call. Compiling regexes in Go (`regexp.MustCompile`) is relatively expensive.
**Action:** Always pre-compile static regular expressions to package-level variables using `regexp.MustCompile`. Also pre-allocate maps using the size of slices when iterating and inserting elements if possible to avoid reallocation.
