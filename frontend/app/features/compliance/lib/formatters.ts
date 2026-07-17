/**
 * Compliance formatters — Phase 1.
 *
 * Pure helpers used by badges, the result panel, and the rule table. They
 * never call out to the network and never invent backend data.
 */
import type { components } from "~/api/ims-api";

import { deviceLocalIsoDate } from "./asOfDate";
import type {
  ComplianceBackendSeverity,
  ComplianceBreach,
  ComplianceBreachStatus,
  ComplianceEffectiveWindow,
  CompliancePortfolioOption,
  ComplianceRule,
  ComplianceRuleCategory,
  ComplianceRuleDerivedStatus,
  ComplianceRuleTypeMetadata,
  ComplianceVerdict,
} from "../types";

type ApiBreach = components["schemas"]["Breach"];
type ApiPortfolio = components["schemas"]["PortfolioResponse"];
type ApiRule = components["schemas"]["RuleInstanceDetail"];
type ApiRuleMetadata = components["schemas"]["RuleMetadata"];

function requiredString(value: string | undefined, field: string): string {
  if (typeof value !== "string" || value.trim() === "") {
    throw new Error(`The API response is missing ${field}.`);
  }
  return value;
}

function requiredNumber(value: number | undefined, field: string): number {
  if (!Number.isFinite(value)) {
    throw new Error(`The API response is missing ${field}.`);
  }
  return value as number;
}

function requiredBoolean(value: boolean | undefined, field: string): boolean {
  if (typeof value !== "boolean") {
    throw new Error(`The API response is missing ${field}.`);
  }
  return value;
}

function asRecord(value: unknown): Record<string, unknown> | undefined {
  return typeof value === "object" && value !== null
    ? { ...(value as Record<string, unknown>) }
    : undefined;
}

function normalizeSeverity(value: string | undefined): ComplianceBackendSeverity {
  switch (value) {
    case "BLOCK":
    case "WARN":
    case "REQUIRE_APPROVAL":
    case "MONITOR":
      return value;
    default:
      throw new Error("The API response contains an unsupported compliance severity.");
  }
}

function normalizeVerdict(value: string | undefined): ComplianceVerdict {
  switch (value) {
    case "PASS":
    case "WARN":
    case "BLOCK":
      return value;
    default:
      throw new Error("The API response contains an unsupported compliance verdict.");
  }
}

function normalizeBreachStatus(value: string | undefined): ComplianceBreachStatus {
  switch (value) {
    case "OPEN":
    case "OVERRIDDEN":
    case "RESOLVED":
      return value;
    default:
      throw new Error("The API response contains an unsupported breach status.");
  }
}

function normalizeCategory(value: string | undefined): ComplianceRuleCategory {
  switch (value) {
    case "REGULATORY":
    case "MANDATE":
    case "HOUSE":
    case "CLIENT":
    case "RESTRICTION":
    case "RATIO":
    case "TEMPORAL":
    case "BEHAVIORAL":
      return value;
    default:
      throw new Error("The API response contains an unsupported rule category.");
  }
}

function normalizeRuleMetadata(raw: ApiRuleMetadata): ComplianceRuleTypeMetadata {
  const supportedTimings = raw.supported_timings ?? [];
  if (!supportedTimings.every((item) => item === "PRE_TRADE" || item === "POST_TRADE" || item === "PERIODIC")) {
    throw new Error("The API response contains an unsupported compliance timing.");
  }

  return {
    type_id: requiredString(raw.type_id, "rule metadata type_id"),
    version: requiredString(raw.version, "rule metadata version"),
    category: normalizeCategory(raw.category),
    default_severity: normalizeSeverity(raw.default_severity),
    supported_timings: supportedTimings,
    supported_scopes: raw.supported_scopes ?? [],
    overridable: requiredBoolean(raw.overridable, "rule metadata overridable"),
    description: raw.description ?? "",
  };
}

