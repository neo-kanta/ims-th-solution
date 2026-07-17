"""Generate Markdown API documentation from router/OpenAPI inventory.

The Go routers are the endpoint source of truth. Swagger metadata is used when
it exists, because some mounted routes do not currently have OpenAPI comments.
"""

from __future__ import annotations

import json
import re
from pathlib import Path


ROOT = Path(__file__).resolve().parents[2]
DOCS = ROOT / "docs" / "api"
SWAGGER = json.loads((ROOT / "backend" / "docs" / "swagger.json").read_text(encoding="utf-8"))
PATHS = SWAGGER.get("paths", {})
DEFS = SWAGGER.get("definitions", {})
METHOD_ORDER = {"GET": 0, "POST": 1, "PUT": 2, "PATCH": 3, "DELETE": 4}


def clean(value: object) -> str:
    text = "" if value is None else str(value)
    replacements = {
        "\u2014": "-",
        "\u2013": "-",
        "\u2192": "->",
        "\u2265": ">=",
        "\u2264": "<=",
        "\u2018": "'",
        "\u2019": "'",
        "\u201c": '"',
        "\u201d": '"',
    }
    for src, dst in replacements.items():
        text = text.replace(src, dst)
    return text.encode("ascii", "ignore").decode("ascii")


def ref_name(schema: dict | None) -> str:
    if not schema:
        return ""
    if "$ref" in schema:
        return schema["$ref"].split("/")[-1]
    if schema.get("type") == "array":
        return f"array[{ref_name(schema.get('items', {}))}]"
    if schema.get("type") == "object":
        return "object"
    return schema.get("type", "")


def schema_sample(schema: dict | None, depth: int = 0):
    if not schema:
        return None
    if "$ref" in schema:
        name = schema["$ref"].split("/")[-1]
        if depth >= 2 or name not in DEFS:
            return f"<{name}>"
        return schema_sample(DEFS[name], depth + 1)
    if schema.get("type") == "array":
        return [schema_sample(schema.get("items", {}), depth + 1) or "<item>"]
    if schema.get("type") == "object" or "properties" in schema:
        out = {}
        for key, prop in list((schema.get("properties") or {}).items())[:14]:
            out[key] = schema_sample(prop, depth + 1)
        return out
    if schema.get("type") in {"integer", "number"}:
        return 0
    if schema.get("type") == "boolean":
        return False
    if schema.get("type") == "string":
        return schema.get("enum", ["string"])[0]
    return "<value>"


def json_block(value) -> str:
    if value is None:
        return "No request body."
    return "```json\n" + clean(json.dumps(value, indent=2, ensure_ascii=False)) + "\n```"


def table(headers: list[str], rows: list[tuple]) -> str:
    lines = [
        "| " + " | ".join(headers) + " |",
        "| " + " | ".join(["---"] * len(headers)) + " |",
    ]
    if not rows:
        rows = [("None",) + tuple("" for _ in headers[1:])]
    for row in rows:
        cells = [clean(cell).replace("\n", " ").replace("|", "\\|") for cell in row]
        lines.append("| " + " | ".join(cells) + " |")
    return "\n".join(lines)


def get_op(method: str, path: str) -> dict | None:
    return PATHS.get(path, {}).get(method.lower())


def op_params(op: dict | None, location: str) -> list[tuple]:
    rows = []
    for param in (op or {}).get("parameters") or []:
        if param.get("in") != location:
            continue
        typ = param.get("type") or ref_name(param.get("schema")) or "string"
        rows.append((param.get("name", ""), typ, "Yes" if param.get("required") else "No", param.get("description", "")))
    return rows


def body_from_op(op: dict | None):
    for param in (op or {}).get("parameters") or []:
        if param.get("in") == "body":
            return param.get("schema"), param.get("description") or param.get("name") or ""
    return None, ""


def responses_from_op(op: dict | None) -> list[tuple]:
    rows = []
    for code, resp in sorted(((op or {}).get("responses") or {}).items(), key=lambda item: str(item[0])):
        schema = resp.get("schema")
        rows.append((code, resp.get("description", ""), ref_name(schema)))
    return rows


def path_params(path: str) -> list[tuple]:
    return [(name, "string", "Yes", "Path parameter") for name in re.findall(r"{([^}]+)}", path)]


def default_rule(path: str, method: str) -> str:
    if path.startswith("/auth"):
        return "Supports identity authentication, session issuance, MFA, and session lifecycle controls."
    if path.startswith("/admin"):
        return "Runs under IAM admin security: JWT, active session, IP allowlist, rate limiting, and function-permission checks."
    if path.startswith("/permissions"):
        if any(token in path for token in ["/submit", "/approve", "/reject", "/merge", "/close", "/cancel", "/approval-steps"]):
            return "Participates in permission change maker-checker governance. Lifecycle transitions are audited and terminal-state changes are guarded by service logic."
        return "Supports Account, Group, Account & Group, Function Permissions, and Data Permissions management."
    if path.startswith("/approvals") or path.startswith("/approval-config"):
        return "Participates in the Approval module maker-checker flow. The backend enforces allowed actions, delegation, subject access, and terminal-state rules."
    if path.startswith("/workflow"):
        return "Supports the IMS daily workflow sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing."
    if path.startswith("/compliance"):
        return "Supports IRG / Compliance validation. Pre-trade and post-trade checks persist check records and breaches; overrides require compliance permission and written reason."
    if path.startswith("/investment/decisions"):
        return "Supports Investment Decision lifecycle. Submit runs workflow gating, research-report rules where configured, IRG pre-trade validation, and approval workflow submission."
    if path.startswith("/investment/research-reports"):
        return "Supports Investment Research / Analysis Report lifecycle and approval readiness for decisions."
    if path.startswith("/investment/executions") or path.startswith("/investment/trade-confirmations"):
        return "Supports Investment Execution and Trade Confirmation stages from approved decision through settlement review."
    if path.startswith("/investment/portfolios"):
        return "Supports portfolio master data, holdings, cash, immutable ledger operations, and valuation. Mutations are audited."
    if path.startswith("/investment"):
        return "Supports Stock Investment Management master data, decision, execution, confirmation, valuation, and audit flows."
    if path.startswith("/market-data"):
        return "Supplies quote and history data used by investment valuation and dashboards."
    if path.startswith("/notifications"):
        return "Supports user notifications and operator email outbox administration for approval, workflow, watchlist, and system events."
    if "/audit" in path:
        return "Reads or exports audit events for operational and regulatory traceability."
    return "Business rule not found in code."


