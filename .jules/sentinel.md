
## 2024-05-30 - Fix overly permissive CORS configuration
**Vulnerability:** The `internal/liveview/api.go` module had multiple HTTP handlers hardcoding `Access-Control-Allow-Origin: *`.
**Learning:** This open CORS configuration could allow any malicious website to read sensitive application data such as status, code content, and architectural records if a user visited it while the Sentinel liveview server was running locally.
**Prevention:** Always restrict CORS to required origins using secure parsing, particularly when dealing with internal development tools that might expose local file content.
