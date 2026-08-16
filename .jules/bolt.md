## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2025-02-20 - Optimize Slugify by avoiding regex compilation

**Learning:** Using `regexp.MustCompile` inside frequently called functions is a performance bottleneck. The overhead of instantiating and matching multiple regex patterns, along with chained `strings.ReplaceAll` calls, results in numerous memory allocations and string copies.
**Action:** Replaced regex and string replacements in the `Slugify` function with a custom, single-pass `strings.Builder` approach. It processes characters sequentially, applying lowercase conversion and character replacement inline. This avoids multiple allocations, dropping latency from ~9163 ns/op to ~271 ns/op (a ~33x speedup).
