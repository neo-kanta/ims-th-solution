//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval"
	approvaladapter "github.com/neo-kanta/ims-th-solution/backend/internal/approval/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment"
	marketdata "github.com/neo-kanta/ims-th-solution/backend/internal/market_data"
	referencedata "github.com/neo-kanta/ims-th-solution/backend/internal/reference_data"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/middleware"
)

// TestE2E_PortfolioComplianceV2_ManagerApprovalToExecution proves the P1
// manager-approval-to-execution compliance lifecycle
// (docs/MANAGER/TASKS.md "Prove Portfolio Compliance V2 End to End") against
// a live, dedicated ims_e2e Postgres database — never ims_dev:
//
//	create rule + parameters -> bind active rule to portfolio ->
//	create and submit an investment decision -> two real manager approval
//	stages (ben, then green, via the seeded PROC_DECISION_DEFAULT process:
//	database/seeds/012_approval_demo_seed.sql) -> execution creation re-runs
//	pre-trade compliance with the execution's actual ordered quantity.
//
// A single rule instance ("exposure.max_order_percent_aum", chosen because
// it is not one of the four rules already GLOBAL-bound by
// database/seeds/009_compliance_rule_seed.sql, so this test's portfolio-scope
// binding is the only thing gating order size against AUM) is bound and
// rebound across the subtests below to prove, against the same approved
// decision:
//   - BLOCK prevents execution, persists no execution row, and leaves an
//     auditable compliance_check_records + compliance_breaches trail.
//   - PASS (a small order that breaches nothing, evaluated against the same
//     active BLOCK binding) permits execution and leaves an auditable PASS
//     check record with no breach row.
//   - Deactivating the binding changes which rule applies: the oversized
//     order that BLOCKed above is no longer resolved as gated by
//     exposure.max_order_percent_aum at all.
//   - Binding the same rule at severity=WARN demonstrates the severity-cap
//     policy in valueobject.Severity.CapVerdict: the rule's raw BLOCK
//     verdict is capped to WARN, and WARN does not prevent execution
//     (matches docs/MANAGER/MEMORY.md "WARN currently does not block
//     execution").
//   - A binding whose effective_from is in the future does not apply today
//     — proving effective-dated scheduling, distinct from deactivation.
//
// Skips cleanly when IMS_TEST_DATABASE_DSN is unset/unreachable. Refuses to
// run against any DSN that doesn't look like an E2E database, mirroring
// backend/cmd/seed-e2e's own guard.
func TestE2E_PortfolioComplianceV2_ManagerApprovalToExecution(t *testing.T) {
	if testing.Short() {
		t.Skip("e2e test")
	}

	dsn := os.Getenv("IMS_TEST_DATABASE_DSN")
	if dsn == "" {
		dsn = "host=localhost port=5437 user=ims_app password=ims_dev_password dbname=ims_e2e sslmode=disable"
	}
	if !strings.Contains(strings.ToLower(dsn), "e2e") {
		t.Fatalf("refusing to run against a non-E2E-looking database (dbname must contain \"e2e\"): %s", dsn)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	probe, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Postgres E2E database unavailable: %v", err)
	}
	probe.Close(ctx)

	if v := os.Getenv("IMS_ALLOW_FAKE_WORKFLOW_STATE"); v == "true" {
		t.Fatalf("IMS_ALLOW_FAKE_WORKFLOW_STATE=true; E2E test refuses fake bindings")
	}

	server, pool := bootComplianceTestServer(t, ctx, dsn)
	defer server.Close()
	defer pool.Close()

	// ── Actors: admin submits; ben (stage 1, FUND_MANAGER_REVIEWERS) and
	// green (stage 2/final, INVESTMENT_SUPERVISORS) approve — the real
	// two-stage process seeded by 012_approval_demo_seed.sql. The approval
	// engine refuses self-approval (runtime_service.go authorizeAction), so
	// admin (the submitter) must not be an approver on this request.
	_, err = pool.Exec(ctx, `
		UPDATE iam_users SET force_password_change = false
		WHERE username IN ('admin', 'ben', 'green')`)
	require.NoError(t, err, "reset force_password_change for admin/ben/green")

	grantAdminFunctionPermissions(t, ctx, pool,
		"INVESTMENT_DECISION_VIEW", "INVESTMENT_DECISION_MANAGE", "INVESTMENT_DECISION_SUBMIT",
		"INVESTMENT_EXECUTION_MANAGE", "INVESTMENT_PORTFOLIO_VIEW",
		"IRG_VIEW_RULES", "IRG_EDIT_RULE_INSTANCE", "IRG_EDIT_BINDING", "WORKFLOW_EXECUTE",
	)
	grantApproverPermissions(t, ctx, pool)

	fundID, portfolioID, instrumentID := seedInvestmentPrereqs(t, ctx, pool)
	defer cleanupInvestmentPrereqs(ctx, pool, fundID)
	seedPortfolioCash(t, ctx, pool, portfolioID, "THB", "100000")
	grantUserDataRight(t, ctx, pool, benUserID, fundID.String())
	grantUserDataRight(t, ctx, pool, greenUserID, fundID.String())

	var portfolioCode, instrumentTicker string
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT code FROM investment__portfolios WHERE id = $1`, portfolioID,
	).Scan(&portfolioCode), "fetch seeded portfolio code")
	require.NoError(t, pool.QueryRow(ctx,
		`SELECT primary_ticker FROM investment__instruments WHERE id = $1`, instrumentID,
	).Scan(&instrumentTicker), "fetch seeded instrument ticker")

	adminToken := loginAdmin(t, server.URL)
	benToken := loginE2EPassword(t, server.URL, "ben", "admin123")
	greenToken := loginE2EPassword(t, server.URL, "green", "admin123")

	// ── 1. Create the rule + parameters, then bind it BLOCK to the portfolio.
	ruleInstanceID := createRuleInstance(t, server.URL, adminToken, createRuleInstanceReq{
		RuleTypeID:    "exposure.max_order_percent_aum",
		Name:          fmt.Sprintf("E2E max order pct AUM %s", uuid.NewString()[:8]),
		Description:   "P1 E2E: caps a single order at 5% of AUM",
		Parameters:    json.RawMessage(`{"max_percent_aum": 5}`),
		EffectiveFrom: time.Now().UTC().Format("2006-01-02"),
	})

	blockBindingID := bindPortfolioRule(t, server.URL, adminToken, portfolioCode, ruleInstanceID, bindRuleReq{
		Severity:      "BLOCK",
		EffectiveFrom: time.Now().UTC().Format("2006-01-02"),
	})

	// Prove the V2 catalog route surfaces the binding we just created —
	// requirement "verify portfolio-code compliance routes".
	rulesEnvelope, rulesStatus := getV2(t, server.URL, adminToken, portfolioCode, "compliance/rules")
	require.Equal(t, http.StatusOK, rulesStatus, "list portfolio rules expected 200; body=%+v", rulesEnvelope)
	require.True(t, catalogHasActiveBinding(t, rulesEnvelope, ruleInstanceID, blockBindingID),
		"newly created binding must appear in the portfolio rule catalog")

	// ── 2. Create a DRAFT decision small enough to pass every rule
	// (including the four GLOBAL rules: cash.availability, min_trading_unit,
	// min_trade_amount, sector_exposure) at submit time: 20 * 100 = 2,000 THB
	// = 2% of the 100,000 THB AUM, comfortably under the 5% limit.
	businessDate := time.Now().UTC().Format("2006-01-02")
	decisionBody := map[string]any{
		"instrument_code": instrumentTicker,
		"business_date":   businessDate,
		"side":            "BUY",
		"quantity":        "20",
		"limit_price":     "100",
		"currency":        "THB",
		"rationale":       "P1 E2E: manager-approval-to-execution compliance proof",
	}
	createEnvelope, createStatus := postV2(t, server.URL, adminToken, portfolioCode, "decisions", decisionBody)
	require.Equal(t, http.StatusCreated, createStatus, "create decision expected 201; body=%+v", createEnvelope)
	decisionData, ok := createEnvelope["data"].(map[string]any)
	require.True(t, ok, "create decision response missing data envelope: %+v", createEnvelope)
	decisionID, _ := decisionData["id"].(string)
	require.NotEmpty(t, decisionID, "created decision must have an id")

	// ── 3. Submit -> pre-trade compliance PASSes (2% < 5%) -> real approval
	// request created against PROC_DECISION_DEFAULT.
	submitEnvelope, submitStatus := postV2(t, server.URL, adminToken, portfolioCode, "decisions/"+decisionID+"/submit", map[string]any{})
	require.Equal(t, http.StatusOK, submitStatus, "submit decision expected 200; body=%+v", submitEnvelope)
	submitData, ok := submitEnvelope["data"].(map[string]any)
	require.True(t, ok, "submit response missing data envelope: %+v", submitEnvelope)
	require.Equal(t, "PENDING_APPROVAL", submitData["status"], "decision must be PENDING_APPROVAL after submit")
	approvalRequestID, _ := submitData["approval_request_id"].(string)
	require.NotEmpty(t, approvalRequestID, "submit must create a real approval request")

	// ── 4. Stage 1: ben (FUND_MANAGER_REVIEWERS) approves via the real
	// inbox -> approve HTTP flow.
	benTaskID := findInboxTaskForSubject(t, server.URL, benToken, decisionID)
	require.NotEmpty(t, benTaskID, "ben must have an inbox task for the submitted decision (stage 1)")
	approveEnvelope, approveStatus := postV1(t, server.URL, benToken, "approvals/tasks/"+benTaskID+"/approve", map[string]any{"comment": "reviewed"})
	require.Equal(t, http.StatusOK, approveStatus, "ben's stage-1 approval expected 200; body=%+v", approveEnvelope)

	// ── 5. Stage 2/final: green (INVESTMENT_SUPERVISORS) approves -> the
	// approval engine's final-decision callback
	// (DecisionApprovalSubjectCallback) fires synchronously and moves the
	// decision to APPROVED.
	greenTaskID := findInboxTaskForSubject(t, server.URL, greenToken, decisionID)
	require.NotEmpty(t, greenTaskID, "green must have an inbox task for the submitted decision (stage 2)")
	finalEnvelope, finalStatus := postV1(t, server.URL, greenToken, "approvals/tasks/"+greenTaskID+"/approve", map[string]any{"comment": "final sign-off"})
	require.Equal(t, http.StatusOK, finalStatus, "green's final approval expected 200; body=%+v", finalEnvelope)

	getEnvelope, getStatus := getV2(t, server.URL, adminToken, portfolioCode, "decisions/"+decisionID)
	require.Equal(t, http.StatusOK, getStatus)
	getData, ok := getEnvelope["data"].(map[string]any)
	require.True(t, ok)
	require.Equal(t, "APPROVED", getData["status"], "decision must be APPROVED after both real approval stages complete")

	// ── 6. BLOCK: attempt execution at 600 units (600*100 = 60,000 THB =
	// 60% of AUM, way over the 5% limit). Still well within cash (100,000
	// available) and every GLOBAL rule, so the portfolio-scoped BLOCK
	// binding is the only thing that can reject this.
	t.Run("BLOCK prevents execution and persists an auditable check record", func(t *testing.T) {
		var executionsBefore int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM investment__executions WHERE decision_id = $1`, decisionID,
		).Scan(&executionsBefore))

		execEnvelope, execStatus := postV2(t, server.URL, adminToken, portfolioCode,
			"decisions/"+decisionID+"/executions", map[string]any{"ordered_quantity": "600"})
		require.Equal(t, http.StatusUnprocessableEntity, execStatus,
			"BLOCK must return a typed 422 rejection (not 500); body=%+v", execEnvelope)
		require.NotEmpty(t, execEnvelope["error"], "422 body must carry a human-readable rejection message")

		var executionsAfter int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM investment__executions WHERE decision_id = $1`, decisionID,
		).Scan(&executionsAfter))
		require.Equal(t, executionsBefore, executionsAfter, "a BLOCK verdict must create no execution row")

		rec := latestCheckRecord(t, ctx, pool, portfolioID, "exposure.max_order_percent_aum")
		require.Equal(t, "BLOCK", rec.Verdict, "raw rule verdict must be BLOCK at 60%% of AUM")
		require.Equal(t, "BLOCK", rec.EffectiveSeverity)
		require.Equal(t, "BLOCK", rec.FinalVerdict)

		var breachCount int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM compliance_breaches WHERE check_record_id = $1`, rec.ID,
		).Scan(&breachCount))
		require.Equal(t, 1, breachCount, "a BLOCK final verdict must persist an auditable breach row")

		decEnvelope, decStatus := getV2(t, server.URL, adminToken, portfolioCode, "decisions/"+decisionID)
		require.Equal(t, http.StatusOK, decStatus)
		decData := decEnvelope["data"].(map[string]any)
		require.Equal(t, "APPROVED", decData["status"], "decision must remain APPROVED (retryable) after a blocked execution attempt")
	})

	// ── 7. PASS: attempt execution at 20 units — the same size as the
	// approved decision itself (2,000 THB = 2% of AUM). The BLOCK binding is
	// still active, but this order does not breach it (nor any GLOBAL rule,
	// including concentration.single_issuer — confirmed PASS at this size by
	// the submit-time check above), so the aggregate verdict is a genuine
	// PASS: execution succeeds AND an auditable PASS check record exists.
	t.Run("PASS permits execution and produces an auditable check record", func(t *testing.T) {
		execEnvelope, execStatus := postV2(t, server.URL, adminToken, portfolioCode,
			"decisions/"+decisionID+"/executions", map[string]any{"ordered_quantity": "20"})
		require.Equal(t, http.StatusCreated, execStatus, "PASS expected 201; body=%+v", execEnvelope)
		execData, ok := execEnvelope["data"].(map[string]any)
		require.True(t, ok, "execution response missing data envelope: %+v", execEnvelope)
		passExecutionID, _ := execData["id"].(string)
		require.NotEmpty(t, passExecutionID, "created execution must have an id")

		var persisted int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM investment__executions WHERE id = $1 AND decision_id = $2`, passExecutionID, decisionID,
		).Scan(&persisted))
		require.Equal(t, 1, persisted, "PASS verdict must persist the execution row")

		rec := latestCheckRecord(t, ctx, pool, portfolioID, "exposure.max_order_percent_aum")
		require.Equal(t, "PASS", rec.Verdict, "raw rule verdict must be PASS at 2%% of AUM")
		require.Equal(t, "BLOCK", rec.EffectiveSeverity, "the BLOCK binding is still active for this check")
		require.Equal(t, "PASS", rec.FinalVerdict)

		var breachCount int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM compliance_breaches WHERE check_record_id = $1`, rec.ID,
		).Scan(&breachCount))
		require.Equal(t, 0, breachCount, "a PASS final verdict must not create a breach row")
	})

	// ── 8. Deactivate the BLOCK binding -> the oversized (600-unit) order
	// that BLOCKed above is no longer gated by exposure.max_order_percent_aum
	// specifically (only the GLOBAL rules remain). Proves deactivation
	// changes which rule applies.
	t.Run("deactivating the binding permits the previously-blocked order", func(t *testing.T) {
		deactivateStatus := deleteV2(t, server.URL, adminToken, portfolioCode,
			fmt.Sprintf("compliance/rules/%s/bindings/%s", ruleInstanceID, blockBindingID))
		require.Equal(t, http.StatusNoContent, deactivateStatus, "deactivate binding expected 204")

		since := time.Now().UTC()
		execEnvelope, execStatus := postV2(t, server.URL, adminToken, portfolioCode,
			"decisions/"+decisionID+"/executions", map[string]any{"ordered_quantity": "600"})
		require.Equal(t, http.StatusCreated, execStatus, "execution after deactivation expected 201; body=%+v", execEnvelope)
		execData, ok := execEnvelope["data"].(map[string]any)
		require.True(t, ok, "execution response missing data envelope: %+v", execEnvelope)
		deactivatedExecutionID, _ := execData["id"].(string)
		require.NotEmpty(t, deactivatedExecutionID, "created execution must have an id")

		var persisted int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM investment__executions WHERE id = $1 AND decision_id = $2`, deactivatedExecutionID, decisionID,
		).Scan(&persisted))
		require.Equal(t, 1, persisted, "the deactivated-binding execution must persist")

		require.False(t, ruleEvaluatedSince(t, ctx, pool, portfolioID, "exposure.max_order_percent_aum", since),
			"with the binding deactivated, exposure.max_order_percent_aum must not be resolved as applicable at all")
	})

	// ── 8. Rebind the SAME rule instance at severity=WARN and re-attempt the
	// same oversized order: raw verdict is still BLOCK (order is still 60%
	// of AUM > 5%), but Severity.CapVerdict downgrades it to WARN. WARN must
	// not prevent execution — the second execution succeeds.
	var warnBindingID string
	t.Run("WARN severity caps the raw BLOCK verdict and does not prevent execution", func(t *testing.T) {
		warnBindingID = bindPortfolioRule(t, server.URL, adminToken, portfolioCode, ruleInstanceID, bindRuleReq{
			Severity:      "WARN",
			EffectiveFrom: time.Now().UTC().Format("2006-01-02"),
		})

		execEnvelope, execStatus := postV2(t, server.URL, adminToken, portfolioCode,
			"decisions/"+decisionID+"/executions", map[string]any{"ordered_quantity": "600"})
		require.Equal(t, http.StatusCreated, execStatus, "WARN must not block execution; body=%+v", execEnvelope)
		execData, ok := execEnvelope["data"].(map[string]any)
		require.True(t, ok)
		warnExecutionID, _ := execData["id"].(string)
		require.NotEmpty(t, warnExecutionID, "created execution must have an id")

		var persisted int
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT COUNT(*) FROM investment__executions WHERE id = $1 AND decision_id = $2`, warnExecutionID, decisionID,
		).Scan(&persisted))
		require.Equal(t, 1, persisted, "the WARN execution must persist")

		rec := latestCheckRecord(t, ctx, pool, portfolioID, "exposure.max_order_percent_aum")
		require.Equal(t, "BLOCK", rec.Verdict, "the rule's raw verdict is still BLOCK at 60%% of AUM")
		require.Equal(t, "WARN", rec.EffectiveSeverity, "binding severity is WARN")
		require.Equal(t, "WARN", rec.FinalVerdict, "CapVerdict must downgrade BLOCK to WARN, not raise it")

		var breach struct {
			Severity string
			Verdict  string
			Status   string
		}
		require.NoError(t, pool.QueryRow(ctx,
			`SELECT severity, verdict, status FROM compliance_breaches WHERE check_record_id = $1`, rec.ID,
		).Scan(&breach.Severity, &breach.Verdict, &breach.Status),
			"a WARN final verdict must still persist an auditable breach row")
		require.Equal(t, "WARN", breach.Severity)
		require.Equal(t, "WARN", breach.Verdict)
		require.Equal(t, "OPEN", breach.Status)
	})

	// ── 9. Deactivate the WARN binding and rebind the same rule instance
	// BLOCK again, but with effective_from = tomorrow. The order that would
	// BLOCK today does not, because the binding is not yet effective —
	// proving effective-dated scheduling distinct from deactivation.
	t.Run("a future effective_from binding does not apply today", func(t *testing.T) {
		deactivateStatus := deleteV2(t, server.URL, adminToken, portfolioCode,
			fmt.Sprintf("compliance/rules/%s/bindings/%s", ruleInstanceID, warnBindingID))
		require.Equal(t, http.StatusNoContent, deactivateStatus)

		tomorrow := time.Now().UTC().AddDate(0, 0, 1).Format("2006-01-02")
		_ = bindPortfolioRule(t, server.URL, adminToken, portfolioCode, ruleInstanceID, bindRuleReq{
			Severity:      "BLOCK",
			EffectiveFrom: tomorrow,
		})

		since := time.Now().UTC()
		execEnvelope, execStatus := postV2(t, server.URL, adminToken, portfolioCode,
			"decisions/"+decisionID+"/executions", map[string]any{"ordered_quantity": "600"})
		require.Equal(t, http.StatusCreated, execStatus,
			"a binding effective only from tomorrow must not gate today's order; body=%+v", execEnvelope)

		require.False(t, ruleEvaluatedSince(t, ctx, pool, portfolioID, "exposure.max_order_percent_aum", since),
			"a binding effective only from tomorrow must not be resolved as applicable today")
	})
}

