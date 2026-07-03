## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-23 - Optimize strings.Split on hot paths
**Learning:** `strings.Split` allocates an intermediate slice of strings to hold all the split components before iterating over them, which is inefficient for memory and performance, especially in hot paths (like parsing tags in deduplication checks).
**Action:** Replace `strings.Split` with manual iteration using `strings.IndexByte` and pre-allocate the required slice capacity using `strings.Count(s, sep) + 1` to skip the intermediate slice allocation entirely. This drops allocations and saves execution time significantly.
