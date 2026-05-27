package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions/domain"
	permcode "github.com/neo-kanta/ims-th-solution/backend/internal/permissions/permission"
	"github.com/neo-kanta/ims-th-solution/backend/platform/database"
)

type PermissionChecker interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
}

type Service struct {
	pool    *pgxpool.Pool
	repo    domain.Repository
	checker PermissionChecker
	now     func() time.Time
}

func NewService(pool *pgxpool.Pool, repo domain.Repository, checker PermissionChecker) *Service {
	return &Service{
		pool:    pool,
		repo:    repo,
		checker: checker,
		now:     func() time.Time { return time.Now().UTC() },
	}
}

type CreateRequestInput struct {
	Title            string
	Description      string
	RequestType      string
	RiskLevel        string
	TargetEntityType string
	TargetEntityID   string
	ActorID          uuid.UUID
}

type UpdateRequestInput struct {
	ID               uuid.UUID
	Title            string
	Description      string
	RequestType      string
	RiskLevel        string
	TargetEntityType string
	TargetEntityID   string
	ActorID          uuid.UUID
}

type ItemInput struct {
	ID          uuid.UUID
	RequestID   uuid.UUID
	ItemType    string
	TargetTable string
	TargetID    string
	ActionType  string
	BeforeJSON  json.RawMessage
	AfterJSON   json.RawMessage
	ActorID     uuid.UUID
}

type DecisionInput struct {
	RequestID uuid.UUID
	StepID    uuid.UUID
	ActorID   uuid.UUID
	Comment   string
}

type LabelInput struct {
	RequestID uuid.UUID
	LabelCode string
	LabelID   uuid.UUID
	ActorID   uuid.UUID
}

type CommentInput struct {
	RequestID uuid.UUID
	CommentID uuid.UUID
	ActorID   uuid.UUID
	Comment   string
}

func (s *Service) ListChangeRequests(ctx context.Context, filter domain.ChangeRequestFilter) ([]domain.ChangeRequest, int, error) {
	return s.repo.ListChangeRequests(ctx, filter)
}

func (s *Service) GetChangeRequest(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error) {
	cr, err := s.repo.GetChangeRequest(ctx, id)
	if err != nil || cr != nil {
		return cr, err
	}
	return nil, domain.NotFound("permission change request", id)
}

func (s *Service) CreateChangeRequest(ctx context.Context, in CreateRequestInput) (*domain.ChangeRequest, error) {
	if in.ActorID == uuid.Nil {
		return nil, domain.Invalid("actor_id is required")
	}
	if strings.TrimSpace(in.Title) == "" {
		return nil, domain.Invalid("title is required")
	}
	requestType := normalizeDefault(in.RequestType, domain.RequestTypePermission)
	risk := normalizeDefault(strings.ToUpper(in.RiskLevel), domain.RiskLow)
	no, err := s.repo.NextRequestNo(ctx)
	if err != nil {
		return nil, err
	}
	cr := &domain.ChangeRequest{
		ID:               uuid.New(),
		RequestNo:        no,
		Title:            strings.TrimSpace(in.Title),
		Description:      strings.TrimSpace(in.Description),
		RequestType:      requestType,
		Status:           domain.RequestStatusDraft,
		RiskLevel:        risk,
		TargetEntityType: strings.TrimSpace(in.TargetEntityType),
		TargetEntityID:   strings.TrimSpace(in.TargetEntityID),
		CreatedBy:        in.ActorID,
	}
	if err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.CreateChangeRequest(ctx, tx, cr); err != nil {
			return err
		}
		if err := s.addAutomaticLabels(ctx, tx, cr.ID, cr.RiskLevel, []string{"module:permission"}, in.ActorID); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &in.ActorID, "CREATED", nil, cr, "")
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_CREATED", "permission_change_request", cr.ID.String(), nil, cr)
	}); err != nil {
		return nil, err
	}
	_ = s.RerunChecks(ctx, cr.ID, in.ActorID)
	return s.GetChangeRequest(ctx, cr.ID)
}

func (s *Service) UpdateChangeRequest(ctx context.Context, in UpdateRequestInput) (*domain.ChangeRequest, error) {
	if in.ActorID == uuid.Nil {
		return nil, domain.Invalid("actor_id is required")
	}
	var out *domain.ChangeRequest
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, in.ID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", in.ID)
		}
		if cr.Status != domain.RequestStatusDraft && cr.Status != domain.RequestStatusChanges {
			return domain.InvalidTransition("only draft or changes-requested requests can be edited")
		}
		before := *cr
		if strings.TrimSpace(in.Title) != "" {
			cr.Title = strings.TrimSpace(in.Title)
		}
		cr.Description = strings.TrimSpace(in.Description)
		if strings.TrimSpace(in.RequestType) != "" {
			cr.RequestType = strings.TrimSpace(in.RequestType)
		}
		if strings.TrimSpace(in.RiskLevel) != "" {
			cr.RiskLevel = strings.ToUpper(strings.TrimSpace(in.RiskLevel))
		}
		cr.TargetEntityType = strings.TrimSpace(in.TargetEntityType)
		cr.TargetEntityID = strings.TrimSpace(in.TargetEntityID)
		if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &in.ActorID, "UPDATED", before, cr, "")
		if err := s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_UPDATED", "permission_change_request", cr.ID.String(), before, cr); err != nil {
			return err
		}
		out = cr
		return nil
	})
	if err != nil {
		return nil, err
	}
	_ = s.RerunChecks(ctx, out.ID, in.ActorID)
	return s.GetChangeRequest(ctx, out.ID)
}

