## 2025-02-27 - [Fix Slowloris vulnerability in LiveView Server]
**Vulnerability:** The HTTP server in the liveview module lacked a `ReadHeaderTimeout` configuration. By default, `http.ListenAndServe` provides an HTTP server with unbounded header timeouts, exposing the server to Slowloris resource exhaustion attacks where malicious clients send headers very slowly to tie up server connections.
**Learning:** In Go, never use the default `http.ListenAndServe()` for a production or even local daemon server without bounds on request headers. An unbounded wait allows potential local DoS vectors.
**Prevention:** Always instantiate an explicit `http.Server` structure and specify at least a `ReadHeaderTimeout` (e.g., 5 seconds) to prevent resource exhaustion attacks.
