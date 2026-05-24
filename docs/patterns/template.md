# Pattern Template — Copy and fill

```markdown
### CAT-XXX: Short Descriptive Title

**What happened:** [1-2 sentences describing the bug/surprise/incident.]

**Detection:** [How to spot this before it happens. What code pattern or test failure to look for.]

**Fix:**
[Code example showing the wrong way and the right way.]

**Real example:** [Git commit SHA or PR number from this project where this occurred.]

**Rule:** [One imperative sentence. The action to take.]
```

## Example (filled in)

### SEC-001: Path Traversal via User-Supplied File Paths

**What happened:** API handler passed raw query param to `os.ReadFile`, allowing arbitrary filesystem reads.

**Detection:** `os.Open`, `os.ReadFile`, or `ioutil.ReadFile` with untrusted input.

**Fix:**
```go
// Wrong:
data, _ := os.ReadFile(r.URL.Query().Get("path"))

// Right:
path := filepath.Clean(rawPath)
if filepath.IsAbs(path) || strings.HasPrefix(path, "..") {
    http.Error(w, "invalid path", http.StatusBadRequest)
    return
}
data, _ := os.ReadFile(path)
```

**Real example:** `0f5a69a` — handleGetCode and handleGetADR path traversal hardening.

**Rule:** Every user-supplied file path goes through `filepath.Clean` + traversal rejection BEFORE any file I/O.
```