func (s *Service) AddItem(ctx context.Context, in ItemInput) (*domain.ChangeItem, error) {
	item := &domain.ChangeItem{
		ID:          uuid.New(),
		RequestID:   in.RequestID,
		ItemType:    normalizeDefault(in.ItemType, "PERMISSION"),
		TargetTable: in.TargetTable,
		TargetID:    in.TargetID,
		ActionType:  strings.ToUpper(strings.TrimSpace(in.ActionType)),
		BeforeJSON:  in.BeforeJSON,
		AfterJSON:   in.AfterJSON,
	}
	if item.ActionType == "" {
		return nil, domain.Invalid("action_type is required")
	}
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, in.RequestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", in.RequestID)
		}
		if cr.Status != domain.RequestStatusDraft && cr.Status != domain.RequestStatusChanges {
			return domain.InvalidTransition("change items can only be changed before review")
		}
		if err := s.repo.AddItem(ctx, tx, item); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &in.ActorID, "ITEM_ADDED", nil, item, "")
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_ITEM_ADDED", "permission_change_item", item.ID.String(), nil, item)
	})
	if err != nil {
		return nil, err
	}
	_ = s.RerunChecks(ctx, in.RequestID, in.ActorID)
	return item, nil
}

func (s *Service) UpdateItem(ctx context.Context, in ItemInput) error {
	item := &domain.ChangeItem{
		ID:          in.ID,
		RequestID:   in.RequestID,
		ItemType:    normalizeDefault(in.ItemType, "PERMISSION"),
		TargetTable: in.TargetTable,
		TargetID:    in.TargetID,
		ActionType:  strings.ToUpper(strings.TrimSpace(in.ActionType)),
		BeforeJSON:  in.BeforeJSON,
		AfterJSON:   in.AfterJSON,
	}
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, in.RequestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", in.RequestID)
		}
		if cr.Status != domain.RequestStatusDraft && cr.Status != domain.RequestStatusChanges {
			return domain.InvalidTransition("change items can only be changed before review")
		}
		if err := s.repo.UpdateItem(ctx, tx, item); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &in.ActorID, "ITEM_UPDATED", nil, item, "")
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_ITEM_UPDATED", "permission_change_item", item.ID.String(), nil, item)
	})
	if err != nil {
		return err
	}
	_ = s.RerunChecks(ctx, in.RequestID, in.ActorID)
	return nil
}

func (s *Service) DeleteItem(ctx context.Context, requestID uuid.UUID, itemID uuid.UUID, actorID uuid.UUID) error {
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", requestID)
		}
		if cr.Status != domain.RequestStatusDraft && cr.Status != domain.RequestStatusChanges {
			return domain.InvalidTransition("change items can only be changed before review")
		}
		if err := s.repo.DeleteItem(ctx, tx, requestID, itemID); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &actorID, "ITEM_DELETED", nil, map[string]any{"item_id": itemID}, "")
		return s.audit(ctx, tx, actorID, "PERMISSION_REQUEST_ITEM_DELETED", "permission_change_item", itemID.String(), nil, nil)
	})
	if err != nil {
		return err
	}
	_ = s.RerunChecks(ctx, requestID, actorID)
	return nil
}

func (s *Service) Submit(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) (*domain.ChangeRequest, error) {
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", requestID)
		}
		if cr.Status != domain.RequestStatusDraft && cr.Status != domain.RequestStatusChanges {
			return domain.InvalidTransition("only draft or changes-requested requests can be submitted")
		}
		itemCount, err := s.repo.CountItems(ctx, requestID)
		if err != nil {
			return err
		}
		if itemCount == 0 {
			return domain.ChecksFailed("request must contain at least one change item", map[string]any{"check_code": "REQUEST_HAS_CHANGE_ITEM"})
		}
		setting, err := s.repo.FindApprovalSetting(ctx, cr.RequestType, cr.RiskLevel)
		if err != nil {
			return err
		}
		if setting == nil {
			return domain.Invalid("no active approval workflow setting found")
		}
		checks, err := s.runChecks(ctx, cr)
		if err != nil {
			return err
		}
		if err := s.repo.ReplaceChecks(ctx, tx, requestID, checks); err != nil {
			return err
		}
		if setting.FailedCheckBlocksSubmit && hasBlocking(checks) {
			return domain.ChecksFailed("required checks failed; request cannot be submitted", map[string]any{"request_id": requestID})
		}
		if err := s.repo.DeleteRequestApprovalState(ctx, tx, requestID); err != nil {
			return err
		}
		for i, templateStep := range setting.Steps {
			status := domain.StepStatusNotStarted
			var startedAt *time.Time
			if i == 0 {
				now := s.now()
				startedAt = &now
				status = domain.StepStatusPending
			}
			step := &domain.ApprovalStep{
				ID:                   uuid.New(),
				RequestID:            requestID,
				WorkflowSettingID:    setting.ID,
				StepNo:               templateStep.StepNo,
				StepName:             templateStep.StepName,
				ApprovalMode:         templateStep.ApprovalMode,
				Status:               status,
				MinApprovalsRequired: templateStep.MinApprovalsRequired,
				StartedAt:            startedAt,
			}
			if err := s.repo.CreateRequestStep(ctx, tx, step); err != nil {
				return err
			}
			if err := s.materializeApprovers(ctx, tx, cr, step, templateStep); err != nil {
				return err
			}
		}
		now := s.now()
		before := *cr
		cr.Status = domain.RequestStatusReadyForReview
		cr.SubmittedAt = &now
		if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
			return err
		}
		if err := s.addAutomaticLabels(ctx, tx, requestID, cr.RiskLevel, []string{"awaiting-review"}, actorID); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &actorID, "SUBMITTED", before, cr, "")
		return s.audit(ctx, tx, actorID, "PERMISSION_REQUEST_SUBMITTED", "permission_change_request", cr.ID.String(), before, cr)
	})
	if err != nil {
		return nil, err
	}
	return s.GetChangeRequest(ctx, requestID)
}

