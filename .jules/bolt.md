## 2026-05-23 - Use scanner.Bytes() for hot loop regex matching
**Learning:** Using `regexp.MatchString(scanner.Text())` in a hot loop (like reading files line by line) creates unnecessary string allocations on every iteration.
**Action:** Use `regexp.Match(scanner.Bytes())` instead to work directly with the byte slice and prevent memory overhead, improving performance particularly on large files.