// ─── Server bootstrap ──────────────────────────────────────────────────────

// bootComplianceTestServer mirrors cmd/server/main.go's wiring more fully
// than the shared bootTestServer in investment_test.go: it additionally
// wires the approval module (submitter/status/batch/canceller ports,
// subject callbacks, subject validators, subject access ports) and the
// Portfolio Compliance V2 rule-administration surface
// (SetPortfolioComplianceAdmin), both required for the manager-approval and
// rule-binding flows this file proves. Kept as a separate helper rather than
// extending bootTestServer to keep this P1 addition isolated from the three
// existing e2e test files that helper already backs.
func bootComplianceTestServer(t *testing.T, ctx context.Context, dsn string) (*httptest.Server, *pgxpool.Pool) {
	t.Helper()

	cfg, err := config.Load()
	require.NoError(t, err, "config.Load")

	cfg.RateLimitBackend = "memory"
	dsnHost, dsnPort, dsnDB, dsnUser, dsnPass := parseDSN(t, dsn)
	cfg.DBHost = dsnHost
	cfg.DBPort = dsnPort
	cfg.DBName = dsnDB
	cfg.DBUser = dsnUser
	cfg.DBPassword = dsnPass

	pool, err := database.NewPool(ctx, cfg)
	require.NoError(t, err, "database.NewPool")

	auditMod := audit.NewModule(pool)
	iamMod, err := iam.NewModule(pool, cfg, nil /* redis */, auditMod.Recorder(), auditMod)
	require.NoError(t, err, "iam.NewModule")

	complianceMod := compliance.NewModule(pool, iamMod)
	workflowMod := workflow.NewModule(pool, iamMod, complianceMod.ContractAdapter())
	investmentMod := investment.NewModule(
		pool,
		complianceMod.ContractAdapter(),
		workflowMod,
		iamMod,
		auditMod.Recorder(),
	)
	referenceDataMod := referencedata.NewModule(pool)
	marketDataMod := marketdata.NewModule(pool, cfg, nil /* redis */, referenceDataMod.Resolver())
	approvalMod := approval.NewModule(pool, iamMod, auditMod.Recorder(), nil /* notifier */, approvaladapter.NewPostgresDelegateResolver(pool), nil /* leave */)

	investmentMod.SetPortfolioComplianceAdmin(complianceMod.PortfolioContractAdapter())
	investmentMod.SetApprovalSubmitter(approvalMod)
	investmentMod.SetApprovalStatusProvider(approvalMod)
	investmentMod.SetApprovalBatchActor(approvalMod)
	investmentMod.SetApprovalCanceller(approvalMod)
	approvalMod.RegisterSubjectCallback("RESEARCH_REPORT", investmentMod.ApprovalSubjectCallback())
	approvalMod.RegisterSubjectCallback("INVESTMENT_DECISION", investmentMod.DecisionApprovalSubjectCallback())
	approvalMod.RegisterSubjectCallback("COMPLIANCE_RELEASE", investmentMod.ComplianceReleaseSubjectCallback())
	approvalMod.RegisterSubjectValidator("RESEARCH_REPORT", investmentMod.ResearchReportSubjectValidator())
	approvalMod.RegisterSubjectValidator("INVESTMENT_DECISION", investmentMod.DecisionSubjectValidator())
	approvalMod.RegisterSubjectValidator("COMPLIANCE_RELEASE", investmentMod.ComplianceReleaseSubjectValidator())
	if accessor := investmentMod.SubjectAccessor(iamMod); accessor != nil {
		approvalMod.RegisterSubjectAccessPort("RESEARCH_REPORT", accessor)
		approvalMod.RegisterSubjectAccessPort("INVESTMENT_DECISION", accessor)
		approvalMod.RegisterSubjectAccessPort("COMPLIANCE_RELEASE", accessor)
	}

	r := chi.NewRouter()
	r.Use(chimw.RequestID)
	r.Use(middleware.NewCORS(cfg.CORSAllowedOrigins))
	r.Use(middleware.SecureHeaders)
	r.Use(iamMod.GlobalRateLimitMiddleware())

	r.Route("/api/v1", func(r chi.Router) {
		iamMod.SetupRoutes(r)
		r.Group(func(r chi.Router) {
			r.Use(iamMod.AuthMiddleware())
			complianceMod.RegisterRoutes(r)
			workflowMod.RegisterRoutes(r)
			investmentMod.RegisterRoutes(r)
			marketDataMod.RegisterRoutes(r)
			approvalMod.RegisterRoutes(r)
		})
	})

	r.Route("/api/v2", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Use(iamMod.AuthMiddleware())
			investmentMod.RegisterRoutesV2(r)
		})
	})

	return httptest.NewServer(r), pool
}