function normalizeEffectiveWindow(value: unknown): ComplianceEffectiveWindow | undefined {
  const record = asRecord(value);
  if (!record) return undefined;
  const from =
    typeof record.valid_from === "string"
      ? record.valid_from
      : typeof record.from === "string"
        ? record.from
        : null;
  const to =
    typeof record.valid_to === "string"
      ? record.valid_to
      : typeof record.to === "string"
        ? record.to
        : null;
  return from || to ? { from, to } : undefined;
}

/** Typed boundary from generated OpenAPI rule DTO to the feature's UI model. */
export function normalizeComplianceRule(raw: ApiRule): ComplianceRule {
  return {
    id: requiredString(raw.id, "rule id"),
    ruleTypeID: requiredString(raw.ruleTypeID, "rule type id"),
    name: raw.name ?? "",
    description: raw.description ?? "",
    currentVersion: requiredNumber(raw.currentVersion, "rule version"),
    isActive: requiredBoolean(raw.isActive, "rule active flag"),
    effectiveWindow: normalizeEffectiveWindow(raw.effectiveWindow),
    createdBy: requiredString(raw.createdBy, "rule creator"),
    createdAt: requiredString(raw.createdAt, "rule created time"),
    updatedAt: requiredString(raw.updatedAt, "rule updated time"),
    type_metadata: raw.type_metadata
      ? normalizeRuleMetadata(raw.type_metadata)
      : undefined,
  };
}

/** Typed boundary from generated OpenAPI breach DTO to the feature's UI model. */
export function normalizeComplianceBreach(raw: ApiBreach): ComplianceBreach {
  return {
    id: requiredString(raw.id, "breach id"),
    checkRecordID: requiredString(raw.checkRecordID, "breach check record id"),
    checkGroupID: requiredString(raw.checkGroupID, "breach check group id"),
    portfolioID: requiredString(raw.portfolioID, "breach portfolio id"),
    contractID: raw.contractID,
    ruleTypeID: requiredString(raw.ruleTypeID, "breach rule type id"),
    ruleInstanceID: requiredString(raw.ruleInstanceID, "breach rule instance id"),
    severity: normalizeSeverity(raw.severity),
    verdict: normalizeVerdict(raw.verdict),
    status: normalizeBreachStatus(raw.status),
    evidence: asRecord(raw.evidence),
    message: raw.message ?? "",
    businessDate: requiredString(raw.businessDate, "breach business date"),
    createdAt: requiredString(raw.createdAt, "breach created time"),
    resolvedAt: raw.resolvedAt,
    resolvedBy: raw.resolvedBy,
  };
}

/** Typed boundary from generated portfolio DTO to a code/name UI option. */
export function normalizeCompliancePortfolio(raw: ApiPortfolio): CompliancePortfolioOption {
  return {
    id: requiredString(raw.id, "portfolio id"),
    code: requiredString(raw.code, "portfolio code"),
    name: requiredString(raw.name, "portfolio name"),
    base_currency: requiredString(raw.base_currency, "portfolio base currency"),
    fund_id: requiredString(raw.fund_id, "portfolio fund id"),
    status: raw.status,
  };
}

export type StatusTone = "success" | "warning" | "error" | "info" | "neutral";

export function verdictTone(verdict: ComplianceVerdict): StatusTone {
  switch (verdict) {
    case "PASS":
      return "success";
    case "WARN":
      return "warning";
    case "BLOCK":
      return "error";
    default:
      return "neutral";
  }
}

export function verdictLabel(verdict: ComplianceVerdict): string {
  switch (verdict) {
    case "PASS":
      return "PASS";
    case "WARN":
      return "WARN";
    case "BLOCK":
      return "BLOCK";
    default:
      return verdict;
  }
}

export function severityTone(severity: ComplianceBackendSeverity): StatusTone {
  switch (severity) {
    case "BLOCK":
      return "error";
    case "WARN":
      return "warning";
    case "REQUIRE_APPROVAL":
      return "info";
    case "MONITOR":
      return "neutral";
    default:
      return "neutral";
  }
}

