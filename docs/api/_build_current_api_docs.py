"""Build the current-code API reference and missing module documents.

Router registration is the endpoint source of truth. The V1 and V2 Swagger
artifacts provide descriptions, parameters, and schemas where annotations
exist. V1 routes that are registered but absent from Swagger are supplemented
from the repository's existing router-backed documentation inventory.
"""

from __future__ import annotations

import importlib.util
import json
import sys
from dataclasses import dataclass
from datetime import date
from pathlib import Path
from typing import Any


ROOT = Path(__file__).resolve().parents[2]
DOCS = ROOT / "docs" / "api"
V1_SWAGGER_PATH = ROOT / "backend" / "docs" / "swagger.json"
V2_SWAGGER_PATH = ROOT / "backend" / "docs" / "v2" / "v2_swagger.json"
LEGACY_BUILDER_PATH = DOCS / "_build_api_docs.py"
METHODS = {"get", "post", "put", "patch", "delete"}
METHOD_ORDER = {"GET": 0, "POST": 1, "PUT": 2, "PATCH": 3, "DELETE": 4}

# Loading the legacy inventory is read-only; keep regeneration from leaving a
# docs/api/__pycache__ directory in the working tree.
sys.dont_write_bytecode = True


def load_json(path: Path) -> dict[str, Any]:
    return json.loads(path.read_text(encoding="utf-8"))


def load_legacy_builder():
    spec = importlib.util.spec_from_file_location("ims_legacy_api_docs", LEGACY_BUILDER_PATH)
    if spec is None or spec.loader is None:
        raise RuntimeError(f"cannot load {LEGACY_BUILDER_PATH}")
    module = importlib.util.module_from_spec(spec)
    spec.loader.exec_module(module)
    return module


V1 = load_json(V1_SWAGGER_PATH)
V2 = load_json(V2_SWAGGER_PATH)
LEGACY = load_legacy_builder()


def text(value: object) -> str:
    return "" if value is None else str(value)


def md_cell(value: object) -> str:
    return text(value).replace("\n", " ").replace("|", "\\|")


def table(headers: list[str], rows: list[list[object] | tuple[object, ...]]) -> str:
    lines = [
        "| " + " | ".join(headers) + " |",
        "| " + " | ".join("---" for _ in headers) + " |",
    ]
    if not rows:
        rows = [["None"] + [""] * (len(headers) - 1)]
    lines.extend("| " + " | ".join(md_cell(cell) for cell in row) + " |" for row in rows)
    return "\n".join(lines)


def operation_map(spec: dict[str, Any], *, exclude_health: bool = False) -> dict[tuple[str, str], dict[str, Any]]:
    result: dict[tuple[str, str], dict[str, Any]] = {}
    for path, path_item in spec.get("paths", {}).items():
        if exclude_health and path == "/health":
            continue
        for method, operation in path_item.items():
            if method in METHODS:
                result[(method.upper(), path)] = operation
    return result


V1_OPERATIONS = operation_map(V1, exclude_health=True)
V2_OPERATIONS = operation_map(V2)


UNANNOTATED_SUMMARIES = {
    ("GET", "/audit/logs"): "List permission-management audit logs",
    ("GET", "/audit/logs/export"): "Export permission-management audit logs as CSV",
    ("GET", "/investment/decisions/{id}"): "Get Investment Decision",
    ("PUT", "/investment/decisions/{id}"): "Update Investment Decision",
    ("GET", "/investment/executions"): "List Trade Executions",
    ("GET", "/investment/executions/{id}"): "Get Trade Execution",
    ("POST", "/investment/executions/{id}/cancel"): "Cancel Trade Execution",
    ("POST", "/investment/executions/{id}/fill"): "Fill Trade Execution",
    ("GET", "/investment/trade-confirmations"): "List Trade Confirmations",
    ("POST", "/investment/trade-confirmations"): "Record Trade Confirmation",
    ("POST", "/investment/trade-confirmations/batch"): "Import Trade Confirmation Batch",
    ("GET", "/investment/trade-confirmations/{id}"): "Get Trade Confirmation",
    ("POST", "/investment/trade-confirmations/{id}/resolve"): "Resolve Trade Confirmation",
}


def module_for(path: str) -> str:
    if path.startswith("/auth"):
        return "Authentication"
    if path.startswith("/admin/users") or path.startswith("/admin/sessions"):
        return "IAM administration"
    if path.startswith("/admin/audit") or path.startswith("/audit/logs"):
        return "Audit"
    if path.startswith("/permissions"):
        return "Permissions"
    if path.startswith("/approval"):
        return "Approval"
    if path.startswith("/workflow"):
        return "Workflow"
    if path.startswith("/compliance"):
        return "Compliance"
    if path.startswith("/investment"):
        return "Investment V1"
    if path.startswith("/market-data"):
        return "Market Data"
    if path.startswith("/reference-data"):
        return "Reference Data"
    if path.startswith("/integration"):
        return "Integration"
    if path.startswith("/notifications"):
        return "Notification"
    if path.startswith("/watchlists"):
        return "Watchlist"
    if path.startswith("/chat"):
        return "Chat"
    if path.startswith("/portfolios"):
        return "Portfolio V2"
    return "Other"


