## 2025-02-28 - Avoid compiling regex inside functions
**Learning:** Compiling regular expressions using `regexp.MustCompile` inside frequently called functions causes unnecessary CPU and memory overhead because the regex is recompiled on every function call.
**Action:** Always hoist `regexp.MustCompile` to package-level variables so they are compiled only once during initialization.
