## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2026-07-16 - Avoid intermediate slice allocations with `strings.Count`
**Learning:** Using `strings.Split(s, sep)` generates an intermediate slice that creates unnecessary heap allocations and GC pressure, especially on heavily loaded paths parsing CSV strings.
**Action:** Replace `strings.Split` with `strings.Count(s, sep) + 1` to exactly pre-allocate the capacity of the final slice and iterate using a single index loop (`s[i] == sep`) or `strings.IndexByte`, extracting substrings manually to skip the intermediate slice allocation.