SOURCE_BY_MODULE = {
    "Authentication": "backend/internal/iam/module.go",
    "IAM administration": "backend/internal/iam/module.go",
    "Audit": "backend/internal/audit/module.go; backend/internal/permissions/module.go",
    "Permissions": "backend/internal/permissions/module.go",
    "Approval": "backend/internal/approval/transport/router.go",
    "Workflow": "backend/internal/workflow/transport/router.go",
    "Compliance": "backend/internal/compliance/module.go",
    "Investment V1": "backend/internal/investment/module.go",
    "Market Data": "backend/internal/market_data/transport/http/router.go",
    "Reference Data": "backend/internal/reference_data/transport/router.go",
    "Integration": "backend/internal/integration/transport/router.go",
    "Notification": "backend/internal/notification/transport/router.go",
    "Watchlist": "backend/internal/watchlist/transport/http/router.go",
    "Chat": "backend/internal/chat/transport/router.go",
    "Portfolio V2": "backend/internal/investment/module.go:RegisterRoutesV2",
}


def build_permission_map() -> dict[tuple[str, str], str]:
    permissions: dict[tuple[str, str], str] = {}
    for name in [
        "IAM",
        "AUDIT",
        "APPROVAL",
        "WORKFLOW",
        "COMPLIANCE",
        "INVESTMENT",
        "PORTFOLIO",
        "NOTIFICATION",
    ]:
        permissions.update(getattr(LEGACY, name))

    for endpoint in LEGACY.permission_endpoints():
        permissions[(endpoint.method, endpoint.path)] = endpoint.permission

    permissions.update(
        {
            ("GET", "/audit/logs"): "permission.audit.view",
            ("GET", "/audit/logs/export"): "permission.audit.export",
            ("GET", "/integration/dashboard/me"): "INTEGRATION_DASHBOARD_VIEW (application check)",
            ("GET", "/integration/tasks/my"): "INTEGRATION_DASHBOARD_VIEW (application check)",
            ("GET", "/integration/tasks/my/summary"): "INTEGRATION_DASHBOARD_VIEW (application check)",
            ("GET", "/integration/dashboard/valuation-summary"): "Authenticated; company aggregate is permission-independent, mine is data-scoped",
            ("GET", "/watchlists"): "WATCHLIST_VIEW + owner/portfolio data scope",
            ("POST", "/watchlists/items"): "WATCHLIST_MANAGE + owner/portfolio data scope",
            ("PATCH", "/watchlists/items/{id}"): "WATCHLIST_MANAGE + owner/portfolio data scope",
            ("DELETE", "/watchlists/items/{id}"): "WATCHLIST_MANAGE + owner/portfolio data scope",
            ("GET", "/watchlists/alerts"): "WATCHLIST_VIEW + owner/portfolio data scope",
            ("POST", "/watchlists/alerts/{id}/acknowledge"): "WATCHLIST_ALERT_ACK + owner/portfolio data scope",
            ("POST", "/watchlists/evaluate"): "WATCHLIST_EVALUATE + optional portfolio data scope",
            ("POST", "/chat"): "Authenticated; each MCP tool is permission-gated; writes also require CHAT_WRITE_ENABLED",
            ("GET", "/chat/sessions"): "Owner-scoped",
            ("GET", "/chat/sessions/{session_id}"): "Owner-scoped; IAM_AUDIT_VIEW may read with audit trail",
            ("GET", "/chat/sessions/{session_id}/messages"): "Owner-scoped; IAM_AUDIT_VIEW may read with audit trail",
        }
    )

    for key in V1_OPERATIONS:
        path = key[1]
        if path.startswith("/reference-data") or path.startswith("/market-data"):
            if not permissions.get(key):
                permissions[key] = "Authenticated; no function-permission middleware"
        elif path in {"/notifications", "/notifications/read-all"} or (
            path.startswith("/notifications/{id}")
        ):
            if not permissions.get(key):
                permissions[key] = "Authenticated; current-user scope"
        elif path.startswith("/auth"):
            if not permissions.get(key):
                permissions[key] = "Public" if path in {"/auth/login", "/auth/refresh"} else "Authenticated"

    return permissions


