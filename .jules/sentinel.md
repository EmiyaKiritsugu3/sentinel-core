## 2024-05-27 - [os.ReadFile Memory Exhaustion]
**Vulnerability:** Found `os.ReadFile` used directly on user-provided file paths which loads the entire file into memory at once.
**Learning:** This is a memory exhaustion/DoS risk (and a violation of internal standard STD-01), especially if dealing with large, minified files or unconstrained payload sizes. It existed because it's a common convenience function used before considering memory footprints.
**Prevention:** Ensure file reads use streaming constructs like `os.Open` with `bufio.Scanner` or `bufio.Reader` instead of reading the entire file contents directly into a single string.
## 2024-05-27 - [os.ReadFile Memory Exhaustion - Fix Reversion]
**Vulnerability:** CodeQL flagged `os.Open` in `internal/liveview/api.go` as using user-provided values.
**Learning:** `filepath.Clean()` does not resolve or block relative paths if it doesn't start with `..`. The path traversal check `strings.HasPrefix(cleanPath, "..")` must be applied, but in `api.go`, `filepath.IsAbs` was checking the wrong variable because of a formatting typo.
**Prevention:** Ensure that `filepath.Clean` and `filepath.IsAbs` operate properly, and always run standard testing to ensure fixes don't reduce SonarCloud coverage. Writing new tests to cover the replaced functions ensures coverage thresholds remain >80% on new code.
