/**
 * Dashboard Type Definitions
 * Strongly-typed data models for investment management dashboard
 */

export type WorkflowStepStatus =
  | "completed"
  | "in-progress"
  | "pending"
  | "locked";
export type AlertSeverity = "critical" | "warning" | "info";
export type ActivityType =
  | "workflow"
  | "investment"
  | "approval"
  | "system"
  | "audit";

export interface DashboardKpi {
  id: string;
  label: string;
  value: string;
  unit?: string;
  change?: {
    value: number;
    direction: "up" | "down" | "neutral";
    period: string;
  };
  trend?: "positive" | "negative" | "neutral";
}

export interface WorkflowStep {
  id: string;
  name: string;
  status: WorkflowStepStatus;
  completedAt?: string;
  estimatedAt?: string;
}

export interface WorkflowStatus {
  businessDate: string;
  steps: WorkflowStep[];
  overallStatus: "on-track" | "at-risk" | "blocked" | "completed";
}

export interface HoldingSummary {
  symbol: string;
  quantity: number;
  value: number;
  unrealizedPL: number;
  unrealizedPLPercent: number;
  dayChange: number;
  dayChangePercent: number;
}

export interface PortfolioSummary {
  totalValue: number;
  unrealizedPL: number;
  unrealizedPLPercent: number;
  dayChange: number;
  dayChangePercent: number;
  topHoldings: HoldingSummary[];
}

export interface WatchlistAlert {
  id: string;
  type:
    | "price-breach"
    | "volatility"
    | "volume"
    | "risk-limit"
    | "irg-flag";
  severity: AlertSeverity;
  symbol?: string;
  title: string;
  description: string;
  timestamp: string;
  actionUrl?: string;
}

export interface NotificationItem {
  id: string;
  type:
    | "approval-request"
    | "system-alert"
    | "compliance-flag"
    | "workflow-update"
    | "execution-status";
  title: string;
  message: string;
  read: boolean;
  timestamp: string;
  actionUrl?: string;
  badge?: {
    label: string;
    variant: "success" | "warning" | "error" | "info";
  };
}

export interface QuickAction {
  id: string;
  label: string;
  icon: string;
  description?: string;
  url: string;
  badge?: string;
  requiresPermission?: string;
}

export interface ActivityItem {
  id: string;
  type: ActivityType;
  actor: string;
  action: string;
  subject: string;
  timestamp: string;
  details?: Record<string, unknown>;
  severity?: "low" | "medium" | "high";
}

export type DashboardOverviewMetricTone =
  | "primary"
  | "info"
  | "danger"
  | "success";
export type DashboardOverviewChangeTone =
  | "success"
  | "warning"
  | "danger"
  | "neutral";
export type DashboardOverviewStageStatus = "complete" | "active" | "upcoming";
export type DashboardOverviewBadgeTone =
  | "info"
  | "warning"
  | "teal"
  | "success"
  | "purple";
export type DashboardOverviewActivityTone =
  | "success"
  | "info"
  | "warning"
  | "teal"
  | "neutral";

export interface DashboardOverviewMetric {
  id: string;
  label: string;
  value: string;
  icon: string;
  tone: DashboardOverviewMetricTone;
  changeLabel: string;
  changeTone: DashboardOverviewChangeTone;
  helperText: string;
}

export interface DashboardOverviewWorkflowStage {
  id: string;
  sequence: number;
  label: string;
  timeLabel: string;
  status: DashboardOverviewStageStatus;
}

export interface DashboardOverviewWorkflow {
  businessDate: string;
  dayStatus: string;
  stages: DashboardOverviewWorkflowStage[];
}

export interface DashboardOverviewContract {
  id: string;
  code: string;
  assetType: string;
  valueLabel: string;
  statusLabel: string;
  statusTone: DashboardOverviewBadgeTone;
  manager: string;
  updatedAt: string;
}

export interface DashboardOverviewActivityItem {
  id: string;
  actor: string;
  message: string;
  timeLabel: string;
  tone: DashboardOverviewActivityTone;
}

export interface DashboardOverviewPendingApproval {
  id: string;
  contractCode: string;
  valueLabel: string;
  dueLabel: string;
  isUrgent: boolean;
}

export interface DashboardPayload {
  workflow: DashboardOverviewWorkflow;
  metrics: DashboardOverviewMetric[];
  contracts: DashboardOverviewContract[];
  activityFeed: DashboardOverviewActivityItem[];
  pendingApprovals: DashboardOverviewPendingApproval[];
  lastRefresh: string;
  canCreateContract: boolean;
  createContractUrl: string;
}

// ---------------------------------------------------------------------------
// Real API types — integration module task feed
// ---------------------------------------------------------------------------

export type TaskType =
  | "RESEARCH_REVIEW"
  | "WORKFLOW_PENDING"
  | "COMPLIANCE_BREACH";