class Endpoint:
    def __init__(
        self,
        method: str,
        path: str,
        permission: str = "",
        purpose: str = "",
        auth: str = "JWT required",
        status: str = "Implemented",
        source: str = "",
        business_rule: str = "",
        query: list[tuple] | None = None,
        body_sample=None,
        body_schema: str = "",
        responses: list[tuple] | None = None,
    ):
        self.method = method.upper()
        self.path = path
        self.permission = permission or "Permission rule not found in code."
        self.auth = auth
        self.status = status
        self.source = source
        self.business_rule = business_rule
        self.query = query
        self.body_sample = body_sample
        self.body_schema = body_schema
        self.responses = responses
        self.op = get_op(self.method, self.path)
        self.purpose = purpose or (self.op or {}).get("summary") or "Endpoint purpose not described in OpenAPI annotations."

    def detail(self) -> str:
        description = (self.op or {}).get("description", "")
        purpose = clean(self.purpose)
        if description and clean(description) not in purpose:
            purpose += " " + clean(description)

        body_schema, _ = body_from_op(self.op)
        schema_label = self.body_schema or ref_name(body_schema)
        sample = self.body_sample if self.body_sample is not None else schema_sample(body_schema)
        body_intro = f"Schema: `{schema_label}`." if sample is not None and schema_label else ""
        body = body_intro + ("\n" if body_intro else "") + json_block(sample)

        responses = self.responses if self.responses is not None else responses_from_op(self.op)
        if not responses:
            responses = [("200", "Success", "")]

        return "\n".join(
            [
                f"### {self.method} {self.path}",
                "",
                "#### Purpose",
                purpose,
                "",
                "#### Business Rule",
                self.business_rule or default_rule(self.path, self.method),
                "",
                "#### Request",
                "",
                "##### Path Parameters",
                table(["Name", "Type", "Required", "Description"], op_params(self.op, "path") or path_params(self.path)),
                "",
                "##### Query Parameters",
                table(["Name", "Type", "Required", "Description"], self.query if self.query is not None else op_params(self.op, "query")),
                "",
                "##### Request Body",
                body,
                "",
                "#### Response",
                table(["HTTP Status", "Description", "Schema"], responses),
                "",
                "#### Source",
                f"`{self.source}`.",
            ]
        ).strip()


def ep(method: str, path: str, permission: str = "", **kwargs) -> Endpoint:
    return Endpoint(method, path, permission=permission, **kwargs)


def write_doc(
    filename: str,
    title: str,
    purpose: str,
    business_context: str,
    authentication: str,
    authorization: str,
    endpoints: list[Endpoint],
    sources: list[str],
    notes: str = "",
) -> None:
    endpoints = sorted(endpoints, key=lambda item: (item.path, METHOD_ORDER.get(item.method, 99)))
    summary = table(
        ["Method", "Path", "Purpose", "Auth", "Permission", "Status"],
        [(e.method, e.path, e.purpose, e.auth, e.permission, e.status) for e in endpoints],
    )
    details = "\n\n".join(e.detail() for e in endpoints)
    source_lines = "\n".join(f"- `{source}`" for source in sources)
    notes_block = f"\n\n## Notes\n{notes.strip()}\n" if notes.strip() else ""
    parts = [
        f"# {title}",
        "",
        "## Purpose",
        purpose,
        "",
        "## Business Context",
        business_context,
        "",
        "## Authentication",
        authentication,
        "",
        "## Authorization / Permission",
        authorization,
        "",
        "## Endpoint Summary",
        summary,
        "",
        "## Endpoint Detail",
        "",
        details,
        "",
        "## Source References",
        source_lines,
    ]
    if notes_block:
        parts.append(notes_block.strip())
    content = "\n".join(parts)
    (DOCS / filename).write_text(clean(content).strip() + "\n", encoding="utf-8")


def eps(mapping: dict[tuple[str, str], str], source: str) -> list[Endpoint]:
    return [ep(method, path, permission, source=source) for (method, path), permission in mapping.items()]


AUTH = {
    ("POST", "/auth/login"): "",
    ("POST", "/auth/refresh"): "",
    ("GET", "/auth/me"): "",
    ("POST", "/auth/logout"): "",
    ("POST", "/auth/logout-all"): "",
    ("POST", "/auth/change-password"): "",
    ("POST", "/auth/mfa/enroll"): "",
    ("GET", "/auth/mfa/status"): "",
    ("GET", "/auth/mfa/dev/totp-code"): "",
    ("POST", "/auth/mfa/verify"): "",
    ("POST", "/auth/mfa/disable"): "",
    ("GET", "/auth/sessions"): "",
    ("POST", "/auth/sessions/{id}/revoke"): "",
}

IAM = {
    ("GET", "/admin/users"): "IAM_USER_VIEW",
    ("POST", "/admin/users"): "IAM_USER_CREATE",
    ("POST", "/admin/users/{id}/disable"): "IAM_USER_DEACTIVATE",
    ("POST", "/admin/users/{id}/enable"): "IAM_USER_DEACTIVATE",
    ("POST", "/admin/users/{id}/lock"): "IAM_USER_UPDATE",
    ("POST", "/admin/users/{id}/unlock"): "IAM_USER_UPDATE",
    ("POST", "/admin/users/{id}/reset-password"): "IAM_USER_UPDATE",
    ("GET", "/admin/users/{id}/sessions"): "IAM_USER_UPDATE",
    ("POST", "/admin/sessions/{id}/revoke"): "IAM_USER_UPDATE",
}

AUDIT = {
    ("GET", "/admin/audit"): "IAM_AUDIT_VIEW",
    ("GET", "/admin/audit/export"): "IAM_AUDIT_VIEW",
}

APPROVAL = {
    ("GET", "/approvals/inbox"): "APPROVAL_VIEW_INBOX",
    ("GET", "/approvals/requests"): "APPROVAL_VIEW_REQUEST",
    ("GET", "/approvals/requests/{requestId}"): "APPROVAL_VIEW_REQUEST",
    ("GET", "/approvals/requests/{requestId}/timeline"): "APPROVAL_AUDIT_VIEW",
    ("GET", "/approvals/subjects/{subjectType}/{subjectId}/status"): "APPROVAL_VIEW_REQUEST",
    ("POST", "/approvals/submit"): "APPROVAL_SUBMIT",
    ("POST", "/approvals/tasks/{taskId}/approve"): "APPROVAL_APPROVE",
    ("POST", "/approvals/tasks/{taskId}/reject"): "APPROVAL_REJECT",
    ("POST", "/approvals/requests/{requestId}/withdraw"): "APPROVAL_WITHDRAW",
    ("POST", "/approvals/requests/{requestId}/cancel"): "APPROVAL_CANCEL",
    ("POST", "/approvals/requests/{requestId}/revoke"): "APPROVAL_REVOKE",
    ("GET", "/approval-config/groups"): "APPROVAL_CONFIG_VIEW",
    ("POST", "/approval-config/groups"): "APPROVAL_GROUP_MANAGE",
    ("PUT", "/approval-config/groups/{id}"): "APPROVAL_GROUP_MANAGE",
    ("GET", "/approval-config/groups/{id}/members"): "APPROVAL_CONFIG_VIEW",
    ("POST", "/approval-config/groups/{id}/members"): "APPROVAL_GROUP_MANAGE",
    ("PUT", "/approval-config/groups/{id}/members/{memberId}"): "APPROVAL_GROUP_MANAGE",
    ("POST", "/approval-config/groups/{id}/members/{memberId}/approve"): "APPROVAL_GROUP_MANAGE",
    ("POST", "/approval-config/groups/{id}/members/{memberId}/revoke"): "APPROVAL_GROUP_MANAGE",
    ("POST", "/approval-config/groups/{id}/members/reorder"): "APPROVAL_GROUP_MANAGE",
    ("GET", "/approval-config/teams"): "APPROVAL_CONFIG_VIEW",
    ("POST", "/approval-config/teams"): "APPROVAL_TEAM_MANAGE",
    ("PUT", "/approval-config/teams/{id}"): "APPROVAL_TEAM_MANAGE",
    ("GET", "/approval-config/teams/{id}/contracts"): "APPROVAL_CONFIG_VIEW",
    ("POST", "/approval-config/teams/{id}/contracts"): "APPROVAL_TEAM_MANAGE",
    ("GET", "/approval-config/teams/{id}/members"): "APPROVAL_CONFIG_VIEW",
    ("POST", "/approval-config/teams/{id}/members"): "APPROVAL_TEAM_MANAGE",
    ("PUT", "/approval-config/teams/{id}/members/{memberId}"): "APPROVAL_TEAM_MANAGE",
    ("DELETE", "/approval-config/teams/{id}/members/{memberId}"): "APPROVAL_TEAM_MANAGE",
    ("GET", "/approval-config/processes"): "APPROVAL_CONFIG_VIEW",
    ("GET", "/approval-config/processes/{id}"): "APPROVAL_CONFIG_VIEW",
    ("POST", "/approval-config/processes"): "APPROVAL_PROCESS_MANAGE",
    ("PUT", "/approval-config/processes/{id}"): "APPROVAL_PROCESS_MANAGE",
    ("POST", "/approval-config/processes/{id}/activate"): "APPROVAL_PROCESS_MANAGE",
    ("POST", "/approval-config/processes/{id}/deactivate"): "APPROVAL_PROCESS_MANAGE",
}

