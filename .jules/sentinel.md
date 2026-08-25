## 2024-05-24 - [Remove Wildcard CORS from API Endpoints]
**Vulnerability:** The API endpoints were using `Access-Control-Allow-Origin: *`, which is an overly permissive CORS configuration. This could allow any origin to access the API endpoints, potentially leading to unauthorized data access.
**Learning:** Hardcoded wildcard CORS is a common vulnerability pattern in development environments that can accidentally slip into production. It should be replaced with explicit origin validation.
**Prevention:** Always validate the `Origin` header dynamically using `url.Parse` and verify the hostname against a whitelist of trusted origins, instead of using wildcards or insecure string prefix checks.
