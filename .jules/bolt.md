## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2024-08-20 - Prevent allocations in bufio.Scanner loops
**Learning:** Using `scanner.Text()` inside a hot loop (like grepping a file line-by-line) allocates a new string for every line, creating unnecessary garbage collection pressure and CPU overhead.
**Action:** Use `scanner.Bytes()` with `regexp.Match()` instead of `regexp.MatchString(scanner.Text())` to test regular expressions directly against the underlying byte slice without allocating a string.
