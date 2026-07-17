package command

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

const (
	ComplianceErrorCodeNotConfigured = "COMPLIANCE_NOT_CONFIGURED"
	ComplianceErrorCodeUnavailable   = "COMPLIANCE_UNAVAILABLE"

	complianceControlGapAuditAction = "INVESTMENT_COMPLIANCE_CONTROL_GAP_DETECTED"
)

// ErrComplianceNotConfigured is returned when a LIVE portfolio has no active,
// effective rule binding on the decision business date.
type ErrComplianceNotConfigured struct {
	CheckGroupID string
}

func (*ErrComplianceNotConfigured) Error() string {
	return "compliance controls are not configured for this LIVE portfolio"
}

// ErrComplianceUnavailable is returned when a configured control could not
// produce a reliable result, for example because asset classification is
// missing for a holding or proposed instrument.
type ErrComplianceUnavailable struct {
	CheckGroupID string
	Reason       string
}

func (e *ErrComplianceUnavailable) Error() string {
	if e == nil || e.Reason == "" {
		return "compliance controls are unavailable for this LIVE portfolio"
	}
	return fmt.Sprintf("compliance controls are unavailable for this LIVE portfolio: %s", e.Reason)
}

func liveComplianceStatusError(
	portfolioType vo.PortfolioType,
	result *contract.ProposedOrderResult,
) error {
	if portfolioType != vo.PortfolioTypeLive {
		return nil
	}
	if result == nil {
		return &ErrComplianceUnavailable{Reason: "compliance check returned nil result"}
	}

	checkGroupID := result.CheckGroupID.String()
	switch result.Status {
	case contract.ComplianceStatusNotConfigured:
		return &ErrComplianceNotConfigured{CheckGroupID: checkGroupID}
	case contract.ComplianceStatusUnavailable:
		return &ErrComplianceUnavailable{
			CheckGroupID: checkGroupID,
			Reason:       summarizeBreaches(result.Breaches),
		}
	case contract.ComplianceStatusEvaluated:
		if result.RulesEvaluated == 0 {
			return &ErrComplianceNotConfigured{CheckGroupID: checkGroupID}
		}
		switch result.Verdict {
		case contract.ComplianceVerdictPass, contract.ComplianceVerdictWarn, contract.ComplianceVerdictBlock:
			return nil
		default:
			return &ErrComplianceUnavailable{
				CheckGroupID: checkGroupID,
				Reason:       fmt.Sprintf("unsupported compliance verdict %q", result.Verdict),
			}
		}
	case "":
		return &ErrComplianceUnavailable{
			CheckGroupID: checkGroupID,
			Reason:       "compliance result did not include a typed status",
		}
	default:
		return &ErrComplianceUnavailable{
			CheckGroupID: checkGroupID,
			Reason:       fmt.Sprintf("unsupported compliance status %q", result.Status),
		}
	}
}

func resolveCompliancePortfolioType(
	ctx context.Context,
	portfolios domain.PortfolioRepository,
	portfolioID uuid.UUID,
) (vo.PortfolioType, error) {
	if portfolios == nil {
		return "", fmt.Errorf("compliance portfolio repository is not initialised")
	}
	portfolio, err := portfolios.GetByID(ctx, portfolioID)
	if err != nil {
		return "", fmt.Errorf("loading portfolio for compliance policy: %w", err)
	}
	if portfolio == nil {
		return "", fmt.Errorf("loading portfolio for compliance policy: portfolio not found")
	}
	switch portfolio.PortfolioType {
	case vo.PortfolioTypeLive, vo.PortfolioTypeSimulation, vo.PortfolioTypeModel:
		return portfolio.PortfolioType, nil
	default:
		return "", fmt.Errorf("loading portfolio for compliance policy: unsupported portfolio type %q", portfolio.PortfolioType)
	}
}

type complianceControlGapAuditInput struct {
	Phase          string
	DecisionID     uuid.UUID
	PortfolioID    uuid.UUID
	PortfolioType  vo.PortfolioType
	ActorID        uuid.UUID
	BusinessDate   time.Time
	InstrumentCode string
}

func auditComplianceControlGap(
	ctx context.Context,
	audit contract.AuditLogger,
	input complianceControlGapAuditInput,
	gateErr error,
) error {
	status, checkGroupID, reason, ok := complianceControlGapDetails(gateErr)
	if !ok {
		return nil
	}
	if audit == nil {
		return fmt.Errorf("recording compliance control gap audit: audit logger is not initialised")
	}
	if err := audit.LogActionStrict(ctx, contract.AuditEntry{
		ActorID:      input.ActorID.String(),
		Action:       complianceControlGapAuditAction,
		Module:       "investment",
		ResourceType: "INVESTMENT_DECISION",
		ResourceID:   input.DecisionID.String(),
		BusinessDate: input.BusinessDate,
		Details: map[string]any{
			"check_group_id":    checkGroupID,
			"compliance_status": status,
			"reason":            reason,
			"phase":             input.Phase,
			"portfolio_id":      input.PortfolioID.String(),
			"portfolio_type":    string(input.PortfolioType),
			"instrument_code":   input.InstrumentCode,
		},
	}); err != nil {
		return fmt.Errorf("recording compliance control gap audit: %w", err)
	}
	return nil
}

func complianceControlGapDetails(err error) (status, checkGroupID, reason string, ok bool) {
	var notConfigured *ErrComplianceNotConfigured
	if errors.As(err, &notConfigured) {
		return ComplianceErrorCodeNotConfigured, notConfigured.CheckGroupID, "NO_ACTIVE_EFFECTIVE_BINDINGS", true
	}
	var unavailable *ErrComplianceUnavailable
	if errors.As(err, &unavailable) {
		reason = unavailable.Reason
		if reason == "" {
			reason = "COMPLIANCE_CONTROL_UNAVAILABLE"
		}
		return ComplianceErrorCodeUnavailable, unavailable.CheckGroupID, reason, true
	}
	return "", "", "", false
}
