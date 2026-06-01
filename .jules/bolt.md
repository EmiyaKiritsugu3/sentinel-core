## 2026-06-01 - Avoid Strings Allocation

**Learning:** `strings.ToLower()` allocates a new string. Doing this inside a loop in a hot path causes high allocation overhead and slower performance.

**Action:** When searching for text where the variations of capitalization are limited and known, check each known casing using `strings.Contains()` with the original string, rather than lowering the string first. This trades readability for speed by eliminating the string allocation.
