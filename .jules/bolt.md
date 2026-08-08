## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-05-24 - Avoid regexp.MustCompile in parsing functions
**Learning:** Initializing `regexp.MustCompile` inside high-throughput parsing functions adds significant compilation overhead to every call.
**Action:** When extracting simple patterns (like `[src=...]`), use native string operations (`strings.Index`, `strings.IndexByte`) which execute an order of magnitude faster.
