## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2025-03-01 - Avoid string allocation in scanner loops
**Learning:** Using `regexp.MatchString(scanner.Text())` inside a `bufio.Scanner` loop causes a new string allocation for every single line scanned, which is very inefficient for large files or grep-like operations.
**Action:** Use `regexp.Match(scanner.Bytes())` to test the regex against the byte slice directly, which avoids the allocation.
