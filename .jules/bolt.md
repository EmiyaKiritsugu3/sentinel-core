## 2026-05-23 - Avoid strings.TrimSpace on unbounded text chunks
**Learning:** `strings.TrimSpace` evaluates both the beginning and the end of a string. When parsing large agent output chunks to check if they start with a thought block prefix (e.g. `<think>`), this causes an unnecessary `O(N)` traversal of potentially massive trailing content (actions, logs, etc.) just to check the prefix.
**Action:** When validating string prefixes with potential leading whitespace, manually scan and skip the leading whitespace using a fast loop rather than calling `strings.TrimSpace`, especially when the string can be unbounded in length.

## 2026-08-03 - [Optimize isExplicitThoughtBlock text parsing]
**Learning:** `strings.TrimLeftFunc` with `unicode.IsSpace` in `isExplicitThoughtBlock` is currently O(N) leading whitespace scan and requires UTF-8 decoding overhead. For simple structural block parsing where we only care about ASCII spaces/newlines before `<think>` or ````thought`, a direct byte-level loop parsing avoids `TrimLeftFunc` and string prefix allocations, accelerating checking by over 4x on cold strings (from ~27ns to ~6ns). This is highly called when parsing agent stream blocks.
**Action:** Replace `strings.TrimLeftFunc` with byte loop iteration for whitespace skipping in `isExplicitThoughtBlock`.
