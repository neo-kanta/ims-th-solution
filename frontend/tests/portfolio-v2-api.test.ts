import { describe, expect, it } from "vitest";
import { v2BaseUrlFrom } from "../app/api/urls";

// Locks the seam that broke during review (twice): useOpenApiClientV2()'s
// base URL concatenated with the generated "/portfolios/{portfolioCode}"
// path key must equal the real backend mount /api/v2/portfolios/{code}
// exactly — not /api/v1/v2/... (Swagger's single basePath bug) and not
// /api/v2/v2/... (the base-URL-swap bug from the first fix attempt).
describe("Portfolio V2 client base URL", () => {
  it("concatenates with the generated path key to the real backend route", () => {
    const v1BaseUrl = "http://localhost:8080/api/v1";
    const pathKey = "/portfolios/{portfolioCode}".replace(
      "{portfolioCode}",
      "TH-EQ-01",
    );

    const requestedUrl = v2BaseUrlFrom(v1BaseUrl) + pathKey;

    expect(requestedUrl).toBe("http://localhost:8080/api/v2/portfolios/TH-EQ-01");
    expect(requestedUrl).not.toContain("/api/v1/");
    expect(requestedUrl).not.toContain("/v2/v2/");
  });

  it("swaps /api/v1 for /api/v2, no double segment", () => {
    const v1BaseUrl = "http://localhost:8080/api/v1";
    expect(v2BaseUrlFrom(v1BaseUrl)).toBe("http://localhost:8080/api/v2");
  });

  it("is a no-op when the configured base URL has no /api/v1 suffix", () => {
    expect(v2BaseUrlFrom("http://localhost:8080")).toBe("http://localhost:8080");
  });
});