func (s *Service) Approve(ctx context.Context, in DecisionInput) (*domain.ChangeRequest, error) {
	return s.decide(ctx, in, domain.ReviewerStatusApproved)
}

func (s *Service) RequestChanges(ctx context.Context, in DecisionInput) (*domain.ChangeRequest, error) {
	return s.decide(ctx, in, domain.ReviewerStatusChanges)
}

func (s *Service) Reject(ctx context.Context, in DecisionInput) (*domain.ChangeRequest, error) {
	return s.decide(ctx, in, domain.ReviewerStatusRejected)
}

func (s *Service) Merge(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) (*domain.ChangeRequest, error) {
	if err := s.requirePermission(ctx, actorID, permcode.CodeChangeRequestMerge); err != nil {
		return nil, err
	}
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", requestID)
		}
		if cr.Status != domain.RequestStatusApproved {
			return domain.MergeNotAllowed("only approved requests can be merged")
		}
		if time.Since(cr.UpdatedAt) > 30*24*time.Hour {
			return domain.MergeNotAllowed("permission request is stale and must be refreshed")
		}
		steps, err := s.repo.ListApprovalSteps(ctx, requestID)
		if err != nil {
			return err
		}
		for _, step := range steps {
			if step.Status != domain.StepStatusApproved && step.Status != domain.StepStatusSkipped {
				return domain.MergeNotAllowed("all required approval steps must be approved")
			}
		}
		blocked, err := s.repo.HasBlockingChecks(ctx, requestID)
		if err != nil {
			return err
		}
		if blocked {
			return domain.MergeNotAllowed("required checks failed")
		}
		items, err := s.repo.ListItems(ctx, requestID)
		if err != nil {
			return err
		}
		for _, item := range items {
			if err := s.applyItem(ctx, tx, cr, item, actorID); err != nil {
				return err
			}
		}
		now := s.now()
		before := *cr
		cr.Status = domain.RequestStatusMerged
		cr.MergedAt = &now
		cr.MergedBy = &actorID
		if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
			return err
		}
		s.event(ctx, tx, cr.ID, nil, &actorID, "MERGED", before, cr, "")
		return s.audit(ctx, tx, actorID, "PERMISSION_REQUEST_MERGED", "permission_change_request", cr.ID.String(), before, cr)
	})
	if err != nil {
		return nil, err
	}
	return s.GetChangeRequest(ctx, requestID)
}

func (s *Service) Close(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) (*domain.ChangeRequest, error) {
	return s.finish(ctx, requestID, actorID, domain.RequestStatusClosed, "PERMISSION_REQUEST_CLOSED")
}

func (s *Service) Cancel(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) (*domain.ChangeRequest, error) {
	return s.finish(ctx, requestID, actorID, domain.RequestStatusCancelled, "PERMISSION_REQUEST_CANCELLED")
}

func (s *Service) AddComment(ctx context.Context, in CommentInput) (*domain.Comment, error) {
	if strings.TrimSpace(in.Comment) == "" {
		return nil, domain.Invalid("comment is required")
	}
	comment := &domain.Comment{ID: uuid.New(), RequestID: in.RequestID, UserID: in.ActorID, Comment: strings.TrimSpace(in.Comment)}
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.AddComment(ctx, tx, comment); err != nil {
			return err
		}
		s.event(ctx, tx, in.RequestID, nil, &in.ActorID, "COMMENTED", nil, comment, "")
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_COMMENTED", "permission_change_request", in.RequestID.String(), nil, comment)
	})
	return comment, err
}

func (s *Service) UpdateComment(ctx context.Context, in CommentInput) error {
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		return s.repo.UpdateComment(ctx, tx, in.RequestID, in.CommentID, in.ActorID, strings.TrimSpace(in.Comment))
	})
}

func (s *Service) DeleteComment(ctx context.Context, in CommentInput) error {
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		return s.repo.DeleteComment(ctx, tx, in.RequestID, in.CommentID, in.ActorID)
	})
}

func (s *Service) AddLabel(ctx context.Context, in LabelInput) error {
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.AddRequestLabel(ctx, tx, in.RequestID, in.LabelCode, in.ActorID); err != nil {
			return err
		}
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_LABEL_ADDED", "permission_change_request", in.RequestID.String(), nil, map[string]string{"label_code": in.LabelCode})
	})
}

func (s *Service) RemoveLabel(ctx context.Context, in LabelInput) error {
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.RemoveRequestLabel(ctx, tx, in.RequestID, in.LabelID); err != nil {
			return err
		}
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_LABEL_REMOVED", "permission_change_request", in.RequestID.String(), map[string]string{"label_id": in.LabelID.String()}, nil)
	})
}

