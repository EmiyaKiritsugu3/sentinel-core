## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-06-06 - Avoid strings.Split for simple delimiter counting
**Learning:** `strings.Split` allocates an array to hold all the resulting substrings, which is computationally and memory expensive. When you only need to count how many items exist separated by a delimiter (e.g., approximating word count by spaces), using `strings.Split` to check `len(...)` is an anti-pattern.
**Action:** Use `strings.Count(str, delimiter)` instead. It achieves the same outcome without allocating any slices, offering a significantly faster and memory-free operation. Note that checking `len(strings.Split(str, " ")) < 3` is equivalent to checking `strings.Count(str, " ") < 2`.
