## 2025-02-14 - Hoist regexp.MustCompile to package level
**Learning:** Compiling regular expressions using `regexp.MustCompile` inside frequently called functions introduces unnecessary CPU and memory overhead on every call.
**Action:** Always hoist `regexp.MustCompile` statements to package-level variables so they are compiled only once during initialization.

## 2025-02-14 - Hoist regexp.MustCompile to package level
**Learning:** Compiling regular expressions using `regexp.MustCompile` inside frequently called functions introduces unnecessary CPU and memory overhead on every call.
**Action:** Always hoist `regexp.MustCompile` statements to package-level variables so they are compiled only once during initialization.
