## 2025-02-28 - Compile regexes outside of loop
**Learning:** `regexp.MustCompile` is slow, so it is better to hoist compiling regex out of frequently called functions/loops so it's only done once during initialization.
**Action:** Replace inline `regexp.MustCompile` with package-level precompiled regexes to improve performance.
