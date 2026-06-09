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
