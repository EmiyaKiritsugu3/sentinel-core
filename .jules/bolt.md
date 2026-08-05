## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-08-05 - Avoid strings.TrimLeftFunc for simple ASCII whitespace skipping
**Learning:** `strings.TrimLeftFunc(text, unicode.IsSpace)` introduces function call overhead and complex Unicode space logic, which is unnecessary when parsing large text blocks where only basic ASCII whitespace needs to be skipped.
**Action:** Replace `strings.TrimLeftFunc` with a simple inline loop for ASCII whitespace (` `, `\t`, `\n`, `\r`) in critical hot paths to reduce execution time.
