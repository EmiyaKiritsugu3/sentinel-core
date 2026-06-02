## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-06-02 - Avoid `strings.Split` just to count substrings
**Learning:** `strings.Split` allocates a new string slice which puts unnecessary pressure on the garbage collector if the only goal is to count the number of elements (e.g., words in a string).
**Action:** When you just need to know the number of occurrences of a substring or if it's less than a certain threshold, use `strings.Count`. For example, replace `len(strings.Split(s, " ")) < 3` with `strings.Count(s, " ") < 2` to eliminate allocations while maintaining identical logic.