WORKFLOW = {
    ("GET", "/workflow/day-states/{contractId}"): "WORKFLOW_VIEW",
    ("GET", "/workflow/day-states/{contractId}/history"): "WORKFLOW_VIEW",
    ("POST", "/workflow/day-states/{contractId}/transitions"): "Per-action workflow permission",
    ("POST", "/workflow/scheduler/run-once"): "WORKFLOW_RUN_SCHEDULER",
    ("GET", "/workflow/daily"): "Settings-based approver/admin rule; no RequirePermission middleware",
    ("POST", "/workflow/daily/execute"): "Settings-based approver/admin rule; no RequirePermission middleware",
    ("GET", "/workflow/daily/transitions"): "Settings-based approver/admin rule; no RequirePermission middleware",
    ("GET", "/workflow/transition-rules"): "Settings-based approver/admin rule; no RequirePermission middleware",
    ("GET", "/workflow/settings"): "Settings-based approver/admin rule; no RequirePermission middleware",
    ("PUT", "/workflow/settings"): "Admin role required by handler; no RequirePermission middleware",
}

COMPLIANCE = {
    ("POST", "/compliance/checks/pre-trade"): "WORKFLOW_EXECUTE",
    ("POST", "/compliance/checks/post-trade"): "WORKFLOW_EXECUTE",
    ("GET", "/compliance/checks/{groupID}"): "IRG_VIEW_RULES",
    ("GET", "/compliance/breaches"): "IRG_VIEW_RULES",
    ("POST", "/compliance/breaches/{breachID}/override"): "IRG_OVERRIDE_BREACH",
    ("GET", "/compliance/rules"): "IRG_VIEW_RULES",
    ("POST", "/compliance/rules"): "IRG_EDIT_RULE_INSTANCE",
}

INVESTMENT = {
    ("GET", "/investment/reference/asset-classes"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/reference/asset-subtypes"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/reference/sectors"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/reference/fund-categories"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/reference/regions"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/reference/countries"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/reference/investment-styles"): "INVESTMENT_REFERENCE_VIEW",
    ("GET", "/investment/funds"): "INVESTMENT_FUND_VIEW",
    ("GET", "/investment/funds/{id}"): "INVESTMENT_FUND_VIEW",
    ("POST", "/investment/funds"): "INVESTMENT_FUND_MANAGE",
    ("PUT", "/investment/funds/{id}"): "INVESTMENT_FUND_MANAGE",
    ("DELETE", "/investment/funds/{id}"): "INVESTMENT_FUND_MANAGE",
    ("GET", "/investment/instruments"): "INVESTMENT_INSTRUMENT_VIEW",
    ("GET", "/investment/instruments/{id}"): "INVESTMENT_INSTRUMENT_VIEW",
    ("POST", "/investment/instruments"): "INVESTMENT_INSTRUMENT_MANAGE",
    ("PUT", "/investment/instruments/{id}"): "INVESTMENT_INSTRUMENT_MANAGE",
    ("POST", "/investment/instruments/{id}/prices"): "INVESTMENT_PRICE_POST",
    ("GET", "/investment/funds/{id}/nav/latest"): "INVESTMENT_VALUATION_VIEW",
    ("GET", "/investment/funds/{id}/allocation"): "INVESTMENT_VALUATION_VIEW",
    ("GET", "/investment/funds/{id}/nav-history"): "INVESTMENT_VALUATION_VIEW",
    ("GET", "/investment/funds/{id}/holdings/valuation"): "INVESTMENT_VALUATION_VIEW",
    ("GET", "/investment/funds/{id}/market-data/status"): "INVESTMENT_VALUATION_VIEW",
    ("POST", "/investment/funds/{id}/market-data/refresh"): "INVESTMENT_VALUATION_RUN",
    ("POST", "/investment/funds/{id}/aum/compute"): "INVESTMENT_FUND_AUM_COMPUTE",
    ("GET", "/investment/decisions"): "INVESTMENT_DECISION_VIEW",
    ("GET", "/investment/decisions/{id}"): "INVESTMENT_DECISION_VIEW",
    ("GET", "/investment/decisions/{id}/details"): "INVESTMENT_DECISION_VIEW",
    ("GET", "/investment/decisions/approval-items"): "INVESTMENT_DECISION_APPROVE",
    ("POST", "/investment/decisions/batch-approve"): "INVESTMENT_DECISION_APPROVE",
    ("POST", "/investment/decisions/batch-reject"): "INVESTMENT_DECISION_APPROVE",
    ("POST", "/investment/decisions"): "INVESTMENT_DECISION_MANAGE",
    ("PUT", "/investment/decisions/{id}"): "INVESTMENT_DECISION_MANAGE",
    ("POST", "/investment/decisions/{id}/submit"): "INVESTMENT_DECISION_SUBMIT",
    ("POST", "/investment/decisions/{id}/cancel"): "INVESTMENT_DECISION_CANCEL",
    ("GET", "/investment/executions"): "INVESTMENT_EXECUTION_VIEW",
    ("GET", "/investment/executions/{id}"): "INVESTMENT_EXECUTION_VIEW",
    ("POST", "/investment/executions"): "INVESTMENT_EXECUTION_MANAGE",
    ("POST", "/investment/executions/{id}/fill"): "INVESTMENT_EXECUTION_MANAGE",
    ("POST", "/investment/executions/{id}/cancel"): "INVESTMENT_EXECUTION_MANAGE",
    ("GET", "/investment/trade-confirmations"): "INVESTMENT_CONFIRMATION_VIEW",
    ("GET", "/investment/trade-confirmations/{id}"): "INVESTMENT_CONFIRMATION_VIEW",
    ("POST", "/investment/trade-confirmations"): "INVESTMENT_CONFIRMATION_MANAGE",
    ("POST", "/investment/trade-confirmations/{id}/resolve"): "INVESTMENT_CONFIRMATION_MANAGE",
    ("POST", "/investment/trade-confirmations/batch"): "INVESTMENT_CONFIRMATION_IMPORT",
    ("GET", "/investment/research-reports"): "INVESTMENT_RESEARCH_VIEW",
    ("GET", "/investment/research-reports/{id}"): "INVESTMENT_RESEARCH_VIEW",
    ("POST", "/investment/research-reports"): "INVESTMENT_RESEARCH_CREATE",
    ("PUT", "/investment/research-reports/{id}"): "INVESTMENT_RESEARCH_UPDATE",
    ("DELETE", "/investment/research-reports/{id}"): "INVESTMENT_RESEARCH_DELETE",
    ("POST", "/investment/research-reports/{id}/submit"): "INVESTMENT_RESEARCH_SUBMIT",
    ("POST", "/investment/research-reports/{id}/cancel-submit"): "INVESTMENT_RESEARCH_CANCEL_SUBMIT",
    ("POST", "/investment/research-reports/{id}/invalidate"): "INVESTMENT_RESEARCH_INVALIDATE",
}

