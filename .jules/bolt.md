## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2024-05-27 - Golang Regex and Replacer Pre-computation
**Learning:** Found frequent dynamic allocations of `strings.NewReplacer` and runtime compilation of `regexp.MustCompile` inside commonly used functions (`SanitizeID` and `Slugify` in `pkg/utils/text.go`). These lead to significant CPU overhead when processing many files/requests.
**Action:** Extract `strings.NewReplacer` and `regexp.MustCompile` into package-level global variables when the patterns/replacements are static. This drastically improves throughput (`SanitizeID` speed increased ~5x).
