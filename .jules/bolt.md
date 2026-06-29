## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-06-29 - strings.Split Allocations
**Learning:** `strings.Split` on large command outputs causes significant memory allocation overhead.
**Action:** Use an allocation-free manual iteration over the string using `strings.IndexByte(s, '\n')` for counting lines and extracting content to avoid allocating a massive array of substrings.
