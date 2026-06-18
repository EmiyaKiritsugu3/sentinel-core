## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-23 - Optimize delimited string parsing by avoiding strings.Split
**Learning:** `strings.Split` allocates an intermediate slice containing all separated parts. When building a filtered or modified list of substrings (like trimming spaces from tags), this results in double slice allocation.
**Action:** To optimize Go performance and reduce intermediate string slice allocations when processing delimited strings, avoid using `strings.Split`. Instead, use `strings.Count(s, sep)` to accurately pre-allocate the capacity of the result slice, then manually iterate over the string to slice and extract the substrings, while ensuring you retain any necessary behavior like `strings.TrimSpace`.
