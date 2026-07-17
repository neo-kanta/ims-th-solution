/**
 * Pure URL helpers for the API client layer. No Nuxt imports here on
 * purpose — `#imports`/`~` auto-imports aren't resolvable outside a Nuxt
 * build context, so anything that needs plain-vitest coverage (no
 * @nuxt/test-utils in this repo) has to live in a file like this one.
 */

/**
 * Derives the Portfolio V2 (docs/api/portfolio-v2-api-ddd.md) client base
 * URL from the configured V1 base URL.
 *
 * V2 is mounted under `/api/v2` alongside V1 under `/api/v1`. Swagger 2.0
 * only supports one `basePath` per spec, so V2 routes are documented as
 * their own spec with `basePath: /api/v2` (see
 * backend/cmd/server/swagger_v2_docs.go), separate from the main spec's
 * `/api/v1`. Both specs' path keys are relative to their own basePath —
 * V2's generated key is `"/portfolios/{portfolioCode}"`, no `/v2` prefix
 * baked in — so this must resolve to `/api/v2` (the same relationship
 * `useOpenApiClient()` has with `/api/v1`), not the bare `/api` root.
 */
export function v2BaseUrlFrom(v1BaseUrl: string): string {
  return v1BaseUrl.replace(/\/api\/v1$/, "/api/v2");
}