// ─── Fixture-user constants and permission grants ─────────────────────────

const (
	benUserID        = "a0000000-0000-0000-0000-000000000010" // ben, database/seeds/004_investment_process_assignment_seed.sql
	greenUserID      = "a0000000-0000-0000-0000-000000000011" // green, same file
	approversGroupID = "b9500000-0000-0000-0000-000000000001"
)

// grantApproverPermissions grants ben and green (the seeded
// FUND_MANAGER_REVIEWERS / INVESTMENT_SUPERVISORS members from
// database/seeds/012_approval_demo_seed.sql) the approval-runtime function
// permissions they need to view their inbox and approve via HTTP. A
// dedicated group is used (rather than adding them to Admin) so this test
// grants exactly what a real reviewer/supervisor needs, nothing more.
func grantApproverPermissions(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()
	const adminUserID = "a0000000-0000-0000-0000-000000000001"

	_, err := pool.Exec(ctx, `
		INSERT INTO permissions_groups (id, name, description, is_active)
		VALUES ($1, 'E2E Compliance Approvers', 'P1 E2E: approval-runtime permissions for ben/green', true)
		ON CONFLICT (id) DO NOTHING`, approversGroupID)
	require.NoError(t, err, "create E2E approvers group")

	for _, userID := range []string{benUserID, greenUserID} {
		_, err := pool.Exec(ctx, `
			INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
			VALUES ($1, $2, $3)
			ON CONFLICT (user_id, group_id) DO NOTHING`,
			userID, approversGroupID, adminUserID)
		require.NoError(t, err, "add %s to E2E approvers group", userID)
	}

	for _, code := range []string{"APPROVAL_VIEW_INBOX", "APPROVAL_VIEW_REQUEST", "APPROVAL_APPROVE", "APPROVAL_REJECT"} {
		_, err := pool.Exec(ctx, `
			INSERT INTO permissions_function_rights (group_id, permission_code, is_granted, created_by)
			VALUES ($1, $2, true, $3)
			ON CONFLICT (group_id, permission_code) DO UPDATE SET is_granted = true`,
			approversGroupID, code, adminUserID)
		require.NoError(t, err, "grant %s to E2E approvers group", code)
	}
}