export type TaskPriority = "HIGH" | "MEDIUM" | "LOW" | "INFO";

export type TaskStatus = "PENDING" | "IN_PROGRESS" | "COMPLETED";

export type ComplianceSeverity =
  | "BLOCK"
  | "WARN"
  | "REQUIRE_APPROVAL"
  | "MONITOR"
  | "";

export interface TaskDTO {
  taskId: string;
  type: TaskType;
  module: string;
  title: string;
  description: string;
  priority: TaskPriority;
  status: TaskStatus;
  businessDate: string;
  sourceRecordId: string;
  sourceType: string;
  actionUrl: string;
  reason: string;
  /** Language-neutral primary identifier for localised title templates. */
  subject: string;
  /** Compliance severity code; empty for non-compliance tasks. */
  severity: ComplianceSeverity;
  canAct: boolean;
  allowedActions: string[];
  contractId: string;
  createdAt: string;
  updatedAt: string;
}

export interface WorkflowStateDTO {
  contractId: string;
  businessDate: string;
  currentState: string;
  updatedAt: string;
}

export interface TaskSummaryDTO {
  total: number;
  byModule: Record<string, number>;
  byPriority: Record<string, number>;
  highPriority: number;
}

export interface DashboardSnapshotDTO {
  tasks: TaskDTO[];
  summary: TaskSummaryDTO;
  workflowStates: WorkflowStateDTO[];
  lastRefreshed: string;
}

export interface TaskListDTO {
  tasks: TaskDTO[];
  summary: TaskSummaryDTO;
}

// ---------------------------------------------------------------------------
// Valuation summary — GET /integration/dashboard/valuation-summary
// ---------------------------------------------------------------------------

export type ValuationScope = "company" | "mine";
export type ValuationSummaryStatus = "AVAILABLE" | "NO_DATA" | "INCOMPLETE";

export interface ValuationSummaryExclusionDTO {
  readonly businessDate: string | null;
  readonly currency: string | null;
  readonly fundCode: string | null;
  readonly portfolioCode: string | null;
  readonly reason: string;
  readonly requiredBusinessDate: string | null;
}

export interface ValuationSummaryCoverageDTO {
  readonly totalFundCount: number;
  readonly includedFundCount: number;
  readonly excludedFundCount: number;
  readonly totalPortfolioCount: number;
  readonly includedPortfolioCount: number;
  readonly excludedPortfolioCount: number;
  readonly excludedCurrencies: readonly string[];
  readonly excludedBusinessDates: readonly string[];
  readonly exclusionReasons: readonly string[];
  readonly exclusions: readonly ValuationSummaryExclusionDTO[];
}

/**
 * Normalized read model for the "AUM Today" / "Today's P&L" cards.
 * Amounts stay as decimal strings — the frontend only formats them, it never
 * recomputes authoritative figures. When dataAvailable is false every
 * numeric field is empty and MUST render as an explicit "not available"
 * state, never a zero.
 */
export interface ValuationSummaryDTO {
  readonly scope: ValuationScope;
  readonly username: string | null;
  readonly status: ValuationSummaryStatus;
  readonly businessDate: string | null;
  readonly currency: string;
  readonly aumToday: string | null;
  readonly todayPnl: string | null;
  readonly todayPnlPercent: string | null;
  readonly asOf: string | null;
  readonly dataAvailable: boolean;
  readonly coverage: ValuationSummaryCoverageDTO;
}

// ---------------------------------------------------------------------------
// AI command bar — disabled/mock-ready state only in Phase 1
// ---------------------------------------------------------------------------

export type AIProviderMode = "disabled" | "stub";

export interface AICommandBarState {
  mode: AIProviderMode;
  query: string;
  loading: boolean;
}

export type DashboardTodoFilter =
  | "my"
  | "approvals"
  | "workflow"
  | "alerts"
  | "done";

export type DashboardTodoAction =
  | "mark-done"
  | "snooze"
  | "more";

export type WorkflowStage =
  | "DAY_OPEN"
  | "MANAGER_APPROVED"
  | "TRANSACTION_CLOSED"
  | "ACCOUNTING_CLOSED";

export const WORKFLOW_STAGES: readonly WorkflowStage[] = [
  "DAY_OPEN",
  "MANAGER_APPROVED",
  "TRANSACTION_CLOSED",
  "ACCOUNTING_CLOSED",
];

export const STATE_RANK: Record<string, number> = {
  NOT_STARTED: 0,
  DAY_OPEN: 1,
  INVESTMENT_DAY_STARTED: 1,
  MANAGER_APPROVED: 2,
  MANAGER_APPROVED_END_OF_DAY: 2,
  TRANSACTION_CLOSED: 3,
  ACCOUNTING_CLOSED: 4,
};
