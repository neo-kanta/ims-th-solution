import {
  assertOpenApiResponse,
  unwrapOpenApiResponse,
  useOpenApiClient,
} from "~/api/openapi";
import type { paths } from "~/api/ims-api";
import type {
  ListItemsResult,
  WatchlistItem,
  ListAlertsResult,
  AlertEvent,
  EvaluateResult,
  CreateItemBody,
  UpdateItemBody,
  AckAlertBody,
  EvaluateBody,
  ItemListQuery,
  AlertListQuery,
} from "../types";

type ItemListParams = NonNullable<paths["/watchlists"]["get"]["parameters"]["query"]>;
type AlertListParams = NonNullable<paths["/watchlists/alerts"]["get"]["parameters"]["query"]>;

export const watchlistApi = {
  async listItems(query: ItemListQuery = {}): Promise<ListItemsResult> {
    const client = useOpenApiClient();
    const params: ItemListParams = {};
    if (query.scope_type) params.scope_type = query.scope_type;
    if (query.portfolio_id) params.portfolio_id = query.portfolio_id;
    if (query.security_id) params.security_id = query.security_id;
    if (query.include_disabled !== undefined) params.include_disabled = query.include_disabled;
    if (query.include_thresholds !== undefined) params.include_thresholds = query.include_thresholds;
    if (query.include_quote !== undefined) params.include_quote = query.include_quote;
    if (query.limit !== undefined) params.limit = query.limit;
    if (query.offset !== undefined) params.offset = query.offset;
    const response = await client.GET("/watchlists", { params: { query: params } });
    return unwrapOpenApiResponse<ListItemsResult>(response);
  },

  async createItem(body: CreateItemBody): Promise<WatchlistItem> {
    const client = useOpenApiClient();
    const response = await client.POST("/watchlists/items", { body });
    return unwrapOpenApiResponse<WatchlistItem>(response);
  },

  async updateItem(id: string, body: UpdateItemBody): Promise<WatchlistItem> {
    const client = useOpenApiClient();
    const response = await client.PATCH("/watchlists/items/{id}", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse<WatchlistItem>(response);
  },

  async deleteItem(id: string): Promise<void> {
    const client = useOpenApiClient();
    const response = await client.DELETE("/watchlists/items/{id}", {
      params: { path: { id } },
    });
    assertOpenApiResponse(response);
  },

  async listAlerts(query: AlertListQuery = {}): Promise<ListAlertsResult> {
    const client = useOpenApiClient();
    const params: AlertListParams = {};
    if (query.scope_type) params.scope_type = query.scope_type;
    if (query.portfolio_id) params.portfolio_id = query.portfolio_id;
    if (query.security_id) params.security_id = query.security_id;
    if (query.rule_id) params.rule_id = query.rule_id;
    if (query.acknowledged !== undefined) params.acknowledged = query.acknowledged;
    if (query.created_from) params.created_from = query.created_from;
    if (query.created_to) params.created_to = query.created_to;
    if (query.limit !== undefined) params.limit = query.limit;
    if (query.offset !== undefined) params.offset = query.offset;
    const response = await client.GET("/watchlists/alerts", { params: { query: params } });
    return unwrapOpenApiResponse<ListAlertsResult>(response);
  },

  async acknowledgeAlert(id: string, body: AckAlertBody = {}): Promise<AlertEvent> {
    const client = useOpenApiClient();
    const response = await client.POST("/watchlists/alerts/{id}/acknowledge", {
      params: { path: { id } },
      body,
    });
    return unwrapOpenApiResponse<AlertEvent>(response);
  },

  async manualEvaluate(body: EvaluateBody = {}): Promise<EvaluateResult> {
    const client = useOpenApiClient();
    const response = await client.POST("/watchlists/evaluate", { body });
    return unwrapOpenApiResponse<EvaluateResult>(response);
  },
};