/**
 * Relative strength for ordering/aggregation only — not a backend enum.
 * WARN and REQUIRE_APPROVAL rank equally: per docs/compliance-module.md
 * §8.3 both cap a BLOCK raw verdict to WARN identically; REQUIRE_APPROVAL is
 * only a UI-facing semantic distinction, not a stronger cap.
 */
export function severityRank(severity: ComplianceBackendSeverity | string): number {
  switch (severity) {
    case "BLOCK":
      return 3;
    case "WARN":
    case "REQUIRE_APPROVAL":
      return 2;
    case "MONITOR":
      return 1;
    default:
      return 0;
  }
}

export function severityLabel(severity: ComplianceBackendSeverity): string {
  switch (severity) {
    case "BLOCK":
      return "Blocker";
    case "WARN":
      return "Warning";
    case "REQUIRE_APPROVAL":
      return "Requires approval";
    case "MONITOR":
      return "Monitor only";
    default:
      return severity;
  }
}

function asIsoDate(value: unknown): string | null {
  if (typeof value !== "string" || value.trim() === "") return null;
  return value.length >= 10 ? value.slice(0, 10) : value;
}

function extractWindow(
  raw: ComplianceEffectiveWindow | undefined,
): { from: string | null; to: string | null } {
  if (!raw) return { from: null, to: null };
  const from = asIsoDate((raw as Record<string, unknown>).from);
  const to = asIsoDate((raw as Record<string, unknown>).to);
  return { from, to };
}

/**
 * Map a backend rule instance to one of the four UI-visible derived states.
 * No HTTP — pure derivation. Effective window uses the device-local date,
 * matching the dashboard's explicitly labelled as-of date.
 */
export function deriveRuleStatus(
  rule: ComplianceRule,
  now: Date = new Date(),
): ComplianceRuleDerivedStatus {
  if (!rule.isActive) return "DISABLED";

  const { from, to } = extractWindow(rule.effectiveWindow);
  const today = deviceLocalIsoDate(now);

  if (from && from > today) return "SCHEDULED";
  if (to && to < today) return "EXPIRED";
  return "ACTIVE";
}

export function ruleStatusTone(
  status: ComplianceRuleDerivedStatus,
): StatusTone {
  switch (status) {
    case "ACTIVE":
      return "success";
    case "SCHEDULED":
      return "info";
    case "EXPIRED":
      return "warning";
    case "DISABLED":
      return "neutral";
    default:
      return "neutral";
  }
}

export function ruleStatusLabel(status: ComplianceRuleDerivedStatus): string {
  switch (status) {
    case "ACTIVE":
      return "Active";
    case "SCHEDULED":
      return "Scheduled";
    case "EXPIRED":
      return "Expired";
    case "DISABLED":
      return "Disabled";
    default:
      return status;
  }
}

export function formatEffectiveWindow(
  raw: ComplianceEffectiveWindow | undefined,
): string {
  const { from, to } = extractWindow(raw);
  if (!from && !to) return "—";
  if (from && !to) return `From ${from}`;
  if (!from && to) return `Until ${to}`;
  return `${from} → ${to}`;
}

export function formatIsoDate(value: string | null | undefined): string {
  if (!value) return "—";
  if (value.length >= 10) return value.slice(0, 10);
  return value;
}

export function formatIsoDateTime(value: string | null | undefined): string {
  if (!value) return "—";
  const d = new Date(value);
  if (Number.isNaN(d.getTime())) return value;
  return d.toISOString().replace("T", " ").slice(0, 16) + " UTC";
}

const UUID_RE =
  /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i;

export function isUuid(value: string): boolean {
  return UUID_RE.test(value.trim());
}

export function isPositiveDecimal(value: string): boolean {
  if (!value || value.trim() === "") return false;
  if (!/^\d+(\.\d+)?$/.test(value.trim())) return false;
  return Number.parseFloat(value) > 0;
}