// grantUserDataRight grants userID data-permission access to contractID
// (here, the test fund) — required by InvestmentSubjectAccessor for both
// submitting and acting on an INVESTMENT_DECISION approval subject.
func grantUserDataRight(t *testing.T, ctx context.Context, pool *pgxpool.Pool, userID, contractID string) {
	t.Helper()
	const adminUserID = "a0000000-0000-0000-0000-000000000001"
	_, err := pool.Exec(ctx, `
		INSERT INTO permissions_data_rights (user_id, contract_id, is_granted, granted_by)
		VALUES ($1, $2::TEXT, true, $3)
		ON CONFLICT (user_id, contract_id) DO UPDATE SET is_granted = true`,
		userID, contractID, adminUserID)
	require.NoError(t, err, "grant data right for %s on %s", userID, contractID)
}

// seedPortfolioCash inserts a starting cash balance so cash.availability
// (the GLOBAL rule bound by database/seeds/009_compliance_rule_seed.sql)
// passes for every order size this test uses, and so
// exposure.max_order_percent_aum has a non-zero AUM to compute against
// (investment_data_adapters.go's getNAV falls back to cash + position market
// value when no valuation snapshot exists).
func seedPortfolioCash(t *testing.T, ctx context.Context, pool *pgxpool.Pool, portfolioID uuid.UUID, currency, balance string) {
	t.Helper()
	_, err := pool.Exec(ctx, `
		INSERT INTO investment__cash_balances (id, portfolio_id, currency, balance)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (portfolio_id, currency) DO UPDATE SET balance = EXCLUDED.balance`,
		uuid.New(), portfolioID, currency, balance)
	require.NoError(t, err, "seed portfolio cash balance")
}