func (s *Service) RerunChecks(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) error {
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", requestID)
		}
		checks, err := s.runChecks(ctx, cr)
		if err != nil {
			return err
		}
		if err := s.repo.ReplaceChecks(ctx, tx, requestID, checks); err != nil {
			return err
		}
		labels := []string{}
		if hasBlocking(checks) {
			labels = append(labels, "blocked")
			for _, c := range checks {
				if strings.Contains(strings.ToLower(c.CheckCode), "compliance") && c.Status == domain.CheckStatusFailed {
					labels = append(labels, "compliance-blocker")
					break
				}
			}
		}
		if cr.Status == domain.RequestStatusApproved && !hasBlocking(checks) {
			labels = append(labels, "ready-to-merge")
		}
		if err := s.addAutomaticLabels(ctx, tx, requestID, cr.RiskLevel, labels, actorID); err != nil {
			return err
		}
		s.event(ctx, tx, requestID, nil, &actorID, "CHECKS_RUN", nil, map[string]any{"count": len(checks)}, "")
		return nil
	})
}

func (s *Service) ListRoles(ctx context.Context) ([]domain.Role, error) { return s.repo.ListRoles(ctx) }
func (s *Service) ListLabels(ctx context.Context) ([]domain.Label, error) {
	return s.repo.ListLabels(ctx)
}
func (s *Service) UpsertLabel(ctx context.Context, label *domain.Label, actorID uuid.UUID) error {
	if strings.TrimSpace(label.LabelCode) == "" {
		return domain.Invalid("label_code is required")
	}
	if strings.TrimSpace(label.LabelName) == "" {
		return domain.Invalid("label_name is required")
	}
	if strings.TrimSpace(label.LabelType) == "" {
		label.LabelType = "MANUAL"
	}
	if strings.TrimSpace(label.Color) == "" {
		label.Color = "#6e7781"
	}
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.UpsertLabel(ctx, tx, label); err != nil {
			return err
		}
		return s.audit(ctx, tx, actorID, "PERMISSION_LABEL_UPSERTED", "permission_label", label.LabelCode, nil, label)
	})
}
func (s *Service) ListUsers(ctx context.Context, search string, page int, limit int) ([]domain.UserSummary, int, error) {
	return s.repo.ListUsers(ctx, search, page, limit)
}
func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserSummary, error) {
	return s.repo.GetUser(ctx, id)
}
func (s *Service) ListGroups(ctx context.Context, search string, page int, limit int) ([]domain.GroupSummary, int, error) {
	return s.repo.ListGroups(ctx, search, page, limit)
}
func (s *Service) GetGroup(ctx context.Context, id uuid.UUID) (*domain.GroupSummary, error) {
	return s.repo.GetGroup(ctx, id)
}
func (s *Service) ListFunctionDefinitions(ctx context.Context) ([]domain.FunctionDefinition, error) {
	return s.repo.ListFunctionDefinitions(ctx)
}
func (s *Service) ListFunctionRights(ctx context.Context) ([]domain.FunctionRight, error) {
	return s.repo.ListFunctionRights(ctx)
}
func (s *Service) ListDataRights(ctx context.Context) ([]domain.DataRight, error) {
	return s.repo.ListDataRights(ctx)
}
func (s *Service) EffectivePermissions(ctx context.Context, userID uuid.UUID) (*domain.EffectivePermissions, error) {
	return s.repo.EffectivePermissions(ctx, userID)
}
func (s *Service) ListApprovalSettings(ctx context.Context) ([]domain.ApprovalSetting, error) {
	return s.repo.ListApprovalSettings(ctx)
}
func (s *Service) GetApprovalSetting(ctx context.Context, id uuid.UUID) (*domain.ApprovalSetting, error) {
	return s.repo.GetApprovalSetting(ctx, id)
}
func (s *Service) ListAuditLogs(ctx context.Context, page int, limit int) ([]domain.AuditLog, int, error) {
	return s.repo.ListAuditLogs(ctx, page, limit)
}
func (s *Service) ListNotificationSettings(ctx context.Context) ([]domain.NotificationSetting, error) {
	return s.repo.ListNotificationSettings(ctx)
}
func (s *Service) UpdateNotificationSetting(ctx context.Context, setting domain.NotificationSetting, actorID uuid.UUID) error {
	return database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		if err := s.repo.UpdateNotificationSetting(ctx, tx, setting); err != nil {
			return err
		}
		return s.audit(ctx, tx, actorID, "PERMISSION_NOTIFICATION_SETTING_UPDATED", "notification_setting", setting.ID.String(), nil, setting)
	})
}

