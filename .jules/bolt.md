## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.
## 2024-07-24 - Do not overwrite test files with benchmark explorations
**Learning:** Overwriting tracked test files (e.g. `service_test.go`) during local benchmarking causes massive codebase regressions when those files are subsequently deleted during cleanup, breaking the test suite completely.
**Action:** Always use distinct, temporary filenames (e.g. `tmp_bench_test.go`) when creating temporary exploration scripts and benchmark files to prevent colliding with and destroying tracked codebase files.
