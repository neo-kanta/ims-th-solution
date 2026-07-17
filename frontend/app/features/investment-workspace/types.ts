export type FundPrivacy = "PRIVATE" | "PUBLIC";

/** One row in the fund workspace's fund switcher. */
export interface FundCard {
  fund_id: string;          // slug used in the URL
  fund_uuid?: string;       // present once backend wires real UUIDs
  code: string;             // e.g. "TH-GOV-LTF"
  short_name: string;       // e.g. "fund-alpha"
  contract_code: string;    // e.g. "A02"
  privacy: FundPrivacy;
  role: "MANAGER" | "DELEGATE" | "APPROVER" | "RISK_VIEWER" | "AUDITOR";
  base_currency: string;
  has_units: boolean;
  status: "ACTIVE" | "SUSPENDED" | "CLOSED";
}