func (s *Service) decide(ctx context.Context, in DecisionInput, decision string) (*domain.ChangeRequest, error) {
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, in.RequestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", in.RequestID)
		}
		if cr.Status != domain.RequestStatusReadyForReview {
			return domain.InvalidTransition("request must be ready for review")
		}
		if cr.CreatedBy == in.ActorID {
			setting, _ := s.repo.FindApprovalSetting(ctx, cr.RequestType, cr.RiskLevel)
			if setting == nil || !setting.AllowCreatorApproval {
				return domain.ApprovalNotAllowed("creator self-approval is not allowed")
			}
		}
		step, err := s.resolveDecisionStep(ctx, in.RequestID, in.StepID)
		if err != nil {
			return err
		}
		if step == nil {
			return domain.InvalidTransition("no pending approval step")
		}
		if step.Status != domain.StepStatusPending {
			return domain.InvalidTransition("only the current pending step can be reviewed")
		}
		roleCode, err := s.validateApprover(ctx, cr, step, in.ActorID)
		if err != nil {
			return err
		}
		if err := s.repo.UpdateStepApproverDecision(ctx, tx, in.RequestID, step.ID, in.ActorID, roleCode, decision, in.Comment); err != nil {
			return err
		}
		now := s.now()
		before := *cr
		switch decision {
		case domain.ReviewerStatusApproved:
			step.ApprovalsReceived++
			if step.ApprovalsReceived >= step.MinApprovalsRequired {
				step.Status = domain.StepStatusApproved
				step.CompletedAt = &now
			}
			if err := s.repo.UpdateApprovalStep(ctx, tx, step); err != nil {
				return err
			}
			if step.Status == domain.StepStatusApproved {
				allDone, err := s.repo.StartNextApprovalStep(ctx, tx, in.RequestID, step.StepNo)
				if err != nil {
					return err
				}
				if allDone {
					cr.Status = domain.RequestStatusApproved
					cr.ApprovedAt = &now
					cr.ApprovedBy = &in.ActorID
					if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
						return err
					}
				}
			}
		case domain.ReviewerStatusChanges:
			step.Status = domain.StepStatusChanges
			step.CompletedAt = &now
			cr.Status = domain.RequestStatusChanges
			if err := s.repo.UpdateApprovalStep(ctx, tx, step); err != nil {
				return err
			}
			if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
				return err
			}
			_ = s.repo.AddRequestLabel(ctx, tx, in.RequestID, "changes-requested", in.ActorID)
		case domain.ReviewerStatusRejected:
			step.Status = domain.StepStatusRejected
			step.CompletedAt = &now
			cr.Status = domain.RequestStatusRejected
			cr.RejectedAt = &now
			cr.RejectedBy = &in.ActorID
			cr.RejectionReason = in.Comment
			if err := s.repo.UpdateApprovalStep(ctx, tx, step); err != nil {
				return err
			}
			if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
				return err
			}
		}
		eventType := "STEP_" + decision
		s.event(ctx, tx, in.RequestID, &step.ID, &in.ActorID, eventType, before, cr, in.Comment)
		return s.audit(ctx, tx, in.ActorID, "PERMISSION_REQUEST_"+decision, "permission_change_request", in.RequestID.String(), before, cr)
	})
	if err != nil {
		return nil, err
	}
	_ = s.RerunChecks(ctx, in.RequestID, in.ActorID)
	return s.GetChangeRequest(ctx, in.RequestID)
}

func (s *Service) resolveDecisionStep(ctx context.Context, requestID uuid.UUID, stepID uuid.UUID) (*domain.ApprovalStep, error) {
	if stepID != uuid.Nil {
		return s.repo.GetApprovalStep(ctx, requestID, stepID)
	}
	return s.repo.GetCurrentPendingStep(ctx, requestID)
}

func (s *Service) validateApprover(ctx context.Context, cr *domain.ChangeRequest, step *domain.ApprovalStep, actorID uuid.UUID) (string, error) {
	user, err := s.repo.GetUser(ctx, actorID)
	if err != nil {
		return "", err
	}
	if user == nil || !user.IsActive || user.IsLocked {
		return "", domain.ApprovalNotAllowed("approver must be active and unlocked")
	}
	roles, err := s.repo.GetUserApprovedRoles(ctx, actorID)
	if err != nil {
		return "", err
	}
	for _, approver := range step.Approvers {
		if approver.ApproverUserID != nil && *approver.ApproverUserID == actorID {
			return approver.ApproverRoleCode, nil
		}
		for _, role := range roles {
			if role.RoleCode == approver.ApproverRoleCode && role.CanApproveRoleAssignment {
				return role.RoleCode, nil
			}
		}
		if approver.ApproverRoleCode == "REQUEST_TARGET_OWNER" && s.isTargetOwner(ctx, cr, actorID) {
			return "REQUEST_TARGET_OWNER", nil
		}
	}
	return "", domain.ApprovalNotAllowed("actor is not an eligible reviewer / approver for the current step")
}

func (s *Service) materializeApprovers(ctx context.Context, tx pgx.Tx, cr *domain.ChangeRequest, step *domain.ApprovalStep, template domain.ApprovalSettingStep) error {
	switch template.ApproverType {
	case "USER":
		return s.repo.CreateStepApprover(ctx, tx, &domain.StepApprover{
			ID:             uuid.New(),
			RequestID:      cr.ID,
			ApprovalStepID: step.ID,
			ApproverUserID: template.RequiredUserID,
			ApprovalStatus: domain.ReviewerStatusPending,
		})
	case "REQUEST_TARGET_OWNER":
		owner := s.targetOwner(ctx, cr)
		return s.repo.CreateStepApprover(ctx, tx, &domain.StepApprover{
			ID:               uuid.New(),
			RequestID:        cr.ID,
			ApprovalStepID:   step.ID,
			ApproverUserID:   owner,
			ApproverRoleCode: "REQUEST_TARGET_OWNER",
			ApprovalStatus:   domain.ReviewerStatusPending,
		})
	default:
		return s.repo.CreateStepApprover(ctx, tx, &domain.StepApprover{
			ID:               uuid.New(),
			RequestID:        cr.ID,
			ApprovalStepID:   step.ID,
			ApproverRoleCode: template.RequiredRoleCode,
			ApprovalStatus:   domain.ReviewerStatusPending,
		})
	}
}

