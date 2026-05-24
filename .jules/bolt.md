## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-24 - Pre-parsing strings in hot loops
**Learning:** `filepath.Base` and string manipulation functions (`strings.TrimPrefix`, `strings.Contains`, `strings.ToLower`) are surprisingly expensive when called repeatedly in hot loops, like the `IsIgnored` filter which runs for every file in the directory tree.
**Action:** When a function validates files iteratively against a static set of rules (like gitignore patterns or internal skip lists), parse the rules once on struct initialization instead of performing string manipulation dynamically inside the loop.