// ─── HTTP helpers ──────────────────────────────────────────────────────────

// loginE2EPassword logs a seeded dev user (ben/green — see
// database/seeds/004_investment_process_assignment_seed.sql) in with the
// shared dev bcrypt hash's password and fails the test on anything but 200.
func loginE2EPassword(t *testing.T, baseURL, username, password string) string {
	t.Helper()
	body := map[string]string{"username": username, "password": password}
	buf, err := json.Marshal(body)
	require.NoError(t, err)
	resp, err := http.Post(baseURL+"/api/v1/auth/login", "application/json", bytes.NewReader(buf))
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	require.Equal(t, http.StatusOK, resp.StatusCode, "login for %s expected 200, got %d body=%s", username, resp.StatusCode, string(raw))
	var envelope struct {
		Data struct {
			AccessToken string `json:"access_token"`
		} `json:"data"`
	}
	require.NoError(t, json.Unmarshal(raw, &envelope))
	require.NotEmpty(t, envelope.Data.AccessToken, "no access_token for %s", username)
	return envelope.Data.AccessToken
}

// postV1 / getV1 / deleteV2 are thin JSON helpers for routes not already
// covered by postV2/getV2 (portfolio_v2_ledger_test.go) — /api/v1/* runtime
// routes (approvals, compliance rule administration) and V2 DELETE.
func postV1(t *testing.T, baseURL, token, subpath string, body map[string]any) (map[string]any, int) {
	t.Helper()
	return doJSON(t, http.MethodPost, fmt.Sprintf("%s/api/v1/%s", baseURL, subpath), token, body)
}