func (s *Service) applyItem(ctx context.Context, tx pgx.Tx, cr *domain.ChangeRequest, item domain.ChangeItem, actorID uuid.UUID) error {
	after := map[string]any{}
	if len(item.AfterJSON) > 0 {
		if err := json.Unmarshal(item.AfterJSON, &after); err != nil {
			return domain.Invalid("invalid after_json for change item")
		}
	}
	switch strings.ToUpper(item.ActionType) {
	case "ASSIGN_ROLE":
		userID, err := uuidFromMap(after, "user_id")
		if err != nil {
			return err
		}
		role, err := s.roleFromItem(ctx, after)
		if err != nil {
			return err
		}
		if err := s.validateRoleAssignment(ctx, cr.CreatedBy, role); err != nil {
			return err
		}
		if err := s.repo.InsertRoleAssignment(ctx, tx, userID, role.ID, actorID); err != nil {
			return err
		}
	case "UPSERT_FUNCTION_RIGHT":
		subjectID, err := uuidFromMap(after, "subject_id")
		if err != nil {
			return err
		}
		code, _ := after["permission_code"].(string)
		if strings.TrimSpace(code) == "" {
			return domain.Invalid("permission_code is required")
		}
		actor := actorID
		if err := s.repo.UpsertFunctionRight(ctx, tx, domain.FunctionRight{
			SubjectType:       strings.ToUpper(stringFromMap(after, "subject_type", "GROUP")),
			SubjectID:         subjectID,
			PermissionCode:    code,
			CanView:           boolFromMap(after, "can_view"),
			CanSearch:         boolFromMap(after, "can_search"),
			CanAdd:            boolFromMap(after, "can_add"),
			CanEdit:           boolFromMap(after, "can_edit"),
			CanDelete:         boolFromMap(after, "can_delete"),
			CanApprove:        boolFromMap(after, "can_approve"),
			CanRevokeApproval: boolFromMap(after, "can_revoke_approval"),
			CanExport:         boolFromMap(after, "can_export"),
			CanConfigure:      boolFromMap(after, "can_configure"),
			ApprovedBy:        &actor,
		}); err != nil {
			return err
		}
	case "UPSERT_DATA_RIGHT":
		subjectID, err := uuidFromMap(after, "subject_id")
		if err != nil {
			return err
		}
		actor := actorID
		if err := s.repo.UpsertDataRight(ctx, tx, domain.DataRight{
			SubjectType: strings.ToUpper(stringFromMap(after, "subject_type", "USER")),
			SubjectID:   subjectID,
			FundID:      uuidPtrFromMap(after, "fund_id"),
			ContractID:  uuidPtrFromMap(after, "contract_id"),
			PortfolioID: uuidPtrFromMap(after, "portfolio_id"),
			AccessLevel: strings.ToUpper(stringFromMap(after, "access_level", "READ")),
			ApprovedBy:  &actor,
		}); err != nil {
			return err
		}
	case "ADD_GROUP_MEMBER":
		userID, err := uuidFromMap(after, "user_id")
		if err != nil {
			return err
		}
		groupID, err := uuidFromMap(after, "group_id")
		if err != nil {
			return err
		}
		return s.repo.InsertGroupMembership(ctx, tx, userID, groupID, actorID)
	case "REMOVE_GROUP_MEMBER":
		userID, err := uuidFromMap(after, "user_id")
		if err != nil {
			return err
		}
		groupID, err := uuidFromMap(after, "group_id")
		if err != nil {
			return err
		}
		return s.repo.DeleteGroupMembership(ctx, tx, userID, groupID)
	default:
		return domain.Invalid("unsupported action_type: " + item.ActionType)
	}
	return s.audit(ctx, tx, actorID, "PERMISSION_CHANGE_ITEM_APPLIED", item.TargetTable, item.TargetID, item.BeforeJSON, item.AfterJSON)
}

