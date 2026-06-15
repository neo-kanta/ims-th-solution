package command

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/domain/valueobject"
	workflowperm "github.com/neo-kanta/ims-th-solution/backend/internal/workflow/permission"
	"github.com/neo-kanta/ims-th-solution/backend/internal/workflow/ports"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
)

// ManagerApprovalRequest is the input for the APPROVE command.
//
// Zero-transaction days:
//   - If the investment module reports Count==0, ZeroTransactionAttestation must
//     be true and AttestationReason must be ≥ 30 characters.
//   - This is enforced by the policy layer, not this struct.
type ManagerApprovalRequest struct {
	ContractID   uuid.UUID
	BusinessDate time.Time
	Actor        vo.ActorContext

	// Zero-transaction attestation (required when no trades exist on the day)
	ZeroTransactionAttestation bool
	AttestationReason          string

	// Notes are an optional free-text remark stored on the approval record.
	Notes *string
}

// ManagerApprovalResult is returned on success.
type ManagerApprovalResult struct {
	TransitionID      uuid.UUID
	WorkflowDayID     uuid.UUID
	ApprovalID        uuid.UUID
	ContractID        uuid.UUID
	BusinessDate      time.Time
	FromState         vo.WorkflowState
	ToState           vo.WorkflowState
	OccurredAt        time.Time
	IsZeroTransaction bool
}

// ManagerApprovalHandler executes the APPROVE workflow transition.
//
// Transaction boundary: this handler owns the transaction.
// The transaction spans: workflow__day_states UPDATE + workflow__approval_records INSERT
// + workflow__transition_log INSERT — all three are committed atomically.
//
// The investment transaction summary is fetched BEFORE the transaction to avoid
// holding the database lock during an external call.
type ManagerApprovalHandler struct {
	pool            *pgxpool.Pool
	dayRepo         domain.WorkflowDayRepository
	logRepo         domain.TransitionLogRepository
	approvalRepo    domain.ApprovalRecordRepository
	investmentQuery ports.InvestmentQueryPort
	policy          *policy.TransitionPolicy
}

// NewManagerApprovalHandler creates the handler with its dependencies.
func NewManagerApprovalHandler(
	pool *pgxpool.Pool,
	dayRepo domain.WorkflowDayRepository,
	logRepo domain.TransitionLogRepository,
	approvalRepo domain.ApprovalRecordRepository,
	investmentQuery ports.InvestmentQueryPort,
	pol *policy.TransitionPolicy,
) *ManagerApprovalHandler {
	return &ManagerApprovalHandler{
		pool:            pool,
		dayRepo:         dayRepo,
		logRepo:         logRepo,
		approvalRepo:    approvalRepo,
		investmentQuery: investmentQuery,
		policy:          pol,
	}
}

