## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2024-05-18 - [Optimize Regexp Compilations]
**Learning:** Compiling regular expressions inside functions dynamically adds significant overhead in Go apps handling string processing. Pre-compiling them at the package level as global variables fixes this overhead since `*regexp.Regexp` objects are safe for concurrent use.
**Action:** When working on string operations in Go, ensure all regular expressions are defined and compiled outside of functions.