func getV1(t *testing.T, baseURL, token, subpath string) (map[string]any, int) {
	t.Helper()
	return doJSON(t, http.MethodGet, fmt.Sprintf("%s/api/v1/%s", baseURL, subpath), token, nil)
}

func deleteV2(t *testing.T, baseURL, token, portfolioCode, subpath string) int {
	t.Helper()
	url := fmt.Sprintf("%s/api/v2/portfolios/%s/%s", baseURL, portfolioCode, subpath)
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	require.NoError(t, err)
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	return resp.StatusCode
}

func doJSON(t *testing.T, method, url, token string, body map[string]any) (map[string]any, int) {
	t.Helper()
	var reader io.Reader
	if body != nil {
		buf, err := json.Marshal(body)
		require.NoError(t, err)
		reader = bytes.NewReader(buf)
	}
	req, err := http.NewRequest(method, url, reader)
	require.NoError(t, err)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	req.Header.Set("Authorization", "Bearer "+token)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	var out map[string]any
	if len(raw) > 0 {
		_ = json.Unmarshal(raw, &out)
	}
	return out, resp.StatusCode
}

// findInboxTaskForSubject calls the real GET /approvals/inbox as the given
// actor's own token and returns the pending task id whose request subject_id
// matches decisionID, or "" if none is found.
func findInboxTaskForSubject(t *testing.T, baseURL, token, decisionID string) string {
	t.Helper()
	envelope, status := getV1(t, baseURL, token, "approvals/inbox?limit=200")
	require.Equal(t, http.StatusOK, status, "get inbox expected 200; body=%+v", envelope)
	data, ok := envelope["data"].(map[string]any)
	require.True(t, ok, "inbox response missing data envelope: %+v", envelope)
	items, _ := data["items"].([]any)
	for _, raw := range items {
		item, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		reqObj, ok := item["request"].(map[string]any)
		if !ok {
			continue
		}
		if subjectID, _ := reqObj["subject_id"].(string); subjectID != decisionID {
			continue
		}
		taskObj, ok := item["task"].(map[string]any)
		if !ok {
			continue
		}
		taskID, _ := taskObj["id"].(string)
		return taskID
	}
	return ""
}

