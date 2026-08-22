## 2026-08-22 - [Fix CORS bypass vulnerability]
**Vulnerability:** The CORS configuration on API endpoints allowed any origin `*`, exposing the API to cross-origin requests.
**Learning:** Hardcoding wildcard allowed origins can leave APIs vulnerable. It is vital to validate incoming origins.
**Prevention:** Instead of wildcard `*`, parse incoming `Origin` headers, and safely check hostnames before allowing requests.
