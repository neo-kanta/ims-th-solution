# System and API Discovery Endpoints

**Status:** Current implementation as of 2026-07-20

These endpoints are mounted outside `/api/v1` and `/api/v2`. They do not use Bearer authentication in the current router. Operators must restrict network access to `/metrics` at the deployment boundary.

## Endpoint summary

| Method | Path | Purpose | Authentication |
| --- | --- | --- | --- |
| GET | `/health` | Backend, PostgreSQL, Redis, and chat/MCP readiness | Public |
| GET | `/metrics` | Prometheus metrics scrape | Public |
| GET | `/swagger/index.html` | V1 Swagger UI | Public |
| GET | `/swagger/doc.json` | V1 Swagger 2.0 document | Public |
| GET | `/swagger/v2/index.html` | Portfolio V2 Swagger UI | Public |
| GET | `/swagger/v2/doc.json` | Portfolio V2 Swagger 2.0 document | Public |

## GET `/health`

Check backend, database, and Redis health status

| HTTP status | Description | Schema |
| --- | --- | --- |
| 200 | OK | HealthResponse |
| 503 | Service Unavailable | HealthResponse |

The health payload contains component status only; it must not expose credentials, provider keys, model identifiers, or private endpoints.

## GET `/metrics`

Returns Prometheus exposition text from the shared metrics registry. The route is intentionally unauthenticated for scraping, so expose it only on a trusted network or behind infrastructure access controls.

## OpenAPI caveat

The V1 Swagger source contains a `/health` operation while declaring `basePath: /api/v1`. The live route is `/health`, not `/api/v1/health`. `/metrics` and the Swagger UI routes are not part of the API base path.

## Source references

- `backend/cmd/server/main.go`
- `backend/cmd/server/swagger_v2_docs.go`
- `backend/platform/health/health.go`
- `backend/platform/metrics/registry.go`
- `backend/docs/swagger.json`
- `backend/docs/v2/v2_swagger.json`
