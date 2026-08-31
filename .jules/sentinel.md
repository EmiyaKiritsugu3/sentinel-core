## 2023-10-27 - Wildcard CORS Vulnerability
**Vulnerability:** The API endpoints used `w.Header().Set("Access-Control-Allow-Origin", "*")`, which allows any website to make cross-origin requests to the API and read its response.
**Learning:** Hardcoding wildcard CORS headers is dangerous.
**Prevention:** Implement strict origin validation by parsing the `Origin` header and checking against a whitelist (e.g., `localhost`, `127.0.0.1`) using `net/url`.
