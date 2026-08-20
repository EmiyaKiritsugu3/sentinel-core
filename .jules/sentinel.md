## 2026-05-23 - Mitigate Slowloris in Default HTTP Server
**Vulnerability:** The default `http.ListenAndServe()` function lacks a `ReadHeaderTimeout`, leaving the server vulnerable to Slowloris resource exhaustion attacks where malicious clients send headers very slowly to tie up server connections.
**Learning:** In Go HTTP applications, default helper methods prioritize simplicity over security. They do not enforce timeouts, which allows a single slow client to exhaust the server's connection pool.
**Prevention:** Always construct a custom `http.Server` struct with explicitly defined timeouts, specifically `ReadHeaderTimeout`, instead of using `http.ListenAndServe()` or `http.ListenAndServeTLS()`.
