## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2025-02-18 - Optimize Context Parsing
**Learning:** Compiling and matching regular expressions in Go (like `regexp.MustCompile`) inside loops or on large dynamically retrieved strings can be a major performance bottleneck. For simple prefix/suffix extractions, standard library functions like `strings.Index` and `strings.TrimSpace` are significantly faster.
**Action:** Replace `regexp` with `strings` functions for simple string extractions to improve speed and reduce allocations.