func (s *Service) runChecks(ctx context.Context, cr *domain.ChangeRequest) ([]domain.Check, error) {
	items, err := s.repo.ListItems(ctx, cr.ID)
	if err != nil {
		return nil, err
	}
	checks := []domain.Check{}
	add := func(code, name, status, severity, msg string) {
		checks = append(checks, domain.Check{ID: uuid.New(), RequestID: cr.ID, CheckCode: code, CheckName: name, Status: status, Severity: severity, Message: msg})
	}
	if len(items) == 0 {
		add("REQUEST_HAS_CHANGE_ITEM", "Request has at least one change item", domain.CheckStatusFailed, domain.CheckSeverityBlocker, "Add at least one change item before submitting.")
	} else {
		add("REQUEST_HAS_CHANGE_ITEM", "Request has at least one change item", domain.CheckStatusPassed, domain.CheckSeverityInfo, "Change items found.")
	}
	if cr.Status == domain.RequestStatusReadyForReview {
		add("CREATOR_SELF_APPROVAL_BLOCKED", "Creator cannot approve own request by default", domain.CheckStatusPassed, domain.CheckSeverityInfo, "Creator self-approval is blocked by workflow settings.")
	}
	for _, item := range items {
		after := map[string]any{}
		_ = json.Unmarshal(item.AfterJSON, &after)
		switch strings.ToUpper(item.ActionType) {
		case "ASSIGN_ROLE":
			role, err := s.roleFromItem(ctx, after)
			if err != nil {
				add("TARGET_ROLE_EXISTS", "Target role exists and is active", domain.CheckStatusFailed, domain.CheckSeverityBlocker, err.Error())
				continue
			}
			add("TARGET_ROLE_EXISTS", "Target role exists and is active", domain.CheckStatusPassed, domain.CheckSeverityInfo, role.RoleCode)
			if err := s.validateRoleAssignment(ctx, cr.CreatedBy, role); err != nil {
				add("ROLE_PRIORITY_SCOPE", "Actor priority and assignment scope allow target role", domain.CheckStatusFailed, domain.CheckSeverityBlocker, err.Error())
			} else {
				add("ROLE_PRIORITY_SCOPE", "Actor priority and assignment scope allow target role", domain.CheckStatusPassed, domain.CheckSeverityInfo, "Role priority and scope are valid.")
			}
			if role.RoleCode == "INTERNAL_AUDITOR" {
				add("INTERNAL_AUDITOR_READ_ONLY", "Internal auditor remains read-only by default", domain.CheckStatusPassed, domain.CheckSeverityInfo, "Internal auditor role has no request or approval authority.")
			}
		case "UPSERT_FUNCTION_RIGHT":
			code := stringFromMap(after, "permission_code", "")
			exists, err := s.repo.FunctionDefinitionExists(ctx, code)
			if err != nil {
				return nil, err
			}
			if !exists {
				add("FUNCTION_PERMISSION_EXISTS", "Function permission code exists", domain.CheckStatusFailed, domain.CheckSeverityBlocker, "Unknown function permission code: "+code)
			} else {
				add("FUNCTION_PERMISSION_EXISTS", "Function permission code exists", domain.CheckStatusPassed, domain.CheckSeverityInfo, code)
			}
			high := isHighRiskPermission(code, after)
			if high && cr.RiskLevel != domain.RiskHigh && cr.RiskLevel != domain.RiskCritical {
				add("HIGH_RISK_PERMISSION_WORKFLOW", "Admin/export/approve/configure permission uses high-risk workflow", domain.CheckStatusFailed, domain.CheckSeverityBlocker, "High-risk permission requires HIGH or CRITICAL risk level.")
			} else {
				add("HIGH_RISK_PERMISSION_WORKFLOW", "High-risk permission workflow", domain.CheckStatusPassed, domain.CheckSeverityInfo, "Risk level is acceptable.")
			}
		case "UPSERT_DATA_RIGHT":
			if uuidPtrFromMap(after, "fund_id") == nil && uuidPtrFromMap(after, "contract_id") == nil && uuidPtrFromMap(after, "portfolio_id") == nil {
				add("DATA_TARGET_EXISTS", "Fund / contract / portfolio target exists", domain.CheckStatusFailed, domain.CheckSeverityBlocker, "At least one data target is required.")
			} else {
				add("DATA_TARGET_EXISTS", "Fund / contract / portfolio target exists", domain.CheckStatusPassed, domain.CheckSeverityInfo, "Data target supplied.")
			}
		}
	}
	add("RESIGNED_STATUS_SOURCE", "Resigned-user status source", domain.CheckStatusWarning, domain.CheckSeverityWarning, "Current IAM schema has no resigned flag; check is informational until HR status exists.")
	return checks, nil
}

func (s *Service) validateRoleAssignment(ctx context.Context, actorID uuid.UUID, target *domain.Role) error {
	roles, err := s.repo.GetUserApprovedRoles(ctx, actorID)
	if err != nil {
		return err
	}
	for _, actorRole := range roles {
		if !actorRole.CanRequestRoleAssignment {
			continue
		}
		if actorRole.PriorityRank <= target.PriorityRank {
			continue
		}
		if scopeAllows(actorRole.AssignmentScope, target.AssignmentScope) {
			return nil
		}
	}
	return &domain.PermissionError{
		Code:    domain.CodeRolePriorityDenied,
		Message: "actor role priority and assignment scope do not allow assigning target role",
		Status:  403,
		Details: map[string]any{"target_role": target.RoleCode},
	}
}

func (s *Service) roleFromItem(ctx context.Context, after map[string]any) (*domain.Role, error) {
	if id, err := uuidFromMap(after, "role_id"); err == nil {
		role, err := s.repo.GetRoleByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if role != nil && role.IsActive {
			return role, nil
		}
	}
	code := stringFromMap(after, "role_code", "")
	if code != "" {
		role, err := s.repo.GetRoleByCode(ctx, code)
		if err != nil {
			return nil, err
		}
		if role != nil && role.IsActive {
			return role, nil
		}
	}
	return nil, domain.NotFound("permission role", stringFromMap(after, "role_code", after["role_id"]))
}

func (s *Service) finish(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID, status string, auditAction string) (*domain.ChangeRequest, error) {
	err := database.WithTransaction(ctx, s.pool, func(tx pgx.Tx) error {
		cr, err := s.repo.LockChangeRequest(ctx, tx, requestID)
		if err != nil {
			return err
		}
		if cr == nil {
			return domain.NotFound("permission change request", requestID)
		}
		before := *cr
		now := s.now()
		cr.Status = status
		cr.ClosedAt = &now
		cr.ClosedBy = &actorID
		if err := s.repo.UpdateChangeRequest(ctx, tx, cr); err != nil {
			return err
		}
		s.event(ctx, tx, requestID, nil, &actorID, status, before, cr, "")
		return s.audit(ctx, tx, actorID, auditAction, "permission_change_request", requestID.String(), before, cr)
	})
	if err != nil {
		return nil, err
	}
	return s.GetChangeRequest(ctx, requestID)
}

