## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-08-06 - Replace regexp with strings.Index for simple substring extraction
**Learning:** Compiling and running regex for simple start/end token extraction is a major bottleneck in performance and memory, especially on large strings.
**Action:** Use `strings.Index` and manual substring slicing instead of `regexp.MustCompile` and `FindAllStringSubmatch` when extracting bounded strings. It's approximately 10x faster and allocates ~90% less memory.