// ─── Compliance rule / binding helpers ─────────────────────────────────────

type createRuleInstanceReq struct {
	RuleTypeID    string
	Name          string
	Description   string
	Parameters    json.RawMessage
	EffectiveFrom string
}

// createRuleInstance calls POST /api/v1/compliance/rules and returns the new
// rule instance's UUID.
func createRuleInstance(t *testing.T, baseURL, token string, req createRuleInstanceReq) string {
	t.Helper()
	body := map[string]any{
		"rule_type_id":   req.RuleTypeID,
		"name":           req.Name,
		"description":    req.Description,
		"parameters":     req.Parameters,
		"effective_from": req.EffectiveFrom,
	}
	envelope, status := postV1(t, baseURL, token, "compliance/rules", body)
	require.Equal(t, http.StatusCreated, status, "create rule instance expected 201; body=%+v", envelope)
	data, ok := envelope["data"].(map[string]any)
	require.True(t, ok, "create rule instance response missing data envelope: %+v", envelope)
	instance, ok := data["instance"].(map[string]any)
	require.True(t, ok, "create rule instance response missing instance: %+v", envelope)
	id, _ := instance["ID"].(string)
	require.NotEmpty(t, id, "created rule instance must have an ID")
	return id
}

type bindRuleReq struct {
	Severity      string
	EffectiveFrom string
}

