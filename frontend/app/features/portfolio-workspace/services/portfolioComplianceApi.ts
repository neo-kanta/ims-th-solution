/**
 * Typed wrapper for Portfolio Compliance V2 routes
 * (docs/compliance-module.md — "Portfolio Compliance V2"). Every call goes
 * through the generated OpenAPI client — see CLAUDE.md for the
 * no-manual-API-clients rule.
 *
 * V2 identifies a portfolio by its business `code` in the URL. Clients never
 * send portfolio_id, fund_id, or contract_id — the backend resolves those
 * from the code and the portfolio record.
 *
 * Endpoints used:
 *   GET    /api/v2/portfolios/{portfolioCode}/compliance/rules
 *   POST   /api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings
 *   DELETE /api/v2/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}
 *   GET    /api/v2/portfolios/{portfolioCode}/compliance/breaches
 */
import { assertOpenApiResponse, unwrapOpenApiResponse, useOpenApiClientV2 } from "~/api/openapi";
import type { components } from "~/api/ims-api";

export type ApiPortfolioRuleCatalogEntry =
  components["schemas"]["PortfolioRuleCatalogEntry"];
export type ApiPortfolioRuleBindingView =
  components["schemas"]["PortfolioRuleBindingView"];
export type ApiPortfolioBreachView = components["schemas"]["PortfolioBreachView"];
export type ApiBindRuleRequest = components["schemas"]["BindRuleRequest"];

export const portfolioComplianceApi = {
  /** Active rule catalog, annotated with this portfolio's binding when one exists. */
  async listRules(
    portfolioCode: string,
  ): Promise<ApiPortfolioRuleCatalogEntry[]> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/compliance/rules",
      { params: { path: { portfolioCode } } },
    );
    return unwrapOpenApiResponse<ApiPortfolioRuleCatalogEntry[]>(response);
  },

  /** Bind a rule instance to this portfolio (scope_type=PORTFOLIO). */
  async bindRule(
    portfolioCode: string,
    ruleInstanceID: string,
    payload: ApiBindRuleRequest,
  ): Promise<ApiPortfolioRuleBindingView> {
    const client = useOpenApiClientV2();
    const response = await client.POST(
      "/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings",
      { params: { path: { portfolioCode, ruleInstanceID } }, body: payload },
    );
    return unwrapOpenApiResponse<ApiPortfolioRuleBindingView>(response);
  },

  /** Deactivate an existing binding on this portfolio. */
  async deactivateBinding(
    portfolioCode: string,
    ruleInstanceID: string,
    bindingID: string,
  ): Promise<void> {
    const client = useOpenApiClientV2();
    const response = await client.DELETE(
      "/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}",
      { params: { path: { portfolioCode, ruleInstanceID, bindingID } } },
    );
    assertOpenApiResponse(response);
  },

  /** Compliance breaches recorded against this portfolio. */
  async listBreaches(portfolioCode: string): Promise<ApiPortfolioBreachView[]> {
    const client = useOpenApiClientV2();
    const response = await client.GET(
      "/portfolios/{portfolioCode}/compliance/breaches",
      { params: { path: { portfolioCode } } },
    );
    return unwrapOpenApiResponse<ApiPortfolioBreachView[]>(response);
  },
};
