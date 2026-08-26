## 2026-08-26 - Restrict CORS Configuration
**Vulnerability:** Liveview HTTP endpoints had overly permissive `Access-Control-Allow-Origin: *` headers, which could lead to origin spoofing vulnerabilities.
**Learning:** `*` origin bypasses browser SOP restrictions for fetching resources from other origins. Even in local environments it can be abused.
**Prevention:** Use dynamic CORS validation, checking the request's origin against a whitelist of trusted hostnames (e.g., `localhost`, `127.0.0.1`) using `net/url`.
