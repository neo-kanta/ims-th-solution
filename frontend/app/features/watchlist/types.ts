import type { components } from "~/api/ims-api";

export type WatchlistItem = components["schemas"]["WatchlistItemResponse"];
export type AlertEvent = components["schemas"]["AlertEventResponse"];
export type ListItemsResult = components["schemas"]["ListItemsResponseData"];
export type ListAlertsResult = components["schemas"]["ListAlertsResponseData"];
export type ThresholdRule = components["schemas"]["ThresholdRuleResponse"];
export type SecurityDescriptor = components["schemas"]["SecurityDescriptor"];
export type PortfolioDescriptor = components["schemas"]["PortfolioDescriptor"];
export type QuoteSnapshot = components["schemas"]["QuoteSnapshot"];
export type PaginationMeta = components["schemas"]["PaginationMeta"];
export type CreateItemBody = components["schemas"]["CreateItemRequest"];
export type UpdateItemBody = components["schemas"]["UpdateItemRequest"];
export type AckAlertBody = components["schemas"]["AcknowledgeAlertRequest"];
export type EvaluateBody = components["schemas"]["EvaluateRequest"];
export type EvaluateResult = components["schemas"]["EvaluateResponseData"];
export type ThresholdRuleInput = components["schemas"]["ThresholdRuleRequest"];

export interface ItemListQuery {
  scope_type?: "PERSONAL" | "PORTFOLIO";
  portfolio_id?: string;
  security_id?: string;
  include_disabled?: boolean;
  include_thresholds?: boolean;
  include_quote?: boolean;
  limit?: number;
  offset?: number;
}

export interface AlertListQuery {
  scope_type?: "PERSONAL" | "PORTFOLIO";
  portfolio_id?: string;
  security_id?: string;
  rule_id?: string;
  acknowledged?: boolean;
  created_from?: string;
  created_to?: string;
  limit?: number;
  offset?: number;
}