// Handle executes the APPROVE transition.
//
// Flow:
//  1. Fetch transaction summary from investment module (external call, before tx).
//  2. BEGIN TRANSACTION.
//  3. SELECT … FOR UPDATE on the workflow day row.
//  4. Run policy checks (pure, no I/O).
//  5. UPDATE workflow__day_states (new state + lock timestamp + version bump).
//  6. INSERT workflow__approval_records.
//  7. INSERT workflow__transition_log.
//  8. COMMIT.
func (h *ManagerApprovalHandler) Handle(
	ctx context.Context,
	req ManagerApprovalRequest,
) (*ManagerApprovalResult, error) {
	// Step 1 — transaction summary (external port, before DB lock)
	txSummary, err := h.investmentQuery.GetTransactionSummary(ctx, req.ContractID, req.BusinessDate)
	if err != nil {
		return nil, fmt.Errorf("fetching transaction summary: %w", err)
	}

	var result *ManagerApprovalResult
	txErr := database.WithTransaction(ctx, h.pool, func(tx pgx.Tx) error {
		// Step 3 — lock the row
		day, err := h.dayRepo.GetForUpdateByBusinessDate(ctx, tx, req.BusinessDate)
		if err != nil {
			return fmt.Errorf("locking workflow day: %w", err)
		}
		if day == nil {
			return &domain.ErrNotFound{
				ContractID:   req.ContractID.String(),
				BusinessDate: req.BusinessDate.Format("2006-01-02"),
			}
		}

		// Step 4 — policy
		if err := h.policy.CanApprove(policy.ApprovalInput{
			CurrentDay:                 day,
			ActorID:                    req.Actor.UserID,
			TransactionCount:           txSummary.Count,
			HasPendingUnreviewed:       txSummary.HasPendingUnreviewed,
			ZeroTransactionAttestation: req.ZeroTransactionAttestation,
			AttestationReason:          req.AttestationReason,
		}); err != nil {
			return err
		}

		// Step 5 — mutate the entity and persist
		now := time.Now().UTC()
		day.CurrentState = vo.StateManagerApprovedEOD
		day.ManagerApprovedAt = &now
		day.ManagerApprovedBy = &req.Actor.UserID
		day.TransactionsLockedAt = &now // lock transactions atomically with approval
		day.UpdatedAt = now
		day.UpdatedBy = req.Actor.UserID
		// Version is incremented inside UpdateState (WHERE id=$1 AND version=$2)

		if err := h.dayRepo.UpdateState(ctx, tx, day); err != nil {
			return err // may be ErrVersionConflict
		}

		// Step 6 — approval record
		approvalID := uuid.New()
		var attestationReason *string
		if req.ZeroTransactionAttestation {
			attestationReason = &req.AttestationReason
		}
		approverRole := workflowperm.PickHighest(req.Actor.Roles)
		if approverRole == "" {
			slog.Warn("approver role snapshot is empty",
				"contract_id", req.ContractID,
				"actor_id", req.Actor.UserID,
				"actor_roles", req.Actor.Roles,
			)
		}
		approvalRec := &entity.ApprovalRecord{
			ID:                approvalID,
			WorkflowDayID:     day.ID,
			ContractID:        req.ContractID,
			BusinessDate:      req.BusinessDate,
			ApproverID:        req.Actor.UserID,
			ApproverUsername:  req.Actor.Username,
			ApproverRole:      approverRole,
			ApprovalStatus:    entity.ApprovalStatusApproved,
			IsZeroTransaction: txSummary.Count == 0,
			AttestationReason: attestationReason,
			ApprovedAt:        now,
			Notes:             req.Notes,
		}
		if err := h.approvalRepo.Insert(ctx, tx, approvalRec); err != nil {
			return fmt.Errorf("inserting approval record: %w", err)
		}

		// Step 7 — transition log
		metadata := map[string]any{
			"transactionCount":           txSummary.Count,
			"hasPendingUnreviewed":       txSummary.HasPendingUnreviewed,
			"zeroTransactionAttestation": txSummary.Count == 0,
		}
		if txSummary.Count == 0 && req.AttestationReason != "" {
			metadata["attestationReason"] = req.AttestationReason
		}

		transition := &entity.WorkflowTransition{
			ID:               uuid.New(),
			WorkflowDayID:    day.ID,
			ContractID:       req.ContractID,
			BusinessDate:     req.BusinessDate,
			FromState:        vo.StateInvestmentDayStarted,
			ToState:          vo.StateManagerApprovedEOD,
			Action:           vo.ActionApprove,
			ActorID:          &req.Actor.UserID,
			ActorType:        req.Actor.ActorType,
			ActorUsername:    req.Actor.Username,
			ActorAccountCode: req.Actor.AccountCode,
			IsAdminOverride:  req.Actor.IsAdminOverride,
			Metadata:         metadata,
			OccurredAt:       now,
			RequestID:        req.Actor.RequestID,
		}
		if err := h.logRepo.Append(ctx, tx, transition); err != nil {
			return fmt.Errorf("appending transition log: %w", err)
		}

		result = &ManagerApprovalResult{
			TransitionID:      transition.ID,
			WorkflowDayID:     day.ID,
			ApprovalID:        approvalID,
			ContractID:        req.ContractID,
			BusinessDate:      req.BusinessDate,
			FromState:         vo.StateInvestmentDayStarted,
			ToState:           vo.StateManagerApprovedEOD,
			OccurredAt:        now,
			IsZeroTransaction: txSummary.Count == 0,
		}
		return nil
	})
	if txErr != nil {
		return nil, txErr
	}
	return result, nil
}
