## 2023-10-27 - Restrictive CORS
**Vulnerability:** Overly permissive CORS configuration (`*`) exposing the API to Cross-Origin Resource Sharing attacks.
**Learning:** Hardcoding `*` for `Access-Control-Allow-Origin` allows any origin to access the API.
**Prevention:** Dynamically parse the `Origin` header and restrict access to specific, allowed origins using the `net/url` package to prevent origin spoofing vulnerabilities.
