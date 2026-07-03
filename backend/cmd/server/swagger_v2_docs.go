// Portfolio V2 (docs/api/portfolio-v2-api-ddd.md) Swagger general API info.
//
// Swagger 2.0 only supports one basePath per generated spec, and V2 routes
// are mounted at /api/v2 — a different basePath than the rest of the API's
// /api/v1 (see main.go's own @basePath). Folding a V2 @Router path into the
// main spec would document it under the wrong basePath (/api/v1/...
// instead of /api/v2/...), so V2 gets its own spec generated from this
// file. See the Makefile's `swagger-v2` target (uses --tags to include only
// V2-tagged operations) and the `swagger` target's `--tags` exclusion of
// the same tag, so the two specs never double up an operation.
//
// @title           IMS Thailand API — Portfolio V2
// @version         2.0.0
// @description     Portfolio V2 endpoints (docs/api/portfolio-v2-api-ddd.md), mounted under /api/v2.
// @host            localhost:8080
// @basePath        /api/v2
// @schemes         http https
//
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description JWT Bearer Token. Format: "Bearer <token>"
package main
