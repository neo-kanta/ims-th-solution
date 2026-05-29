import type {
  ResearchRecommendation,
  ResearchReportStatus,
  ResearchReviewStatus,
} from "../types";

type BadgeVariant =
  | "success"
  | "warning"
  | "error"
  | "info"
  | "neutral"
  | "draft"
  | "purple"
  | "orange"
  | "teal";

/**
 * Field length budgets — kept in sync with the backend DTO and the
 * `investment__research_reports` Postgres column widths. If the backend
 * widens a column, bump the number here as well.
 *
 * The backend currently does NOT enforce these via a validator (see review
 * finding #3), so the frontend enforces them defensively. Overrun is hard-
 * stopped at the form layer to avoid a 500 from a Postgres CHECK / length
 * violation.
 */
export const RESEARCH_FIELD_LIMITS = {
  reportNo: 60,
  instrumentCode: 40,
  instrumentName: 255,
  instrumentType: 40,
  market: 40,
  reportTitle: 255,
  currency: 3,
} as const;

/** Lower bound on the analyst's substantive content. Matches backend policy. */
export const MIN_INVESTMENT_ANALYSIS_LENGTH = 25;

/**
 * RFC 4122 UUID matcher — accepts both lowercase and uppercase. Backend
 * uses `uuid.Parse` which is also case-insensitive, so client-side checks
 * stay in lockstep.
 */
const UUID_PATTERN =
  /^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$/;

export function isUuid(value: string): boolean {
  return UUID_PATTERN.test(value.trim());
}

/**
 * ISO-3 currency code matcher. Backend CHECK constraint is
 *     currency = '' OR currency ~ '^[A-Z]{3}$'
 * so we enforce the same here. Empty is allowed because currency is
 * optional in the migration.
 */
const CURRENCY_PATTERN = /^[A-Z]{3}$/;

export function isCurrencyCode(value: string): boolean {
  const v = value.trim().toUpperCase();
  return v === "" || CURRENCY_PATTERN.test(v);
}

/**
 * Format an ISO-8601 timestamp string as Asia/Bangkok local time, matching
 * the project-wide rule that backend stores UTC and the UI displays UTC+7
 * (AIREAD.md §7). Falls back to the raw value if parsing fails so a
 * malformed timestamp never breaks the detail page render.
 */
export function formatBangkokDateTime(iso: string | null | undefined): string {
  if (!iso) return "—";
  const date = new Date(iso);
  if (Number.isNaN(date.getTime())) return iso;
  try {
    return new Intl.DateTimeFormat("en-GB", {
      timeZone: "Asia/Bangkok",
      year: "numeric",
      month: "short",
      day: "2-digit",
      hour: "2-digit",
      minute: "2-digit",
    }).format(date);
  } catch {
    return iso;
  }
}

export function recommendationBadgeVariant(
  recommendation: ResearchRecommendation,
): BadgeVariant {
  switch (recommendation) {
    case "BUY":
      return "success";
    case "SELL":
      return "error";
    case "HOLD":
      return "warning";
    default:
      return "neutral";
  }
}

export function reportStatusBadgeVariant(
  status: ResearchReportStatus,
): BadgeVariant {
  switch (status) {
    case "DRAFT":
      return "draft";
    case "ACTIVE":
      return "success";
    case "EXPIRED":
      return "neutral";
    case "REJECTED":
      return "error";
    default:
      return "neutral";
  }
}

export function reviewStatusBadgeVariant(
  status: ResearchReviewStatus,
): BadgeVariant {
  switch (status) {
    case "NOT_SUBMITTED":
      return "neutral";
    case "SUBMITTED":
      return "info";
    case "REVIEW_COMPLETED":
      return "purple";
    default:
      return "neutral";
  }
}

export function reportStatusLabel(
  status: ResearchReportStatus,
  t: (key: any) => string,
): string {
  switch (status) {
    case "DRAFT":
      return t("investmentResearch.status.draft");
    case "ACTIVE":
      return t("investmentResearch.status.active");
    case "EXPIRED":
      return t("investmentResearch.status.expired");
    case "REJECTED":
      return t("investmentResearch.status.rejected");
    default:
      return status;
  }
}

export function reviewStatusLabel(
  status: ResearchReviewStatus,
  t: (key: any) => string,
): string {
  switch (status) {
    case "NOT_SUBMITTED":
      return t("investmentResearch.reviewStatusValues.notSubmitted");
    case "SUBMITTED":
      return t("investmentResearch.reviewStatusValues.submitted");
    case "REVIEW_COMPLETED":
      return t("investmentResearch.reviewStatusValues.reviewCompleted");
    default:
      return status;
  }
}

export function recommendationLabel(
  rec: ResearchRecommendation,
  t: (key: any) => string,
): string {
  switch (rec) {
    case "BUY":
      return t("investmentResearch.recommendationValues.buy");
    case "SELL":
      return t("investmentResearch.recommendationValues.sell");
    case "HOLD":
      return t("investmentResearch.recommendationValues.hold");
    default:
      return rec;
  }
}


export function canEditReport(reviewStatus: ResearchReviewStatus): boolean {
  return reviewStatus !== "REVIEW_COMPLETED";
}

export function canDeleteReport(reviewStatus: ResearchReviewStatus): boolean {
  return reviewStatus === "NOT_SUBMITTED";
}

export function canSubmitReport(reviewStatus: ResearchReviewStatus): boolean {
  return reviewStatus === "NOT_SUBMITTED";
}

export function canCancelSubmitReport(
  reviewStatus: ResearchReviewStatus,
): boolean {
  return reviewStatus === "SUBMITTED";
}
