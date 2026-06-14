## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-05-23 - Optimize case-insensitive match on small slices
**Learning:** For checking case-insensitive overlap or matches between very short string slices (e.g., tags), allocating a map and running `strings.ToLower` across elements creates measurable heap allocations and GC pressure.
**Action:** When matching elements between small string slices, use an allocation-free nested loop with `strings.EqualFold` instead of allocating an intermediate map and lowercased string copies.

## 2026-05-23 - Fast-path literal containment over ToLower matching
**Learning:** `strings.ToLower` copies the entire string to a new allocation. When searching for a specific keyword in an unbounded string, directly using `strings.Contains` for expected cases (e.g., lowercase and capitalized variations) serves as a rapid fast-path check that avoids allocations entirely in the hot path.
**Action:** Before falling back to a full string transformation like `strings.ToLower` for case-insensitive checks, prepend checks using `strings.Contains` for the most common specific casing variants. Ensure the slower `strings.ToLower` path is retained for edge-case correctness.

## 2026-05-23 - Count occurrences to approximate split length without slice allocations
**Learning:** `strings.Split` splits a string by a delimiter into a new string slice, resulting in multiple allocations on the heap (for the slice backing array and each separated string segment). This is extremely wasteful when the only purpose is checking the count of segments.
**Action:** When approximating the word count or segment count, use `strings.Count(s, sep)` instead. It achieves the exact same math logically (as length of split elements `n` correlates strictly to `n-1` occurrences of a single delimiter) without making any heap allocations.