// bindPortfolioRule calls the V2 bind-rule route and returns the new
// binding's UUID.
func bindPortfolioRule(t *testing.T, baseURL, token, portfolioCode, ruleInstanceID string, req bindRuleReq) string {
	t.Helper()
	envelope, status := postV2(t, baseURL, token, portfolioCode,
		fmt.Sprintf("compliance/rules/%s/bindings", ruleInstanceID),
		map[string]any{"severity": req.Severity, "effective_from": req.EffectiveFrom})
	require.Equal(t, http.StatusCreated, status, "bind rule expected 201; body=%+v", envelope)
	data, ok := envelope["data"].(map[string]any)
	require.True(t, ok, "bind rule response missing data envelope: %+v", envelope)
	bindingID, _ := data["binding_id"].(string)
	require.NotEmpty(t, bindingID, "binding response must have a binding_id")
	return bindingID
}

// catalogHasActiveBinding inspects the GET .../compliance/rules response
// (a JSON array under "data") for an entry matching ruleInstanceID whose
// binding.binding_id matches bindingID and is active.
func catalogHasActiveBinding(t *testing.T, envelope map[string]any, ruleInstanceID, bindingID string) bool {
	t.Helper()
	items, ok := envelope["data"].([]any)
	require.True(t, ok, "rule catalog response missing data array: %+v", envelope)
	for _, raw := range items {
		entry, ok := raw.(map[string]any)
		if !ok {
			continue
		}
		if entry["rule_instance_id"] != ruleInstanceID {
			continue
		}
		binding, ok := entry["binding"].(map[string]any)
		if !ok {
			continue
		}
		if binding["binding_id"] == bindingID && binding["is_active"] == true {
			return true
		}
	}
	return false
}

// ─── Audit trail assertions ─────────────────────────────────────────────────

type checkRecordRow struct {
	ID                string
	Verdict           string
	EffectiveSeverity string
	FinalVerdict      string
}

// latestCheckRecord returns the most recently persisted
// compliance_check_records row for (portfolioID, ruleTypeID) — the
// auditable evidence of the most recent evaluation of that rule.
func latestCheckRecord(t *testing.T, ctx context.Context, pool *pgxpool.Pool, portfolioID uuid.UUID, ruleTypeID string) checkRecordRow {
	t.Helper()
	var rec checkRecordRow
	err := pool.QueryRow(ctx, `
		SELECT id, verdict, effective_severity, final_verdict
		FROM compliance_check_records
		WHERE portfolio_id = $1 AND rule_type_id = $2
		ORDER BY checked_at DESC, created_at DESC
		LIMIT 1`, portfolioID, ruleTypeID,
	).Scan(&rec.ID, &rec.Verdict, &rec.EffectiveSeverity, &rec.FinalVerdict)
	require.NoError(t, err, "no compliance_check_records row found for portfolio=%s rule_type=%s", portfolioID, ruleTypeID)
	return rec
}

// ruleEvaluatedSince reports whether any compliance_check_records row for
// (portfolioID, ruleTypeID) was created at or after since. Used to prove a
// rule was NOT resolved as applicable (deactivated binding, or a binding not
// yet effective) — the pipeline simply never evaluates it, rather than
// evaluating it and passing.
func ruleEvaluatedSince(t *testing.T, ctx context.Context, pool *pgxpool.Pool, portfolioID uuid.UUID, ruleTypeID string, since time.Time) bool {
	t.Helper()
	var count int
	err := pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM compliance_check_records
		WHERE portfolio_id = $1 AND rule_type_id = $2 AND checked_at >= $3`,
		portfolioID, ruleTypeID, since,
	).Scan(&count)
	require.NoError(t, err, "counting compliance_check_records for portfolio=%s rule_type=%s", portfolioID, ruleTypeID)
	return count > 0
}