func (s *Service) requirePermission(ctx context.Context, actorID uuid.UUID, code string) error {
	if s.checker == nil {
		return nil
	}
	ok, err := s.checker.HasFunctionPermission(ctx, actorID.String(), code)
	if err != nil {
		return err
	}
	if !ok {
		return domain.Forbidden("insufficient permission: " + code)
	}
	return nil
}

func (s *Service) addAutomaticLabels(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, risk string, labels []string, actorID uuid.UUID) error {
	if risk != "" {
		labels = append(labels, "risk:"+strings.ToLower(risk))
	}
	for _, label := range labels {
		if strings.TrimSpace(label) == "" {
			continue
		}
		if err := s.repo.AddRequestLabel(ctx, tx, requestID, label, actorID); err != nil {
			return err
		}
	}
	return nil
}

func (s *Service) audit(ctx context.Context, tx pgx.Tx, actorID uuid.UUID, action string, entityType string, entityID string, before any, after any) error {
	return s.repo.AppendAuditLog(ctx, tx, domain.AuditLog{
		ID:          uuid.New(),
		ActorUserID: &actorID,
		Action:      action,
		Module:      "permission",
		EntityType:  entityType,
		EntityID:    entityID,
		BeforeJSON:  mustJSON(before),
		AfterJSON:   mustJSON(after),
	})
}

func (s *Service) event(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stepID *uuid.UUID, actorID *uuid.UUID, eventType string, before any, after any, comment string) {
	_ = s.repo.AppendEvent(ctx, tx, domain.WorkflowEvent{
		ID:             uuid.New(),
		RequestID:      requestID,
		ApprovalStepID: stepID,
		ActorUserID:    actorID,
		EventType:      eventType,
		BeforeJSON:     mustJSON(before),
		AfterJSON:      mustJSON(after),
		Comment:        comment,
	})
}

func (s *Service) targetOwner(ctx context.Context, cr *domain.ChangeRequest) *uuid.UUID {
	// The repository intentionally keeps investment internals out of the
	// permission domain. Until a target-owner port exists, use assigned_to as
	// the explicit owner override when provided.
	return cr.AssignedTo
}

func (s *Service) isTargetOwner(ctx context.Context, cr *domain.ChangeRequest, actorID uuid.UUID) bool {
	owner := s.targetOwner(ctx, cr)
	return owner != nil && *owner == actorID
}

func hasBlocking(checks []domain.Check) bool {
	for _, check := range checks {
		if check.Status == domain.CheckStatusFailed && check.Severity == domain.CheckSeverityBlocker {
			return true
		}
	}
	return false
}

func scopeAllows(actor string, target string) bool {
	actor = strings.ToUpper(strings.TrimSpace(actor))
	target = strings.ToUpper(strings.TrimSpace(target))
	if actor == "ALL_BUSINESS" {
		return true
	}
	if actor == target {
		return true
	}
	switch actor {
	case "INVESTMENT":
		return target == "INVESTMENT" || target == "RESEARCH"
	case "OPERATIONS":
		return target == "OPERATIONS" || target == "TA"
	case "TECHNICAL_ONLY":
		return target == "TECHNICAL_ONLY"
	default:
		return false
	}
}

func isHighRiskPermission(code string, after map[string]any) bool {
	c := strings.ToLower(code)
	if strings.Contains(c, "admin") || strings.Contains(c, "export") || strings.Contains(c, "approve") || strings.Contains(c, "configure") {
		return true
	}
	return boolFromMap(after, "can_approve") || boolFromMap(after, "can_export") || boolFromMap(after, "can_configure")
}

func normalizeDefault(value string, fallback string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return fallback
	}
	return value
}

func uuidFromMap(values map[string]any, key string) (uuid.UUID, error) {
	raw, ok := values[key]
	if !ok {
		return uuid.Nil, domain.Invalid(key + " is required")
	}
	id, err := uuid.Parse(fmt.Sprint(raw))
	if err != nil {
		return uuid.Nil, domain.Invalid(key + " must be a UUID")
	}
	return id, nil
}

func uuidPtrFromMap(values map[string]any, key string) *uuid.UUID {
	raw, ok := values[key]
	if !ok || raw == nil || fmt.Sprint(raw) == "" {
		return nil
	}
	id, err := uuid.Parse(fmt.Sprint(raw))
	if err != nil {
		return nil
	}
	return &id
}

func boolFromMap(values map[string]any, key string) bool {
	raw, ok := values[key]
	if !ok {
		return false
	}
	v, ok := raw.(bool)
	return ok && v
}

func stringFromMap(values map[string]any, key string, fallback any) string {
	raw, ok := values[key]
	if !ok || raw == nil || fmt.Sprint(raw) == "" {
		return fmt.Sprint(fallback)
	}
	return strings.TrimSpace(fmt.Sprint(raw))
}

func mustJSON(value any) json.RawMessage {
	if value == nil {
		return json.RawMessage("{}")
	}
	if raw, ok := value.(json.RawMessage); ok {
		if len(raw) == 0 {
			return json.RawMessage("{}")
		}
		return raw
	}
	b, err := json.Marshal(value)
	if err != nil {
		return json.RawMessage("{}")
	}
	return b
}