V1_PERMISSIONS = build_permission_map()
V2_PERMISSIONS = {
    ("GET", "/portfolios/{portfolioCode}"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/portfolios/{portfolioCode}/holdings"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/portfolios/{portfolioCode}/cash"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/portfolios/{portfolioCode}/transactions"): "INVESTMENT_PORTFOLIO_VIEW",
    ("GET", "/portfolios/{portfolioCode}/valuations"): "INVESTMENT_VALUATION_VIEW",
    ("GET", "/portfolios/{portfolioCode}/valuations/latest"): "INVESTMENT_VALUATION_VIEW",
    ("POST", "/portfolios/{portfolioCode}/transactions/simulate"): "INVESTMENT_LEDGER_SIMULATE",
    ("POST", "/portfolios/{portfolioCode}/transactions"): "INVESTMENT_LEDGER_POST",
    ("POST", "/portfolios/{portfolioCode}/transactions/{transactionId}/reverse"): "INVESTMENT_LEDGER_REVERSE",
    ("GET", "/portfolios/{portfolioCode}/decisions"): "INVESTMENT_DECISION_VIEW",
    ("GET", "/portfolios/{portfolioCode}/decisions/{decisionId}"): "INVESTMENT_DECISION_VIEW",
    ("POST", "/portfolios/{portfolioCode}/decisions"): "INVESTMENT_DECISION_MANAGE",
    ("POST", "/portfolios/{portfolioCode}/decisions/{decisionId}/submit"): "INVESTMENT_DECISION_SUBMIT",
    ("POST", "/portfolios/{portfolioCode}/decisions/{decisionId}/cancel"): "INVESTMENT_DECISION_CANCEL",
    ("POST", "/portfolios/{portfolioCode}/decisions/{decisionId}/executions"): "INVESTMENT_EXECUTION_MANAGE",
    ("POST", "/portfolios/{portfolioCode}/executions/{executionId}/fill"): "INVESTMENT_EXECUTION_MANAGE",
    ("POST", "/portfolios/{portfolioCode}/executions/{executionId}/cancel"): "INVESTMENT_EXECUTION_MANAGE",
    ("POST", "/portfolios/{portfolioCode}/executions/{executionId}/confirmations"): "INVESTMENT_CONFIRMATION_MANAGE",
    ("POST", "/portfolios/{portfolioCode}/confirmations/{confirmationId}/resolve"): "INVESTMENT_CONFIRMATION_MANAGE",
    ("GET", "/portfolios/{portfolioCode}/compliance/rules"): "IRG_VIEW_RULES",
    ("GET", "/portfolios/{portfolioCode}/compliance/breaches"): "IRG_VIEW_RULES",
    ("POST", "/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings"): "IRG_EDIT_BINDING",
    ("DELETE", "/portfolios/{portfolioCode}/compliance/rules/{ruleInstanceID}/bindings/{bindingID}"): "IRG_EDIT_BINDING",
    ("POST", "/portfolios/{portfolioCode}/compliance/checks/pre-trade"): "WORKFLOW_EXECUTE",
    ("POST", "/portfolios/{portfolioCode}/compliance/checks/post-trade"): "WORKFLOW_EXECUTE",
}


@dataclass(frozen=True)
class Route:
    version: str
    method: str
    path: str
    summary: str
    permission: str
    source: str
    status: str = "Implemented"

    @property
    def full_path(self) -> str:
        if self.version == "v1":
            return "/api/v1" + self.path
        if self.version == "v2":
            return "/api/v2" + self.path
        return self.path

    @property
    def module(self) -> str:
        return module_for(self.path)


def build_v1_routes() -> list[Route]:
    registered: set[tuple[str, str]] = set(V1_OPERATIONS)
    for name in [
        "AUTH",
        "IAM",
        "AUDIT",
        "APPROVAL",
        "WORKFLOW",
        "COMPLIANCE",
        "INVESTMENT",
        "PORTFOLIO",
        "MARKET_DATA",
        "NOTIFICATION",
    ]:
        registered.update(getattr(LEGACY, name))
    registered.update((endpoint.method, endpoint.path) for endpoint in LEGACY.permission_endpoints())
    registered.update({("GET", "/audit/logs"), ("GET", "/audit/logs/export")})

    permission_endpoints = {
        (endpoint.method, endpoint.path): endpoint for endpoint in LEGACY.permission_endpoints()
    }
    routes: list[Route] = []
    for method, path in registered:
        operation = V1_OPERATIONS.get((method, path), {})
        summary = operation.get("summary") or UNANNOTATED_SUMMARIES.get((method, path), "")
        if not summary and (method, path) in permission_endpoints:
            summary = permission_endpoints[(method, path)].purpose
        if not summary:
            summary = "Registered API operation"
        module = module_for(path)
        status = "Conditionally mounted when the chat provider initializes" if module == "Chat" else "Implemented"
        routes.append(
            Route(
                version="v1",
                method=method,
                path=path,
                summary=summary,
                permission=V1_PERMISSIONS.get((method, path), "Authenticated; route-specific rule not found"),
                source=SOURCE_BY_MODULE[module],
                status=status,
            )
        )
    return sorted(routes, key=lambda route: (route.module, route.path, METHOD_ORDER.get(route.method, 99)))


def build_v2_routes() -> list[Route]:
    routes = []
    for (method, path), operation in V2_OPERATIONS.items():
        routes.append(
            Route(
                version="v2",
                method=method,
                path=path,
                summary=operation.get("summary", "Registered API operation"),
                permission=V2_PERMISSIONS.get((method, path), "Authenticated; route-specific rule not found"),
                source=SOURCE_BY_MODULE["Portfolio V2"],
            )
        )
    return sorted(routes, key=lambda route: (route.path, METHOD_ORDER.get(route.method, 99)))


V1_ROUTES = build_v1_routes()
V2_ROUTES = build_v2_routes()


def schema_name(schema: dict[str, Any] | None) -> str:
    if not schema:
        return ""
    if "$ref" in schema:
        return schema["$ref"].split("/")[-1]
    if schema.get("type") == "array":
        return f"array[{schema_name(schema.get('items')) or 'object'}]"
    if "allOf" in schema:
        data_schema = None
        names = []
        for part in schema["allOf"]:
            name = schema_name(part)
            if name:
                names.append(name)
            data_schema = (part.get("properties") or {}).get("data") or data_schema
        if data_schema:
            return f"SuccessResponse<{schema_name(data_schema) or 'object'}>"
        return " + ".join(names) or "object"
    if schema.get("type"):
        return schema["type"]
    if schema.get("properties"):
        return "object"
    return ""


def schema_sample(schema: dict[str, Any] | None, definitions: dict[str, Any], depth: int = 0):
    if not schema:
        return None
    if "$ref" in schema:
        name = schema["$ref"].split("/")[-1]
        if depth >= 3 or name not in definitions:
            return f"<{name}>"
        return schema_sample(definitions[name], definitions, depth + 1)
    if schema.get("type") == "array":
        return [schema_sample(schema.get("items"), definitions, depth + 1) or "<item>"]
    if "allOf" in schema:
        merged: dict[str, Any] = {}
        for part in schema["allOf"]:
            value = schema_sample(part, definitions, depth + 1)
            if isinstance(value, dict):
                merged.update(value)
        return merged
    if schema.get("type") == "object" or schema.get("properties"):
        result = {}
        for key, prop in list((schema.get("properties") or {}).items())[:18]:
            result[key] = schema_sample(prop, definitions, depth + 1)
        return result
    if schema.get("type") in {"integer", "number"}:
        return schema.get("example", 0)
    if schema.get("type") == "boolean":
        return schema.get("example", False)
    if schema.get("type") == "string":
        if schema.get("example") is not None:
            return schema["example"]
        return (schema.get("enum") or ["string"])[0]
    return "<value>"


def body_parameter(operation: dict[str, Any]) -> dict[str, Any] | None:
    return next((param for param in operation.get("parameters", []) if param.get("in") == "body"), None)


def response_schema_label(code: str, response: dict[str, Any], *, success_wrapped: bool, sse: bool) -> str:
    label = schema_name(response.get("schema"))
    if sse and code.startswith("2"):
        return f"text/event-stream frames of {label or 'ChatStreamEvent'}"
    if success_wrapped and code in {"200", "201", "202"} and not label.startswith("SuccessResponse<"):
        return f"SuccessResponse<{label or 'object'}>"
    return label


def route_rule(module: str, path: str) -> str:
    if path == "/integration/dashboard/valuation-summary":
        return (
            "`scope=company` is the default and returns a company-wide aggregate to any authenticated user without "
            "item-level exclusions. `scope=mine` includes portfolios managed by the caller inside accessible funds. "
            "Only LIVE portfolios contribute to official totals."
        )
    if module == "Integration":
        return "Dashboard and task endpoints resolve the authenticated identity server-side and do not accept a user ID from the client."
    if module == "Reference Data":
        return "Canonical security identity and provider mappings are owned by Reference Data; deletes of provider mappings are soft deletes."
    if module == "Chat":
        return (
            "Chat is mounted only when its configured LLM provider initializes. Session reads are owner-scoped; "
            "auditor access requires `IAM_AUDIT_VIEW` and is audited. Tool execution is permission-gated, and mutating "
            "tools additionally require `CHAT_WRITE_ENABLED=true`."
        )
    if module == "Watchlist":
        return "Application services enforce personal ownership or portfolio data scope after the route-level Watchlist permission check."
    if module == "Portfolio V2":
        return (
            "The route resolves `portfolioCode` to the internal portfolio UUID. Portfolio identity is validated again "
            "for nested decisions, executions, confirmations, transactions, compliance records, and valuations."
        )
    return "Business rule is defined by the owning handler and application service."


@dataclass(frozen=True)
class ModuleDoc:
    filename: str
    title: str
    module: str
    version: str
    purpose: str
    auth: str
    authorization: str
    sources: tuple[str, ...]
    prefixes: tuple[str, ...]
    success_wrapped: bool = True
    notes: str = ""


MODULE_DOCS = (
    ModuleDoc(
        filename="integration-api.md",
        title="Integration API",
        module="Integration",
        version="v1",
        purpose="Provides authenticated dashboard snapshots, personal task feeds, task counts, and official company/mine AUM and P&L summaries.",
        auth="Bearer JWT is required for every endpoint. The caller identity comes from validated JWT claims.",
        authorization="Dashboard and task endpoints require `INTEGRATION_DASHBOARD_VIEW` in the application layer. The valuation summary is available to every authenticated user under the explicit aggregate-only policy documented below.",
        sources=(
            "backend/internal/integration/transport/router.go",
            "backend/internal/integration/transport/handler/dashboard_handler.go",
            "backend/internal/integration/application/query",
            "backend/internal/integration/permission/policies.go",
            "backend/pkg/contract/valuation_summary.go",
            "backend/docs/swagger.json",
        ),
        prefixes=("/integration",),
        notes="Company valuation coverage is aggregate-only and must not expose excluded fund or portfolio identities. Drill-down APIs keep their existing data-scope rules.",
    ),
    ModuleDoc(
        filename="reference-data-api.md",
        title="Reference Data API",
        module="Reference Data",
        version="v1",
        purpose="Manages canonical IMS securities, provider-symbol mappings, and the unmapped-symbol review queue used by market-data ingestion.",
        auth="Bearer JWT is required for every endpoint.",
        authorization="The current router has no function-permission middleware. The module's permission catalog is intentionally empty, so every authenticated user can currently call these routes.",
        sources=(
            "backend/internal/reference_data/transport/router.go",
            "backend/internal/reference_data/transport/handler/securities_handler.go",
            "backend/internal/reference_data/application/service.go",
            "backend/internal/reference_data/permission/policies.go",
            "backend/docs/swagger.json",
        ),
        prefixes=("/reference-data",),
        notes="The current handler wraps successful JSON payloads in `{\"data\": ...}` even where older Swagger annotations name only the inner DTO.",
    ),
    ModuleDoc(
        filename="chat-api.md",
        title="Chat API",
        module="Chat",
        version="v1",
        purpose="Provides the optional AI financial assistant stream and persisted conversation-history reads.",
        auth="Bearer JWT is required. The module is mounted only when the configured LLM provider initializes successfully.",
        authorization="Session reads are owner-scoped. `IAM_AUDIT_VIEW` permits audited cross-owner reads. Every MCP tool call passes a permission gate; mutating tools also require `CHAT_WRITE_ENABLED=true`.",
        sources=(
            "backend/internal/chat/module.go",
            "backend/internal/chat/transport/router.go",
            "backend/internal/chat/transport/handler/chat_handler.go",
            "backend/internal/chat/transport/handler/session_handler.go",
            "backend/internal/chat/application/service",
            "backend/docs/swagger.json",
        ),
        prefixes=("/chat",),
        notes="`POST /api/v1/chat` is Server-Sent Events. Use `fetch` plus a readable stream because browser `EventSource` cannot send the Authorization header.",
    ),
    ModuleDoc(
        filename="watchlist-current-api.md",
        title="Watchlist API - Current Implementation",
        module="Watchlist",
        version="v1",
        purpose="Manages personal and portfolio-scoped security watchlists, threshold rules, alert events, acknowledgement, and operator-triggered evaluation.",
        auth="Bearer JWT is required for every endpoint.",
        authorization="Routes require `WATCHLIST_VIEW`, `WATCHLIST_MANAGE`, `WATCHLIST_ALERT_ACK`, or `WATCHLIST_EVALUATE`. Application services additionally enforce personal ownership or portfolio data scope.",
        sources=(
            "backend/internal/watchlist/transport/http/router.go",
            "backend/internal/watchlist/transport/http/handler.go",
            "backend/internal/watchlist/application",
            "backend/internal/watchlist/permission/policies.go",
            "backend/docs/swagger.json",
        ),
        prefixes=("/watchlists",),
        notes="This is the live route/DTO reference. `watchlist-api.md` remains the original design contract and contains future-policy discussion that is not automatically part of the implementation.",
    ),
    ModuleDoc(
        filename="portfolio-v2-api.md",
        title="Portfolio V2 API - Current Implementation",
        module="Portfolio V2",
        version="v2",
        purpose="Provides the implemented portfolio-code API for portfolio reads, ledger operations, valuations, decisions, executions, confirmations, and portfolio-scoped compliance.",
        auth="Bearer JWT is required for every endpoint under `/api/v2`.",
        authorization="Every route group applies the investment, workflow, or IRG permission shown in the endpoint table. Existing data-scope and aggregate-membership checks continue in the application and repository path.",
        sources=(
            "backend/cmd/server/main.go",
            "backend/internal/investment/module.go:RegisterRoutesV2",
            "backend/internal/investment/transport/handler/portfolio_v2_handler.go",
            "backend/internal/investment/transport/handler/portfolio_v2_ledger_handler.go",
            "backend/internal/investment/transport/handler/portfolio_v2_decision_handler.go",
            "backend/internal/investment/transport/handler/portfolio_v2_execution_handler.go",
            "backend/internal/investment/transport/handler/portfolio_v2_compliance_handler.go",
            "backend/docs/v2/v2_swagger.json",
        ),
        prefixes=("/portfolios",),
        notes="V2 is additive; V1 remains mounted. Only the 25 endpoints listed here are implemented. The broader `portfolio-v2-api-ddd.md` file is a design reference and includes routes that are still future scope.",
    ),
)


def routes_for_doc(config: ModuleDoc) -> list[Route]:
    routes = V1_ROUTES if config.version == "v1" else V2_ROUTES
    return [route for route in routes if any(route.path.startswith(prefix) for prefix in config.prefixes)]


def render_endpoint_detail(route: Route, spec: dict[str, Any], config: ModuleDoc) -> str:
    operation = operation_map(spec, exclude_health=config.version == "v1").get((route.method, route.path), {})
    description = operation.get("description", route.summary)
    parameters = operation.get("parameters", [])
    parameter_rows = []
    for param in parameters:
        if param.get("in") == "body":
            continue
        parameter_rows.append(
            [
                param.get("name", ""),
                param.get("in", ""),
                param.get("type") or schema_name(param.get("schema")) or "string",
                "Yes" if param.get("required") else "No",
                param.get("description", ""),
            ]
        )

    body = body_parameter(operation)
    if body:
        body_schema = body.get("schema", {})
        sample = schema_sample(body_schema, spec.get("definitions", {}))
        body_text = f"Schema: `{schema_name(body_schema) or 'object'}`.\n\n```json\n{json.dumps(sample, indent=2, ensure_ascii=False)}\n```"
    else:
        body_text = "No request body."

    sse = "text/event-stream" in operation.get("produces", [])
    response_rows = []
    for code, response in sorted(operation.get("responses", {}).items(), key=lambda item: str(item[0])):
        response_rows.append(
            [
                code,
                response.get("description", ""),
                response_schema_label(
                    str(code),
                    response,
                    success_wrapped=config.success_wrapped and not sse,
                    sse=sse,
                ),
            ]
        )

    return "\n".join(
        [
            f"### {route.method} `{route.full_path}`",
            "",
            description,
            "",
            f"**Permission:** {route.permission}",
            "",
            f"**Business rule:** {route_rule(config.module, route.path)}",
            "",
            "#### Parameters",
            "",
            table(["Name", "Location", "Type", "Required", "Description"], parameter_rows),
            "",
            "#### Request body",
            "",
            body_text,
            "",
            "#### Responses",
            "",
            table(["HTTP status", "Description", "Runtime schema"], response_rows),
            "",
            f"**Source:** `{route.source}`",
        ]
    )


def render_module_doc(config: ModuleDoc) -> str:
    routes = routes_for_doc(config)
    spec = V1 if config.version == "v1" else V2
    summary_rows = [
        [route.method, f"`{route.full_path}`", route.summary, route.permission, route.status]
        for route in routes
    ]
    details = "\n\n".join(render_endpoint_detail(route, spec, config) for route in routes)
    sources = "\n".join(f"- `{source}`" for source in config.sources)
    base_path = "/api/v1" if config.version == "v1" else "/api/v2"
    content = [
        f"# {config.title}",
        "",
        f"**Status:** Current implementation as of {date.today().isoformat()}",
        f"**Base path:** `{base_path}`",
        f"**Endpoint count:** {len(routes)}",
        "",
        "## Purpose",
        "",
        config.purpose,
        "",
        "## Authentication",
        "",
        config.auth,
        "",
        "## Authorization",
        "",
        config.authorization,
        "",
        "## Runtime response conventions",
        "",
        "Successful JSON handlers in this module use `{\"data\": ..., \"message\": ...}`; `message` is optional. "
        "A 204 response has no body. Error responses generally use `{\"error\": ..., \"code\": ..., \"details\": ...}`, "
        "although some newer typed-error paths use `{\"error_code\": ..., \"message\": ..., \"request_id\": ...}`. "
        "Clients should rely on the status and documented machine code when present, not exact human text.",
        "",
        "## Endpoint summary",
        "",
        table(["Method", "Path", "Purpose", "Authorization", "Status"], summary_rows),
        "",
        "## Endpoint details",
        "",
        details,
        "",
        "## Source references",
        "",
        sources,
    ]
    if config.notes:
        content.extend(["", "## Notes", "", config.notes])
    return "\n".join(content).strip() + "\n"


def render_system_doc() -> str:
    health_op = operation_map(V1).get(("GET", "/health"), {})
    health_responses = [
        [code, response.get("description", ""), schema_name(response.get("schema"))]
        for code, response in sorted(health_op.get("responses", {}).items())
    ]
    return (
        "# System and API Discovery Endpoints\n\n"
        f"**Status:** Current implementation as of {date.today().isoformat()}\n\n"
        "These endpoints are mounted outside `/api/v1` and `/api/v2`. They do not use Bearer authentication in the current router. "
        "Operators must restrict network access to `/metrics` at the deployment boundary.\n\n"
        "## Endpoint summary\n\n"
        + table(
            ["Method", "Path", "Purpose", "Authentication"],
            [
                ["GET", "`/health`", "Backend, PostgreSQL, Redis, and chat/MCP readiness", "Public"],
                ["GET", "`/metrics`", "Prometheus metrics scrape", "Public"],
                ["GET", "`/swagger/index.html`", "V1 Swagger UI", "Public"],
                ["GET", "`/swagger/doc.json`", "V1 Swagger 2.0 document", "Public"],
                ["GET", "`/swagger/v2/index.html`", "Portfolio V2 Swagger UI", "Public"],
                ["GET", "`/swagger/v2/doc.json`", "Portfolio V2 Swagger 2.0 document", "Public"],
            ],
        )
        + "\n\n## GET `/health`\n\n"
        + health_op.get("description", "Returns component readiness.")
        + "\n\n"
        + table(["HTTP status", "Description", "Schema"], health_responses)
        + "\n\nThe health payload contains component status only; it must not expose credentials, provider keys, model identifiers, or private endpoints.\n\n"
        "## GET `/metrics`\n\n"
        "Returns Prometheus exposition text from the shared metrics registry. The route is intentionally unauthenticated for scraping, so expose it only on a trusted network or behind infrastructure access controls.\n\n"
        "## OpenAPI caveat\n\n"
        "The V1 Swagger source contains a `/health` operation while declaring `basePath: /api/v1`. The live route is `/health`, not `/api/v1/health`. `/metrics` and the Swagger UI routes are not part of the API base path.\n\n"
        "## Source references\n\n"
        "- `backend/cmd/server/main.go`\n"
        "- `backend/cmd/server/swagger_v2_docs.go`\n"
        "- `backend/platform/health/health.go`\n"
        "- `backend/platform/metrics/registry.go`\n"
        "- `backend/docs/swagger.json`\n"
        "- `backend/docs/v2/v2_swagger.json`\n"
    )


def render_current_reference() -> str:
    swagger_v1_count = len(V1_OPERATIONS)
    missing_v1 = sorted(
        {(route.method, route.path) for route in V1_ROUTES} - set(V1_OPERATIONS),
        key=lambda item: (item[1], METHOD_ORDER.get(item[0], 99)),
    )
    module_rows = []
    for module in sorted({route.module for route in V1_ROUTES}):
        module_routes = [route for route in V1_ROUTES if route.module == module]
        status = "Conditional" if module == "Chat" else "Implemented"
        module_rows.append([module, "V1", len(module_routes), status, SOURCE_BY_MODULE[module]])
    module_rows.append(["Portfolio V2", "V2", len(V2_ROUTES), "Implemented", SOURCE_BY_MODULE["Portfolio V2"]])
    module_rows.append(["System/operations", "Root", 2, "Implemented", "backend/cmd/server/main.go"])

    sections = []
    for module in sorted({route.module for route in V1_ROUTES}):
        routes = [route for route in V1_ROUTES if route.module == module]
        sections.extend(
            [
                f"### {module}",
                "",
                table(
                    ["Method", "Path", "Purpose", "Authorization", "Status"],
                    [[route.method, f"`{route.full_path}`", route.summary, route.permission, route.status] for route in routes],
                ),
                "",
            ]
        )
    sections.extend(
        [
            "### Portfolio V2",
            "",
            table(
                ["Method", "Path", "Purpose", "Authorization", "Status"],
                [[route.method, f"`{route.full_path}`", route.summary, route.permission, route.status] for route in V2_ROUTES],
            ),
        ]
    )

    missing_rows = [[method, f"`/api/v1{path}`", module_for(path)] for method, path in missing_v1]
    content = [
        "# IMS Current API Reference",
        "",
        f"**Verified against workspace code:** {date.today().isoformat()}",
        "**Runtime source of truth:** `backend/cmd/server/main.go` plus each module's route registration",
        "**Payload metadata:** `backend/docs/swagger.json` and `backend/docs/v2/v2_swagger.json`",
        "",
        "## Scope and authority",
        "",
        "This document is the canonical catalog of routes mounted by the current server. Router registration wins when Markdown and Swagger disagree. Swagger supplies request/response metadata only where an operation is annotated.",
        "",
        f"The current server registers {len(V1_ROUTES)} possible V1 operations, {len(V2_ROUTES)} V2 operations, plus `/health` and `/metrics`. "
        "The four Chat operations are conditional: they are present only when the configured LLM provider initializes. Swagger UI and raw-spec routes are discovery endpoints and are listed separately in `system-api.md`.",
        "",
        "## Base URLs and interactive documentation",
        "",
        table(
            ["Surface", "Local URL", "Notes"],
            [
                ["V1 API", "`http://localhost:8080/api/v1`", "Primary modular-monolith API"],
                ["Portfolio V2 API", "`http://localhost:8080/api/v2`", "Additive portfolio-code routes"],
                ["V1 Swagger UI", "`http://localhost:8080/swagger/index.html`", "Swagger 2.0 V1 spec"],
                ["V2 Swagger UI", "`http://localhost:8080/swagger/v2/index.html`", "Separate Swagger 2.0 spec because V2 has a different base path"],
                ["Health", "`http://localhost:8080/health`", "Outside both API base paths"],
                ["Metrics", "`http://localhost:8080/metrics`", "Prometheus; restrict at the network boundary"],
            ],
        ),
        "",
        "## Authentication and request conventions",
        "",
        "`POST /api/v1/auth/login` and `POST /api/v1/auth/refresh` are public and rate-limited. Other V1 routes and every V2 route require `Authorization: Bearer <access-token>`. Authentication validates issuer, audience, expiry, active user status, and active server-side session when the token carries a session ID.",
        "",
        "JSON success handlers normally return `{\"data\": ..., \"message\": ...}` with optional `message`; 204 responses have no body. Legacy errors use `{\"error\": ..., \"code\": ..., \"details\": ...}` while typed domain-error paths use `{\"error_code\": ..., \"message\": ..., \"details\": ..., \"request_id\": ...}`. Some authentication middleware errors are JSON text written through `http.Error`, so clients must primarily use the HTTP status and parse machine codes when present.",
        "",
        "Money, prices, quantities, and percentages are generally decimal strings. Timestamps are RFC3339/UTC and business dates use `YYYY-MM-DD`. Pagination is endpoint-specific: older APIs use either `page`/`limit` or `offset`/`limit`, so use the parameters documented for that operation.",
        "",
        "## Module coverage",
        "",
        table(["Module", "Surface", "Operations", "Runtime status", "Route source"], module_rows),
        "",
        "## OpenAPI coverage gaps",
        "",
        f"The V1 Swagger artifact describes {swagger_v1_count} mounted V1 operations after excluding its misplaced `/health` entry. "
        f"The following {len(missing_v1)} registered V1 operations are not present in that Swagger artifact and therefore will not appear in generated frontend types unless annotations are added:",
        "",
        table(["Method", "Runtime path", "Module"], missing_rows),
        "",
        "The V2 Swagger artifact covers all 25 currently registered V2 routes. Its success schemas often name the inner DTO even though the handlers wrap JSON responses in `SuccessResponse`; `portfolio-v2-api.md` records the runtime envelope.",
        "",
        "## Endpoint catalog",
        "",
        *sections,
        "",
        "## Source references",
        "",
        "- `backend/cmd/server/main.go`",
        "- `backend/internal/*/module.go`",
        "- `backend/internal/*/transport/router.go`",
        "- `backend/internal/*/transport/http/router.go`",
        "- `backend/docs/swagger.json`",
        "- `backend/docs/v2/v2_swagger.json`",
        "- `docs/api/_build_api_docs.py` for registered V1 routes missing from Swagger",
        "",
        "## Regeneration",
        "",
        "Run `python -B docs/api/_build_current_api_docs.py` from the repository root after regenerating Swagger. The script writes only the current-code reference files listed in its `main` function and does not rewrite the legacy module documents.",
    ]
    return "\n".join(content).strip() + "\n"


def main() -> None:
    outputs = {
        "current-api-reference.md": render_current_reference(),
        "system-api.md": render_system_doc(),
    }
    for config in MODULE_DOCS:
        outputs[config.filename] = render_module_doc(config)
    for filename, content in outputs.items():
        (DOCS / filename).write_text(content, encoding="utf-8")
    print(f"Wrote {len(outputs)} current API documents to {DOCS}")


if __name__ == "__main__":
    main()
