## 2024-05-14 - Optimize regexp compilation
**Learning:** `regexp.MustCompile` inside a function recompiles the regex on every call. It's a waste of CPU.
**Action:** Always hoist `regexp.MustCompile` to package-level variables so they're only compiled once during initialization.
## 2024-05-14 - Optimize regexp compilation
**Learning:** `regexp.MustCompile` inside a function recompiles the regex on every call. It's a waste of CPU.
**Action:** Always hoist `regexp.MustCompile` to package-level variables so they're only compiled once during initialization.
