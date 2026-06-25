package command

import (
	"context"
	"errors"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"

	"github.com/google/uuid"
)

// BatchPermissionChecker gates access to decisions by contract data scope.
// Satisfied by adapter.IAMPermissionPort wired in investment/module.go.
type BatchPermissionChecker interface {
	HasDataPermission(ctx context.Context, userID string, scopeID string) (bool, error)
}

// BatchApproveRequest is the input for approving multiple decision headers.
type BatchApproveRequest struct {
	DecisionNos []string
	Comment     string
	ActorID     uuid.UUID
}

// BatchRejectRequest is the input for rejecting multiple decision headers.
type BatchRejectRequest struct {
	DecisionNos []string
	Reason      string
	ActorID     uuid.UUID
}

// BatchApprovalResult holds the per-decision outcome for batch operations.
type BatchApprovalResult struct {
	DecisionNo string
	OK         bool
	Error      string
}

// DecisionBatchApprovalHandler handles batch approve/reject for the approval screen.
type DecisionBatchApprovalHandler struct {
	decisions  domain.DecisionRepository
	batchActor contract.ApprovalBatchActor
	perms      BatchPermissionChecker
}

// NewDecisionBatchApprovalHandler wires the handler.
func NewDecisionBatchApprovalHandler(
	decisions domain.DecisionRepository,
	batchActor contract.ApprovalBatchActor,
) *DecisionBatchApprovalHandler {
	return &DecisionBatchApprovalHandler{
		decisions:  decisions,
		batchActor: batchActor,
	}
}

// SetBatchActor injects the approval batch actor post-construction.
func (h *DecisionBatchApprovalHandler) SetBatchActor(a contract.ApprovalBatchActor) {
	if h != nil {
		h.batchActor = a
	}
}

// SetPermissionChecker injects the data-permission gate post-construction.
func (h *DecisionBatchApprovalHandler) SetPermissionChecker(p BatchPermissionChecker) {
	if h != nil {
		h.perms = p
	}
}

// BatchApprove approves all supplied decision headers for the current actor.
// Each decision must be in PENDING_APPROVAL and have a pending task assigned
// to actorID. Returns per-decision results; the caller decides how to surface
// partial failures.
func (h *DecisionBatchApprovalHandler) BatchApprove(ctx context.Context, req BatchApproveRequest) ([]BatchApprovalResult, error) {
	if req.ActorID == uuid.Nil {
		return nil, errors.New("actor_id is required")
	}
	if len(req.DecisionNos) == 0 {
		return nil, errors.New("at least one decision_no is required")
	}
	if h.batchActor == nil {
		return nil, errors.New("approval engine not wired")
	}

	results := make([]BatchApprovalResult, 0, len(req.DecisionNos))
	for _, no := range req.DecisionNos {
		res := BatchApprovalResult{DecisionNo: no}
		// Load only the fields needed for the permission check — never expose
		// amount, rationale, status, fund, or portfolio before authorisation.
		ref, err := h.decisions.FindDecisionSubjectRefByNumber(ctx, no)
		if err != nil || ref == nil {
			// Return "not found" uniformly — never leak existence via distinct errors.
			res.Error = "decision not found"
			results = append(results, res)
			continue
		}
		// Data-permission check must come before any approval-state disclosure.
		// Unauthorized callers receive the same "not found" response as for a
		// nonexistent decision so they cannot infer existence from error type.
		if h.perms != nil {
			ok, permErr := h.perms.HasDataPermission(ctx, req.ActorID.String(), ref.ContractID.String())
			if permErr != nil || !ok {
				res.Error = "decision not found"
				results = append(results, res)
				continue
			}
		}
		if ref.ApprovalRequestID == nil {
			res.Error = "no approval request associated with this decision"
			results = append(results, res)
			continue
		}
		if err := h.batchActor.ApproveByRequest(ctx, *ref.ApprovalRequestID, req.ActorID, req.Comment); err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}
		res.OK = true
		results = append(results, res)
	}
	return results, nil
}

// BatchReject rejects all supplied decision headers for the current actor.
func (h *DecisionBatchApprovalHandler) BatchReject(ctx context.Context, req BatchRejectRequest) ([]BatchApprovalResult, error) {
	if req.ActorID == uuid.Nil {
		return nil, errors.New("actor_id is required")
	}
	if len(req.DecisionNos) == 0 {
		return nil, errors.New("at least one decision_no is required")
	}
	if req.Reason == "" {
		return nil, errors.New("reason is required for rejection")
	}
	if h.batchActor == nil {
		return nil, errors.New("approval engine not wired")
	}

	results := make([]BatchApprovalResult, 0, len(req.DecisionNos))
	for _, no := range req.DecisionNos {
		res := BatchApprovalResult{DecisionNo: no}
		ref, err := h.decisions.FindDecisionSubjectRefByNumber(ctx, no)
		if err != nil || ref == nil {
			res.Error = "decision not found"
			results = append(results, res)
			continue
		}
		if h.perms != nil {
			ok, permErr := h.perms.HasDataPermission(ctx, req.ActorID.String(), ref.ContractID.String())
			if permErr != nil || !ok {
				res.Error = "decision not found"
				results = append(results, res)
				continue
			}
		}
		if ref.ApprovalRequestID == nil {
			res.Error = "no approval request associated with this decision"
			results = append(results, res)
			continue
		}
		if err := h.batchActor.RejectByRequest(ctx, *ref.ApprovalRequestID, req.ActorID, req.Reason); err != nil {
			res.Error = err.Error()
			results = append(results, res)
			continue
		}
		res.OK = true
		results = append(results, res)
	}
	return results, nil
}
