export type ResearchRecommendation = "BUY" | "SELL" | "HOLD";

export type ResearchReportStatus =
  | "DRAFT"
  | "ACTIVE"
  | "EXPIRED"
  | "REJECTED";

export type ResearchReviewStatus =
  | "NOT_SUBMITTED"
  | "SUBMITTED"
  | "REVIEW_COMPLETED";

export const RECOMMENDATIONS: readonly ResearchRecommendation[] = [
  "BUY",
  "SELL",
  "HOLD",
];

export const REPORT_STATUSES: readonly ResearchReportStatus[] = [
  "DRAFT",
  "ACTIVE",
  "EXPIRED",
  "REJECTED",
];

export const REVIEW_STATUSES: readonly ResearchReviewStatus[] = [
  "NOT_SUBMITTED",
  "SUBMITTED",
  "REVIEW_COMPLETED",
];

export interface ResearchReport {
  id: string;
  report_no: string;
  report_date: string;
  effective_date?: string | null;

  owner_user_id: string;
  author_user_id: string;
  applicable_contract_id?: string | null;

  instrument_type: string;
  instrument_code: string;
  instrument_name: string;
  market: string;
  currency: string;

  recommendation: ResearchRecommendation;
  report_title: string;

  company_overview: string;
  company_outlook: string;
  esg_comment: string;
  financial_status: string;
  investment_analysis: string;

  rejection_reason: string;
  post_submission_note: string;

  report_status: ResearchReportStatus;
  review_status: ResearchReviewStatus;

  created_at: string;
  created_by?: string | null;
  updated_at: string;
  updated_by?: string | null;
}

export interface ResearchReportListPayload {
  items: ResearchReport[];
  total: number;
  page: number;
  limit: number;
}

export interface ResearchReportFilters {
  report_status?: ResearchReportStatus;
  review_status?: ResearchReviewStatus;
  recommendation?: ResearchRecommendation;
  instrument_code?: string;
  owner_user_id?: string;
  report_date_from?: string;
  report_date_to?: string;
  search?: string;
  page?: number;
  limit?: number;
}

export interface CreateResearchReportInput {
  report_no: string;
  report_date: string;
  effective_date?: string;
  owner_user_id: string;
  author_user_id: string;
  applicable_contract_id?: string;
  instrument_type?: string;
  instrument_code: string;
  instrument_name?: string;
  market?: string;
  currency?: string;
  recommendation: ResearchRecommendation;
  report_title?: string;
  company_overview?: string;
  company_outlook?: string;
  esg_comment?: string;
  financial_status?: string;
  investment_analysis: string;
}

export interface UpdateResearchReportInput {
  report_date?: string;
  effective_date?: string;
  owner_user_id?: string;
  author_user_id?: string;
  applicable_contract_id?: string | null;
  instrument_type?: string;
  instrument_code?: string;
  instrument_name?: string;
  market?: string;
  currency?: string;
  recommendation?: ResearchRecommendation;
  report_title?: string;
  company_overview?: string;
  company_outlook?: string;
  esg_comment?: string;
  financial_status?: string;
  investment_analysis?: string;
  rejection_reason?: string;
  post_submission_note?: string;
  report_status?: ResearchReportStatus;
}