PORTFOLIO = {
    ("GET", "/investment/portfolios"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/investment/portfolios/{id}"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/investment/portfolios/{id}/holdings"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/investment/portfolios/{id}/cash"): "INVESTMENT_PORTFOLIO_VIEW",
    ("POST", "/investment/portfolios"): "INVESTMENT_PORTFOLIO_MANAGE",
    ("PUT", "/investment/portfolios/{id}"): "INVESTMENT_PORTFOLIO_MANAGE",
    ("DELETE", "/investment/portfolios/{id}"): "INVESTMENT_PORTFOLIO_MANAGE",
    ("GET", "/investment/portfolios/{id}/transactions"): "INVESTMENT_PORTFOLIO_VIEW",
    ("POST", "/investment/portfolios/{id}/transactions"): "INVESTMENT_LEDGER_POST",
    ("POST", "/investment/portfolios/{id}/transactions/simulate"): "INVESTMENT_LEDGER_SIMULATE",
    ("POST", "/investment/portfolios/{id}/transactions/{txnId}/reverse"): "INVESTMENT_LEDGER_REVERSE",
    ("GET", "/investment/portfolios/{id}/valuations"): "INVESTMENT_VALUATION_VIEW",
    ("GET", "/investment/portfolios/{id}/valuations/latest"): "INVESTMENT_VALUATION_VIEW",
    ("POST", "/investment/portfolios/{id}/valuations/run"): "INVESTMENT_VALUATION_RUN",
}

MARKET_DATA = {
    ("GET", "/market-data/quote"): "",
    ("GET", "/market-data/history"): "",
    ("POST", "/market-data/import"): "",
    ("GET", "/market-data/provider-health"): "",
    ("POST", "/market-data/import-batches"): "",
    ("POST", "/market-data/import-batches/{batch_id}/run"): "",
    ("GET", "/market-data/import-batches/{batch_id}"): "",
    ("GET", "/market-data/import-batches/{batch_id}/errors"): "",
    ("GET", "/market-data/screen/watchlist"): "",
    ("GET", "/market-data/screen/search"): "",
}

NOTIFICATION = {
    ("GET", "/notifications"): "",
    ("POST", "/notifications/{id}/read"): "",
    ("POST", "/notifications/read-all"): "",
    ("GET", "/notifications/email-outbox"): "NOTIFICATION_VIEW",
    ("GET", "/notifications/email-outbox/{outbox_id}"): "NOTIFICATION_VIEW",
    ("POST", "/notifications/email-outbox/{outbox_id}/retry"): "NOTIFICATION_RETRY",
    ("POST", "/notifications/email/test"): "NOTIFICATION_TEST",
    ("GET", "/notifications/email/health"): "NOTIFICATION_HEALTH",
}


PERMISSION_SPECS = [
    ("GET", "/permissions/change-requests", "List permission change requests", "permission.change_request.review", None, [("status", "string", "No", "Request status"), ("risk", "string", "No", "Risk level"), ("label", "string", "No", "Label filter"), ("module", "string", "No", "Module filter"), ("search", "string", "No", "Search text"), ("requester", "uuid", "No", "Requester user ID"), ("reviewer", "uuid", "No", "Reviewer user ID"), ("page", "integer", "No", "Page number"), ("limit", "integer", "No", "Page size")], "Paged change request object"),
    ("GET", "/permissions/change-requests/{id}", "Get a permission change request", "permission.change_request.review", None, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests", "Create a permission change request", "permission.change_request.create", {"title": "string", "description": "string", "request_type": "FUNCTION_PERMISSION", "risk_level": "MEDIUM", "target_entity_type": "USER", "target_entity_id": "uuid"}, None, "ChangeRequest"),
    ("PUT", "/permissions/change-requests/{id}", "Update a permission change request", "permission.change_request.create", {"title": "string", "description": "string", "request_type": "FUNCTION_PERMISSION", "risk_level": "MEDIUM", "target_entity_type": "USER", "target_entity_id": "uuid"}, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/submit", "Submit a permission change request", "permission.change_request.submit", None, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/approve", "Approve a permission change request", "permission.change_request.approve", {"comment": "string"}, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/request-changes", "Request changes on a permission change request", "permission.change_request.approve", {"comment": "string"}, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/reject", "Reject a permission change request", "permission.change_request.reject", {"comment": "string"}, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/merge", "Merge an approved permission change request", "permission.change_request.merge", None, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/close", "Close a permission change request", "permission.change_request.close", None, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/cancel", "Cancel a permission change request", "permission.change_request.cancel", None, None, "ChangeRequest"),
    ("GET", "/permissions/change-requests/{id}/items", "List change request items", "permission.change_request.review", None, None, "array[ChangeItem]"),
    ("POST", "/permissions/change-requests/{id}/items", "Add a change request item", "permission.change_request.create", {"item_type": "FUNCTION_RIGHT", "target_table": "permissions_function_rights", "target_id": "uuid", "action_type": "ADD", "before_json": {}, "after_json": {}}, None, "ChangeItem"),
    ("PUT", "/permissions/change-requests/{id}/items/{itemId}", "Update a change request item", "permission.change_request.create", {"item_type": "FUNCTION_RIGHT", "target_table": "permissions_function_rights", "target_id": "uuid", "action_type": "UPDATE", "before_json": {}, "after_json": {}}, None, "204 No Content"),
    ("DELETE", "/permissions/change-requests/{id}/items/{itemId}", "Delete a change request item", "permission.change_request.create", None, None, "204 No Content"),
    ("GET", "/permissions/change-requests/{id}/approval-steps", "List approval steps", "permission.change_request.review", None, None, "array[ApprovalStep]"),
    ("POST", "/permissions/change-requests/{id}/approval-steps/{stepId}/approve", "Approve a permission approval step", "permission.change_request.approve", {"comment": "string"}, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/approval-steps/{stepId}/request-changes", "Request changes on a permission approval step", "permission.change_request.approve", {"comment": "string"}, None, "ChangeRequest"),
    ("POST", "/permissions/change-requests/{id}/approval-steps/{stepId}/reject", "Reject a permission approval step", "permission.change_request.reject", {"comment": "string"}, None, "ChangeRequest"),
    ("GET", "/permissions/change-requests/{id}/comments", "List change request comments", "permission.change_request.review", None, None, "array[Comment]"),
    ("POST", "/permissions/change-requests/{id}/comments", "Add a change request comment", "permission.change_request.review", {"comment": "string"}, None, "Comment"),
    ("PUT", "/permissions/change-requests/{id}/comments/{commentId}", "Update a change request comment", "permission.change_request.review", {"comment": "string"}, None, "204 No Content"),
    ("DELETE", "/permissions/change-requests/{id}/comments/{commentId}", "Delete a change request comment", "permission.change_request.review", None, None, "204 No Content"),
    ("GET", "/permissions/change-requests/{id}/checks", "List change request checks", "permission.change_request.review", None, None, "array[Check]"),
    ("POST", "/permissions/change-requests/{id}/rerun-checks", "Rerun change request checks", "permission.change_request.review", None, None, "array[Check]"),
    ("GET", "/permissions/change-requests/{id}/diff", "Get change request diff", "permission.change_request.review", None, None, "Diff object"),
    ("GET", "/permissions/labels", "List permission labels", "permission.change_request.review", None, None, "array[Label]"),
    ("POST", "/permissions/labels", "Create or update a permission label", "permission.change_request.create", {"label_code": "string", "label_name": "string", "label_type": "RISK", "color": "#ff0000", "description": "string", "is_active": True}, None, "Label"),
    ("PUT", "/permissions/labels/{id}", "Update a permission label", "permission.change_request.create", {"label_code": "string", "label_name": "string", "label_type": "RISK", "color": "#ff0000", "description": "string", "is_active": True}, None, "Label"),
    ("POST", "/permissions/change-requests/{id}/labels", "Add a label to a change request", "permission.change_request.create", {"label_code": "string"}, None, "204 No Content"),
    ("DELETE", "/permissions/change-requests/{id}/labels/{labelId}", "Remove a label from a change request", "permission.change_request.create", None, None, "204 No Content"),
    ("GET", "/permissions/roles", "List roles", "permission.roles.view", None, None, "array[Role]"),
    ("GET", "/permissions/roles/{id}", "Get a role", "permission.roles.view", None, None, "Role"),
    ("GET", "/permissions/role-assignment-policies", "List role assignment policies", "permission.roles.view", None, None, "array[RoleAssignmentPolicy]"),
    ("POST", "/permissions/users/{userId}/role-assignment-request", "Create a role assignment request", "permission.change_request.create", {"role_id": "uuid", "role_code": "string", "reason": "string"}, None, "Role assignment request object"),
    ("GET", "/permissions/users", "List permission users", "permission.users.view", None, [("search", "string", "No", "Search text"), ("page", "integer", "No", "Page number"), ("limit", "integer", "No", "Page size")], "Paged UserSummary object"),
    ("GET", "/permissions/users/{id}", "Get permission user detail", "permission.users.view", None, None, "UserSummary"),
    ("GET", "/permissions/groups", "List permission groups", "permission.groups.view", None, [("search", "string", "No", "Search text"), ("page", "integer", "No", "Page number"), ("limit", "integer", "No", "Page size")], "Paged GroupSummary object"),
    ("GET", "/permissions/groups/{id}", "Get permission group detail", "permission.groups.view", None, None, "GroupSummary"),
    ("GET", "/permissions/function-definitions", "List function definitions", "permission.function_rights.view", None, None, "array[FunctionDefinition]"),
    ("GET", "/permissions/function-rights", "List function rights", "permission.function_rights.view", None, None, "array[FunctionRight]"),
    ("GET", "/permissions/effective/users/{userId}", "Get effective permissions for a user", "permission.users.view", None, None, "EffectivePermission"),
    ("GET", "/permissions/data-rights", "List data rights", "permission.data_rights.view", None, None, "array[DataRight]"),
    ("GET", "/permissions/approval-settings", "List permission approval settings", "permission.approval_settings.view", None, None, "array[ApprovalSetting]"),
    ("GET", "/permissions/approval-settings/{id}", "Get permission approval setting", "permission.approval_settings.view", None, None, "ApprovalSetting"),
    ("GET", "/permissions/notification-settings", "List permission notification settings", "permission.notification.view", None, None, "array[NotificationSetting]"),
    ("PUT", "/permissions/notification-settings", "Update permission notification settings", "permission.notification.edit", {"settings": [{"event_code": "string", "channel": "email", "enabled": True, "target_scope": "REQUESTER", "template_subject": "string", "template_body": "string"}]}, None, "204 No Content"),
]


def permission_endpoints() -> list[Endpoint]:
    endpoints = []
    for method, path, purpose, permission, body, query, response in PERMISSION_SPECS:
        if response == "204 No Content":
            responses = [("204", "No Content", ""), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]
        else:
            code = "201" if method == "POST" and any(token in purpose.lower() for token in ["create", "add"]) else "200"
            responses = [(code, "Created" if code == "201" else "OK", response), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("422", "Unprocessable Entity", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]
        endpoints.append(ep(method, path, permission, purpose=purpose, source="backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go", query=query, body_sample=body, responses=responses))
    return endpoints


def main() -> None:
    DOCS.mkdir(parents=True, exist_ok=True)

    auth_eps = eps(AUTH, "backend/internal/iam/module.go; backend/internal/iam/transport/handler")
    for item in auth_eps:
        if item.path in {"/auth/login", "/auth/refresh"}:
            item.auth = "Not required"

    write_doc(
        "auth-api.md",
        "Authentication API",
        "Provides login, token refresh, current-user identity, password change, MFA enrollment/verification, and session management for IMS users.",
        "Auth APIs establish the identity used by workflow, approval, permission, compliance, and investment modules. JWT claims carry user subject, username, roles/groups, and session identity used by downstream permission checks.",
        "Login and refresh are public but rate-limited. All other endpoints require a valid Bearer JWT and active server-side session.",
        "No function-permission middleware is mounted for user self-service auth endpoints. Sensitive operations are protected by JWT, active-session validation, and rate limiting; permission rule not found in code.",
        auth_eps,
        ["backend/internal/iam/module.go", "backend/internal/iam/transport/handler/auth_handler.go", "backend/internal/iam/transport/handler/mfa_handler.go", "backend/internal/iam/transport/handler/session_handler.go", "backend/docs/swagger.json"],
    )

    write_doc(
        "iam-api.md",
        "IAM API",
        "Provides administrative account and session operations for IAM users.",
        "IAM supports the Account part of IMS permission management by maintaining user accounts, lock/disable state, password reset state, and active sessions used by JWT validation.",
        "All IAM admin endpoints require Bearer JWT, active session validation, admin IP allowlist, and admin rate limiting.",
        "Route-level permission checks are enforced with `IAM_USER_VIEW`, `IAM_USER_CREATE`, `IAM_USER_UPDATE`, or `IAM_USER_DEACTIVATE`.",
        eps(IAM, "backend/internal/iam/module.go; backend/internal/iam/transport/handler/admin_handler.go"),
        ["backend/internal/iam/module.go", "backend/internal/iam/permission/policies.go", "backend/internal/iam/transport/handler/admin_handler.go", "backend/internal/iam/transport/handler/session_handler.go", "backend/docs/swagger.json"],
    )

    audit_eps = eps(AUDIT, "backend/internal/audit/module.go; backend/internal/audit/transport/handler/audit_handler.go")
    audit_eps.append(ep("GET", "/audit/logs", "permission.audit.view", purpose="List permission-management audit logs", source="backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go", query=[("page", "integer", "No", "Page number"), ("limit", "integer", "No", "Page size")], responses=[("200", "OK", "Paged permission audit log object"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]))
    audit_eps.append(ep("GET", "/audit/logs/export", "permission.audit.export", purpose="Export permission-management audit logs as CSV", source="backend/internal/permissions/module.go; backend/internal/permissions/transport/handler/permission_handler.go", query=[("page", "integer", "No", "Page number"), ("limit", "integer", "No", "Page size")], responses=[("200", "CSV export", "text/csv"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]))
    write_doc(
        "audit-api.md",
        "Audit API",
        "Provides immutable audit-event query and export endpoints for system and permission-governance audit trails.",
        "Audit APIs support regulatory traceability across login/session events, IAM administration, permission change governance, workflow, approval, compliance, and investment lifecycle events.",
        "All audit endpoints require Bearer JWT. `/admin/audit` endpoints also inherit IAM admin IP allowlist and admin/export rate limits.",
        "`/admin/audit` and `/admin/audit/export` require `IAM_AUDIT_VIEW`; `/audit/logs` requires `permission.audit.view`; `/audit/logs/export` requires `permission.audit.export`.",
        audit_eps,
        ["backend/internal/audit/module.go", "backend/internal/audit/transport/handler/audit_handler.go", "backend/internal/permissions/module.go", "backend/internal/permissions/transport/handler/permission_handler.go", "backend/internal/iam/module.go", "backend/docs/swagger.json"],
    )

    write_doc(
        "permission-api.md",
        "Permission API",
        "Provides permission-governance APIs for accounts, groups, roles, function permissions, data permissions, effective access, change requests, labels, approval settings, and notification settings.",
        "Permission APIs preserve the IMS Permission Management model: Account, Group, Account & Group, Function Permissions, and Data Permissions. Mutating access changes are routed through permission change requests and maker-checker approval steps.",
        "All permission endpoints require Bearer JWT through `/api/v1`.",
        "Route-level permissions use dot-notation codes such as `permission.users.view`, `permission.groups.view`, `permission.function_rights.view`, `permission.data_rights.view`, and `permission.change_request.approve`.",
        permission_endpoints(),
        ["backend/internal/permissions/module.go", "backend/internal/permissions/transport/handler/permission_handler.go", "backend/internal/permissions/application/service/permission_service.go", "backend/internal/permissions/domain/models.go", "backend/internal/permissions/permission/policies.go"],
        "The permissions module is more granular than the legacy `PERMISSIONS_VIEW` and `PERMISSIONS_MANAGE` provider codes. The fine-grained dot-notation permissions are seeded separately.",
    )

    write_doc(
        "approval-api.md",
        "Approval API",
        "Provides the generic maker-checker approval runtime and approval-group, team, and process configuration APIs.",
        "Approval APIs support review, approve, reject, submit, withdraw, cancel, revoke, delegation, and maker-checker controls for IMS subjects such as Investment Research Reports, Investment Decisions, and Compliance Release requests.",
        "All approval endpoints are mounted under `/api/v1` behind Bearer JWT authentication.",
        "Every approval route is gated by `middleware.RequirePermission` using the approval permission catalog. Subject-level access is also checked through registered subject access ports for investment subjects.",
        eps(APPROVAL, "backend/internal/approval/transport/router.go"),
        ["backend/internal/approval/transport/router.go", "backend/internal/approval/transport/handler/runtime_handler.go", "backend/internal/approval/transport/handler/config_handler.go", "backend/internal/approval/application/service/runtime_service.go", "backend/internal/approval/permission/policies.go", "backend/docs/swagger.json"],
        "Maker-checker is enforced in the approval runtime: the submitter cannot approve their own request. Delegation is checked when computing the viewer task and when authorizing approve/reject actions.",
    )

    write_doc(
        "workflow-api.md",
        "Workflow API",
        "Provides daily workflow state, transition execution, scheduler controls, transition rules, and workflow approver settings.",
        "Workflow APIs implement the IMS daily control sequence: Investment Day Start, Manager Approval, Transaction Closing, and Accounting Closing. Investment ledger and decision submission checks depend on this state.",
        "All workflow endpoints require Bearer JWT through the authenticated `/api/v1` route group.",
        "Legacy UUID day-state reads require `WORKFLOW_VIEW`; scheduler requires `WORKFLOW_RUN_SCHEDULER`; legacy transition execution resolves per-action workflow permissions in the handler. Business-readable daily endpoints use settings-based approver/admin checks instead of `RequirePermission` middleware.",
        eps(WORKFLOW, "backend/internal/workflow/transport/router.go"),
        ["backend/internal/workflow/transport/router.go", "backend/internal/workflow/transport/handler/workflow_handler.go", "backend/internal/workflow/transport/handler/daily_handler.go", "backend/internal/workflow/permission/policies.go", "backend/docs/swagger.json"],
    )

    investment_eps = eps(INVESTMENT, "backend/internal/investment/module.go")
    investment_fallbacks = {
        ("GET", "/investment/executions"): ("List Trade Executions", None, [("decision_id", "uuid", "No", "Decision UUID"), ("contract_id", "uuid", "No", "Contract/fund UUID"), ("business_date", "string", "No", "Business date; required with contract_id")], [("200", "OK", "Execution list object"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("POST", "/investment/executions"): ("Create Trade Execution", {"decision_id": "uuid", "ordered_quantity": "100.00", "ordered_amount": "100000.00", "broker_reference": "BRK-001"}, None, [("201", "Created", "ExecutionResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("GET", "/investment/executions/{id}"): ("Get Trade Execution", None, None, [("200", "OK", "ExecutionResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("POST", "/investment/executions/{id}/fill"): ("Fill Trade Execution", {"executed_quantity": "100.00", "executed_amount": "100000.00", "execution_price": "1000.00", "status": "EXECUTED", "broker_reference": "BRK-001"}, None, [("200", "OK", "ExecutionResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("POST", "/investment/executions/{id}/cancel"): ("Cancel Trade Execution", {"reason": "string"}, None, [("200", "OK", "ExecutionResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("GET", "/investment/trade-confirmations"): ("List Trade Confirmations", None, [("execution_id", "uuid", "No", "Execution UUID"), ("contract_id", "uuid", "No", "Contract/fund UUID"), ("business_date", "string", "No", "Business date; required with contract_id")], [("200", "OK", "Trade confirmation list object"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("POST", "/investment/trade-confirmations"): ("Record Trade Confirmation", {"execution_id": "uuid", "confirmed_quantity": "100.00", "confirmed_amount": "100000.00", "confirmed_price": "1000.00", "broker_reference": "BRK-001", "import_batch_id": "uuid"}, None, [("201", "Created", "TradeConfirmationResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("422", "Unprocessable Entity", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("POST", "/investment/trade-confirmations/batch"): ("Import Trade Confirmation Batch", {"source_filename": "broker-confirmations.json", "rows": [{"execution_id": "uuid", "confirmed_quantity": "100.00", "confirmed_amount": "100000.00", "confirmed_price": "1000.00", "broker_reference": "BRK-001"}]}, None, [("201", "Created", "ConfirmationBatchImportResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("GET", "/investment/trade-confirmations/{id}"): ("Get Trade Confirmation", None, None, [("200", "OK", "TradeConfirmationResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
        ("POST", "/investment/trade-confirmations/{id}/resolve"): ("Resolve Trade Confirmation", {"target_status": "MATCHED", "discrepancy_reason": "string"}, None, [("200", "OK", "TradeConfirmationResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("422", "Unprocessable Entity", "ErrorResponse"), ("500", "Internal Server Error", "ErrorResponse")]),
    }
    for item in investment_eps:
        fallback = investment_fallbacks.get((item.method, item.path))
        if fallback and not item.op:
            item.purpose, item.body_sample, item.query, item.responses = fallback
        if item.path == "/investment/decisions/{id}" and item.method == "GET" and not item.op:
            item.purpose = "Get Investment Decision"
            item.responses = [("200", "OK", "DecisionResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse")]
        if item.path == "/investment/decisions/{id}" and item.method == "PUT" and not item.op:
            item.purpose = "Update Investment Decision"
            item.body_schema = "UpdateDecisionRequest"
            item.body_sample = {"fund_id": "uuid", "portfolio_id": "uuid", "instrument_id": "uuid", "side": "BUY", "quantity": "100.00", "amount": "100000.00", "business_date": "2026-06-26", "decision_type": "SINGLE_ORDER", "research_report_id": "uuid"}
            item.responses = [("200", "OK", "DecisionResponse"), ("400", "Bad Request", "ErrorResponse"), ("401", "Unauthorized", "ErrorResponse"), ("403", "Forbidden", "ErrorResponse"), ("404", "Not Found", "ErrorResponse"), ("409", "Conflict", "ErrorResponse"), ("422", "Unprocessable Entity", "ErrorResponse")]
    write_doc(
        "investment-api.md",
        "Investment API",
        "Provides APIs for fund and instrument master data, research analysis reports, investment decisions, executions, trade confirmations, fund valuation, AUM, and live market-data overlays.",
        "Investment APIs support the IMS investment process: Investment Research / Analysis Report, Investment Decision, Investment Execution, Trade Confirmation, and Investment Review. Decision submission integrates workflow gating, IRG pre-trade checks, approval workflows, and audit logging.",
        "All investment endpoints require Bearer JWT through `/api/v1`.",
        "Each route group is gated with an investment permission code such as `INVESTMENT_RESEARCH_SUBMIT`, `INVESTMENT_DECISION_SUBMIT`, `INVESTMENT_EXECUTION_MANAGE`, or `INVESTMENT_CONFIRMATION_IMPORT`.",
        investment_eps,
        ["backend/internal/investment/module.go", "backend/internal/investment/transport/handler", "backend/internal/investment/transport/dto/request/requests.go", "backend/internal/investment/transport/dto/response/responses.go", "backend/internal/investment/application/command", "backend/internal/investment/permission/policies.go", "backend/docs/swagger.json"],
        "Portfolio master, holdings, cash, ledger, and portfolio valuation endpoints are documented separately in `portfolio-api.md` even though they are mounted by the investment module.",
    )

    write_doc(
        "portfolio-api.md",
        "Portfolio API",
        "Provides portfolio master, holdings, cash balance, transaction ledger, simulation, reversal, and valuation APIs.",
        "Portfolio APIs support the portfolio and ledger side of Stock Investment Management. Ledger posts and simulations support investment execution/review and are constrained by workflow day state and, where applicable, IRG checks.",
        "All portfolio endpoints require Bearer JWT through `/api/v1`.",
        "Portfolio read/manage routes use `INVESTMENT_PORTFOLIO_VIEW` and `INVESTMENT_PORTFOLIO_MANAGE`; ledger and valuation routes use `INVESTMENT_LEDGER_POST`, `INVESTMENT_LEDGER_SIMULATE`, `INVESTMENT_LEDGER_REVERSE`, `INVESTMENT_VALUATION_VIEW`, and `INVESTMENT_VALUATION_RUN`.",
        eps(PORTFOLIO, "backend/internal/investment/module.go"),
        ["backend/internal/investment/module.go", "backend/internal/investment/transport/handler/investment_handler.go", "backend/internal/investment/application/command/post_transaction.go", "backend/internal/investment/application/command/reverse_transaction.go", "backend/internal/investment/application/service/valuation_runner.go", "backend/internal/investment/permission/policies.go", "backend/docs/swagger.json"],
    )

    write_doc(
        "compliance-api.md",
        "Compliance API",
        "Provides IRG rule execution, compliance check-group lookup, breach listing, breach override, and rule instance administration.",
        "Compliance APIs support pre-trade validation before Investment Decision approval and post-trade validation during workflow Transaction Closing. BLOCK/WARN results become audit-ready check records and breaches.",
        "All compliance endpoints require Bearer JWT through `/api/v1`.",
        "Route groups enforce `WORKFLOW_EXECUTE`, `IRG_VIEW_RULES`, `IRG_OVERRIDE_BREACH`, and `IRG_EDIT_RULE_INSTANCE`.",
        eps(COMPLIANCE, "backend/internal/compliance/module.go; backend/internal/compliance/transport/router.go"),
        ["backend/internal/compliance/module.go", "backend/internal/compliance/transport/router.go", "backend/internal/compliance/transport/handler/compliance_handler.go", "backend/internal/compliance/application/command", "backend/docs/swagger.json"],
        "Pre-trade and post-trade HTTP endpoints are permission-gated with `WORKFLOW_EXECUTE`. In-process compliance checks invoked by Investment or Workflow use contract adapters and do not pass through HTTP permission middleware.",
    )

    write_doc(
        "market-data-api.md",
        "Market Data API",
        "Provides market quote, historical price, import, import-batch, provider health, and frontend market-data screen endpoints.",
        "Market data supports investment valuation, intraday holdings views, and market-data refresh workflows. Investment uses this module through a contract quote provider for intraday valuation overlays.",
        "All market-data endpoints are mounted inside authenticated `/api/v1`; Bearer JWT is required.",
        "Permission rule not found in code. `backend/internal/market_data/permission/policies.go` declares no operator-grantable permission codes and the router does not apply `RequirePermission` middleware.",
        eps(MARKET_DATA, "backend/internal/market_data/transport/http/router.go"),
        ["backend/internal/market_data/module.go", "backend/internal/market_data/transport/http/router.go", "backend/internal/market_data/transport/http/handler.go", "backend/internal/market_data/permission/policies.go", "backend/docs/swagger.json"],
    )

    write_doc(
        "notification-api.md",
        "Notification API",
        "Provides in-app notification-center endpoints and email outbox administration endpoints.",
        "Notifications are emitted by approval, workflow stuck-day monitoring, and watchlist alert flows. Email outbox APIs let operators inspect and retry delivery records without exposing SMTP secrets.",
        "All notification endpoints require Bearer JWT. Email admin routes are mounted only when the email handler and permission checker are configured.",
        "User-facing notification-center routes have no function-permission middleware. Email outbox routes enforce `NOTIFICATION_VIEW`, `NOTIFICATION_RETRY`, `NOTIFICATION_TEST`, or `NOTIFICATION_HEALTH`.",
        eps(NOTIFICATION, "backend/internal/notification/transport/router.go"),
        ["backend/internal/notification/module.go", "backend/internal/notification/transport/router.go", "backend/internal/notification/transport/handler/notification_handler.go", "backend/internal/notification/transport/handler/email_outbox_handler.go", "backend/internal/notification/permission/policies.go", "backend/docs/swagger.json"],
    )

    template = """# {Module Name} API

## Purpose
Explain the API module in business terms. State the backend source package and the base path.

## Business Context
Explain which IMS business process this API supports. Preserve workflow, investment-process, permission-management, approval/delegation, and IRG/compliance terms when applicable.

## Authentication
State whether Bearer JWT authentication is required. Identify any public endpoints, active-session checks, IP allowlist, or rate limits found in code.

## Authorization / Permission
List required permission code, role, or access rule found in route middleware or handler logic. If no rule is found, write: "Permission rule not found in code."

## Endpoint Summary

| Method | Path | Purpose | Auth | Permission | Status |
|---|---|---|---|---|---|
| GET | `/example` | Example purpose | JWT required | `EXAMPLE_VIEW` | Implemented |

## Endpoint Detail

### {METHOD} {PATH}

#### Purpose
Explain what this endpoint does.

#### Business Rule
Explain workflow, approval, permission, compliance, or validation rules involved. If no business rule is visible in code, write "Business rule not found in code."

#### Request

##### Path Parameters
| Name | Type | Required | Description |
|---|---|---|---|
| id | UUID string | Yes | Business object ID |

##### Query Parameters
| Name | Type | Required | Description |
|---|---|---|---|
| page | integer | No | Page number |

##### Request Body
Schema: `{RequestDtoName}`.
```json
{}
```

#### Response
| HTTP Status | Description | Schema |
|---|---|---|
| 200 | OK | `{ResponseDtoName}` |
| 400 | Bad request / validation error | `ErrorResponse` |
| 401 | Missing or invalid JWT | `ErrorResponse` |
| 403 | Missing permission or access rule | `ErrorResponse` |

#### Source
`backend/internal/{module}/transport/router.go`; handler, DTO, service, permission, and tests used to validate this endpoint.

## Source References
- `backend/internal/{module}/transport/router.go`
- `backend/internal/{module}/transport/handler/...`
- `backend/internal/{module}/transport/dto/...`
- `backend/internal/{module}/application/...`
- `backend/internal/{module}/permission/policies.go`
- `backend/docs/swagger.json` when annotations exist
"""
    (DOCS / "_template.md").write_text(clean(template), encoding="utf-8")

    readme_rows = [
        ("Authentication", "auth-api.md", "Implemented", "Login, refresh, MFA, sessions, current user"),
        ("IAM", "iam-api.md", "Implemented", "Admin account and session management"),
        ("Permissions", "permission-api.md", "Implemented", "Account, Group, Function Permission, Data Permission, and change-request governance"),
        ("Approval", "approval-api.md", "Implemented", "Generic maker-checker runtime, delegation, groups, teams, process config"),
        ("Workflow", "workflow-api.md", "Implemented", "Investment Day Start, Manager Approval, Transaction Closing, Accounting Closing"),
        ("Investment", "investment-api.md", "Implemented", "Research, decision, execution, confirmation, fund/instrument APIs"),
        ("Compliance / IRG", "compliance-api.md", "Implemented", "Pre-trade/post-trade checks, breaches, overrides, rule instances"),
        ("Portfolio", "portfolio-api.md", "Implemented", "Portfolio master, holdings, cash, ledger, valuation"),
        ("Market Data", "market-data-api.md", "Implemented", "Quotes, history, imports, provider health, screen DTOs"),
        ("Notification", "notification-api.md", "Implemented", "In-app notifications and email outbox admin"),
        ("Audit", "audit-api.md", "Implemented", "System audit and permission-governance audit feeds"),
    ]
    readme = f"""# IMS API Documentation

## Purpose
This directory contains maintainable Markdown API documentation for the IMS backend. The documents are grounded in Go route registration, handlers, DTOs, services, permission middleware, and generated Swagger/OpenAPI metadata where annotations exist.

## Business Context
IMS is an Investment Management System. The API surface supports daily workflow control, stock investment management, permission governance, maker-checker approval, IRG / Compliance validation, portfolio ledger and valuation, market data, notifications, IAM, and audit.

## Documentation Index
{table(["Module", "Document", "Status", "Scope"], [(m, f"[{doc}]({doc})", s, scope) for m, doc, s, scope in readme_rows])}

## Not Implemented Yet
No requested API module is marked "Not implemented yet" in this repository. All required module files above have matching backend routes.

## Authentication Baseline
Most APIs are mounted under `/api/v1` behind Bearer JWT authentication. `/auth/login` and `/auth/refresh` are public but rate-limited. IAM admin and audit admin endpoints additionally inherit admin IP allowlist and admin/export rate limits.

## Authorization Baseline
Route-level authorization is documented from `middleware.RequirePermission(...)` calls where present. If a route is authenticated but no function-permission middleware or handler rule is visible, the endpoint states: "Permission rule not found in code."

## Automation Workflow
1. Update Go Swagger annotations when handler DTOs or routes change.
2. Regenerate backend Swagger with the repository's existing Swagger command (`make swagger` when available).
3. Run `python docs/api/_build_api_docs.py` from the repository root.
4. Compare route registration against `backend/docs/swagger.json`; router registration remains the source of truth when Swagger annotations are missing.
5. Preserve business terms from `docs/investment-module.md`, `docs/compliance-module.md`, `docs/handoff/workflow-backend.md`, and `docs/handoff/approval-module.md`.

## Source References
- `backend/cmd/server/main.go`
- `backend/docs/swagger.json`
- `backend/internal/*/module.go`
- `backend/internal/*/transport/router.go`
- `backend/internal/*/transport/handler/*.go`
- `backend/internal/*/transport/dto/**/*.go`
- `backend/internal/*/permission/policies.go`
- `docs/investment-module.md`
- `docs/compliance-module.md`
- `docs/handoff/workflow-backend.md`
- `docs/handoff/approval-module.md`
"""
    (DOCS / "README.md").write_text(clean(readme), encoding="utf-8")


if __name__ == "__main__":
    main()
