## 2024-05-24 - Cross-Origin Resource Sharing (CORS) Overly Permissive
**Vulnerability:** LiveView API endpoints use `w.Header().Set("Access-Control-Allow-Origin", "*")`.
**Learning:** Using a wildcard `*` for CORS `Access-Control-Allow-Origin` allows any origin to read the response, which could expose local telemetry, source code, and architectural data to malicious websites.
**Prevention:** Explicitly restrict CORS to trusted development origins, such as `http://localhost:5173`.
