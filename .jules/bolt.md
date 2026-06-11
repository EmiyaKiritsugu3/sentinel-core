## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-23 - Avoid allocations in case-insensitive string matching
**Learning:** Checking string match case-insensitively using `strings.ToLower` on every call causes unnecessary heap allocations. Using `strings.Count(s, sep)` is also faster and avoids allocations compared to `len(strings.Split(s, sep))`.
**Action:** Use fast paths for expected common casings (e.g., `strings.Contains`) while keeping the original slow-path operation as a fallback for edge cases. Always benchmark the 'no-match' scenario to ensure added checks do not cause performance regressions.
