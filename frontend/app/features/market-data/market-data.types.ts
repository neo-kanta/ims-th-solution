/**
 * Market Data — type definitions
 *
 * Backend contracts expected (not yet implemented — see backend module status):
 *   GET    /api/v1/market-data/providers/status
 *   POST   /api/v1/market-data/sync
 *   GET    /api/v1/market-data/securities/search?q=
 *   GET    /api/v1/market-data/watchlist
 *   POST   /api/v1/market-data/watchlist
 *   DELETE /api/v1/market-data/watchlist/:id
 *   POST   /api/v1/market-data/bonds/upload
 *   GET    /api/v1/market-data/sync-activity
 */

export type ProviderId = "alpha-vantage" | "yahoo-finance" | "manual";
export type ProviderStatus = "connected" | "rate-limited" | "idle" | "error";

export interface MarketDataProvider {
  id: ProviderId;
  name: string;
  initials: string;
  endpoint: string;
  status: ProviderStatus;
  usageLabel?: string;
  usageRatio?: number;
  symbolsCount?: number;
  recordsCount?: number;
  lastSync?: string;
  lastUpload?: string;
  errors24h?: number;
  description?: string;
  supportedFiles?: string;
}

export type AssetClass = "Equity" | "Bond" | "ETF" | "FX" | "NVDR";
export type SecurityStatus = "fresh" | "stale" | "failed" | "rate-limited";

export interface SecuritySearchResult {
  id: string;
  symbol: string;
  name: string;
  source: string;
  assetClass: AssetClass;
  rating?: string;
  price: number;
  currency: string;
  change?: number;
  ytm?: number;
  providers: ProviderId[];
  watching: boolean;
}

export interface WatchlistItem {
  id: string;
  symbol: string;
  name: string;
  assetClass: AssetClass;
  source: string;
  providers: ProviderId[];
  price: number;
  currency: string;
  change?: number;
  spark: number[];
  volume?: string;
  ytm?: number;
  lastUpdate: string;
  status: SecurityStatus;
  statusDetail?: string;
  pinned?: boolean;
}

export type SyncSeverity = "info" | "success" | "warning" | "danger";

export interface SyncEvent {
  id: string;
  timestamp: string;
  provider: string;
  message: string;
  duration?: string;
  code?: string;
  severity: SyncSeverity;
}

export interface BondUploadMetadata {
  fileName: string;
  uploadedAt: string;
  rows: number;
  uploadedBy: string;
}

export type WatchlistFilter = "all" | "stocks" | "bonds" | "fx" | "stale" | "pinned";
export type SearchFilter = "all" | "stocks" | "bonds" | "etfs" | "fx" | "sources";

// ---------------------------------------------------------------------------
// Security Detail view model
//
// These DTOs feed the /market-data/securities/[securityId] page. They are a
// frontend aggregation — the backend currently exposes identity + quote +
// history + mappings as separate endpoints, so the composable stitches them
// together. The shapes stay generic so a future backend aggregator endpoint
// (e.g. /market-data/screen/security/{id}) can drop straight in.
// ---------------------------------------------------------------------------

/** Canonical freshness states reported on the detail page. */
export type DetailFreshnessStatus =
  | "FRESH"
  | "STALE"
  | "EXPIRED"
  | "NOT_IMPORTED"
  | "FAILED"
  | "RATE_LIMITED";

/** Resolution status for the security's provider mappings. */
export type DetailMappingStatus =
  | "MAPPED"
  | "UNMAPPED"
  | "REVIEW_REQUIRED"
  | "CONFLICTED";

/** Provider mapping shape projected for the detail view. */
export interface ProviderMappingView {
  mappingId: string;
  providerCode: string;
  providerSymbol: string;
  providerExchange?: string;
  providerAssetType?: string;
  providerCurrency?: string;
  priority?: number;
  status: string;
  isPrimary: boolean;
}

/** Latest quote shape projected for the detail view. Decimal values stay as
 *  strings so the component can format them without losing precision. */
export interface MarketQuoteView {
  provider?: string;
  lastPrice?: string;
  changeAmount?: string;
  changePercent?: string;
  open?: string;
  high?: string;
  low?: string;
  previousClose?: string;
  volume?: number;
  currency?: string;
  asOf?: string;
  capturedAt?: string;
  stale?: boolean;
  staleReason?: string;
}

/** Single point in the price history series. */
export interface MarketPricePoint {
  date: string;
  close: number;
  adjustedClose?: number;
  open?: number;
  high?: number;
  low?: number;
  volume?: number;
}

/** Aggregate data-quality block surfaced on the detail page. */
export interface SecurityDataQuality {
  freshness: DetailFreshnessStatus;
  mapping: DetailMappingStatus;
  providerStatus: string;
  lastSyncResult?: string;
  lastSyncedAt?: string;
  errorCode?: string;
  errorMessage?: string;
  rateLimited?: boolean;
}

/** Sync-activity row. */
export interface SyncActivityItem {
  id: string;
  timestamp: string;
  provider: string;
  operation: string;
  status: "success" | "warning" | "error" | "info";
  durationMs?: number;
  message?: string;
  errorCode?: string;
}

/** The single view model the detail page consumes. */
export interface MarketDataSecurityDetail {
  securityId: string;
  imsSymbol: string;
  displaySymbol: string;
  name: string;
  assetType: string;
  currency?: string;
  countryCode?: string;
  exchangeMic?: string;
  isin?: string;
  figi?: string;
  cusip?: string;
  status: string;
  providerMappings: ProviderMappingView[];
  latestQuote?: MarketQuoteView;
  history: MarketPricePoint[];
  dataQuality: SecurityDataQuality;
  syncActivity: SyncActivityItem[];
  watching: boolean;
  pinned: boolean;
}
