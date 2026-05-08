// Command contract-check boots a minimal copy of the production module
// graph and asserts that every cross-module contract the investment module
// consumes is wired to a real (non-fake, non-nil) implementation.
//
// Exit codes:
//   - 0: every binding is real (or env-allowed fake).
//   - 1: at least one binding is nil or bound to a fake without an allow-list.
//   - 2: bootstrap failure (config / DB).
//
// Usage:
//
//	cd backend
//	go run cmd/contract-check/main.go
//
// CI hook: run after `make migrate-up && make seed` against the same Postgres
// the test suite uses. Failures must block the merge.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit"
	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance"
	"github.com/neo-kanta/ims-th-solution/backend/internal/iam"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/platform/config"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
	"github.com/neo-kanta/ims-th-solution/backend/platform/logging"
)

const (
	exitOK             = 0
	exitContractFailed = 1
	exitBootstrapFail  = 2
)

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		return exitBootstrapFail
	}

	logger := logging.NewLogger(cfg.LogLevel)
	slog.SetDefault(logger)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := database.NewPool(ctx, cfg)
	if err != nil {
		slog.Error("contract-check: database connection failed", "error", err)
		return exitBootstrapFail
	}
	defer pool.Close()

	auditModule := audit.NewModule(pool)

	iamModule, err := iam.NewModule(pool, cfg, nil, auditModule.Recorder(), auditModule)
	if err != nil {
		slog.Error("contract-check: IAM module construction failed", "error", err)
		return exitBootstrapFail
	}

	complianceModule := compliance.NewModule(pool, iamModule)

	bindings := buildBindings(complianceModule)

	outcomes, err := verifyBindings(bindings, os.LookupEnv)
	for _, o := range outcomes {
		switch o.Status {
		case "ok":
			slog.Info("contract-check: binding ok", "binding", o.Name)
		case "fake-allowed":
			slog.Warn("contract-check: binding bound to fake (env-allowed)", "binding", o.Name, "detail", o.Message)
		case "fake-rejected":
			slog.Error("contract-check: binding bound to fake", "binding", o.Name, "detail", o.Message)
		case "nil":
			slog.Error("contract-check: binding is nil", "binding", o.Name, "detail", o.Message)
		default:
			slog.Error("contract-check: unexpected status", "binding", o.Name, "status", o.Status)
		}
	}

	if err != nil {
		slog.Error("contract-check failed", "error", err)
		return exitContractFailed
	}

	slog.Info("contract-check passed", "bindings", len(outcomes))
	return exitOK
}

// complianceContracts is the surface of the compliance module the contract
// check inspects. The module's ContractAdapter() return type already
// satisfies both ComplianceChecker and PostTradeVerifier; this small
// interface lets the verify step accept any production-equivalent stand-in.
type complianceContracts interface {
	contract.ComplianceChecker
	contract.PostTradeVerifier
}

// buildBindings produces the list of contracts the contract-check inspects.
// Phase 0: the investment module's only consumed contract is ComplianceChecker.
// Phase 1+ adds investment-side bindings for WorkflowStateProvider and the
// audit recorder; appending entries here is the only required change.
func buildBindings(complianceModule *compliance.Module) []binding {
	var resolvedCompliance complianceContracts
	if complianceModule != nil {
		resolvedCompliance = complianceModule.ContractAdapter()
	}

	return []binding{
		{
			Name:     "investment.ComplianceChecker",
			Resolved: contract.ComplianceChecker(resolvedCompliance),
		},
		{
			Name:     "workflow.PostTradeVerifier",
			Resolved: contract.PostTradeVerifier(resolvedCompliance),
		},
	}
}
