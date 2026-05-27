package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/permissions/domain"
)

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) NextRequestNo(ctx context.Context) (string, error) {
	var n int64
	if err := r.pool.QueryRow(ctx, "SELECT nextval('permission_change_request_no_seq')").Scan(&n); err != nil {
		return "", fmt.Errorf("next permission request number: %w", err)
	}
	return fmt.Sprintf("PCR-%06d", n), nil
}

func (r *PostgresRepository) ListChangeRequests(ctx context.Context, filter domain.ChangeRequestFilter) ([]domain.ChangeRequest, int, error) {
	where := []string{"1=1"}
	args := []any{}
	idx := 1
	if filter.Status != "" {
		where = append(where, fmt.Sprintf("cr.status = $%d", idx))
		args = append(args, filter.Status)
		idx++
	}
	if filter.RiskLevel != "" {
		where = append(where, fmt.Sprintf("cr.risk_level = $%d", idx))
		args = append(args, filter.RiskLevel)
		idx++
	}
	if filter.Requester != nil {
		where = append(where, fmt.Sprintf("cr.created_by = $%d", idx))
		args = append(args, *filter.Requester)
		idx++
	}
	if filter.Reviewer != nil {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1 FROM permission_request_step_approvers a
			WHERE a.request_id = cr.id AND a.approver_user_id = $%d
		)`, idx))
		args = append(args, *filter.Reviewer)
		idx++
	}
	if filter.Label != "" {
		where = append(where, fmt.Sprintf(`EXISTS (
			SELECT 1
			FROM permission_change_request_labels rl
			JOIN permission_labels l ON l.id = rl.label_id
			WHERE rl.request_id = cr.id AND l.label_code = $%d
		)`, idx))
		args = append(args, filter.Label)
		idx++
	}
	if filter.Search != "" {
		where = append(where, fmt.Sprintf("(cr.request_no ILIKE $%d OR cr.title ILIKE $%d)", idx, idx))
		args = append(args, "%"+filter.Search+"%")
		idx++
	}
	if filter.DateFrom != nil {
		where = append(where, fmt.Sprintf("cr.created_at >= $%d", idx))
		args = append(args, *filter.DateFrom)
		idx++
	}
	if filter.DateTo != nil {
		where = append(where, fmt.Sprintf("cr.created_at <= $%d", idx))
		args = append(args, *filter.DateTo)
		idx++
	}

	whereSQL := strings.Join(where, " AND ")
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM permission_change_requests cr WHERE "+whereSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count permission change requests: %w", err)
	}

	limit, offset := pageLimit(filter.Page, filter.Limit)
	args = append(args, limit, offset)
	rows, err := r.pool.Query(ctx, changeRequestSelect+" WHERE "+whereSQL+
		fmt.Sprintf(" ORDER BY cr.updated_at DESC, cr.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1), args...)
	if err != nil {
		return nil, 0, fmt.Errorf("list permission change requests: %w", err)
	}
	defer rows.Close()

	out := []domain.ChangeRequest{}
	for rows.Next() {
		cr, err := scanChangeRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		labels, err := r.ListRequestLabels(ctx, cr.ID)
		if err != nil {
			return nil, 0, err
		}
		cr.Labels = labels
		out = append(out, *cr)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) GetChangeRequest(ctx context.Context, id uuid.UUID) (*domain.ChangeRequest, error) {
	cr, err := r.getChangeRequest(ctx, r.pool, id, false)
	if err != nil || cr == nil {
		return cr, err
	}
	cr.Items, _ = r.ListItems(ctx, id)
	cr.Steps, _ = r.ListApprovalSteps(ctx, id)
	cr.Comments, _ = r.ListComments(ctx, id)
	cr.Checks, _ = r.ListChecks(ctx, id)
	cr.Labels, _ = r.ListRequestLabels(ctx, id)
	cr.Events, _ = r.ListEvents(ctx, id)
	return cr, nil
}

func (r *PostgresRepository) CreateChangeRequest(ctx context.Context, tx pgx.Tx, cr *domain.ChangeRequest) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_change_requests (
			id, request_no, title, description, request_type, status, risk_level,
			target_entity_type, target_entity_id, created_by, assigned_to
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, cr.ID, cr.RequestNo, cr.Title, cr.Description, cr.RequestType, cr.Status, cr.RiskLevel,
		nullableString(cr.TargetEntityType), nullableString(cr.TargetEntityID), cr.CreatedBy, cr.AssignedTo)
	if err != nil {
		return fmt.Errorf("insert permission change request: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateChangeRequest(ctx context.Context, tx pgx.Tx, cr *domain.ChangeRequest) error {
	tag, err := tx.Exec(ctx, `
		UPDATE permission_change_requests
		   SET title=$2, description=$3, request_type=$4, status=$5, risk_level=$6,
		       target_entity_type=$7, target_entity_id=$8, assigned_to=$9,
		       submitted_at=$10, approved_at=$11, approved_by=$12,
		       merged_at=$13, merged_by=$14, rejected_at=$15, rejected_by=$16,
		       rejection_reason=$17, closed_at=$18, closed_by=$19
		 WHERE id=$1
	`, cr.ID, cr.Title, cr.Description, cr.RequestType, cr.Status, cr.RiskLevel,
		nullableString(cr.TargetEntityType), nullableString(cr.TargetEntityID), cr.AssignedTo,
		cr.SubmittedAt, cr.ApprovedAt, cr.ApprovedBy, cr.MergedAt, cr.MergedBy,
		cr.RejectedAt, cr.RejectedBy, nullableString(cr.RejectionReason), cr.ClosedAt, cr.ClosedBy)
	if err != nil {
		return fmt.Errorf("update permission change request: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("permission change request", cr.ID)
	}
	return nil
}

func (r *PostgresRepository) LockChangeRequest(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*domain.ChangeRequest, error) {
	return r.getChangeRequest(ctx, tx, id, true)
}

func (r *PostgresRepository) ListItems(ctx context.Context, requestID uuid.UUID) ([]domain.ChangeItem, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, request_id, item_type, COALESCE(target_table,''), COALESCE(target_id,''),
		       action_type, before_json, after_json, created_at
		FROM permission_change_items
		WHERE request_id=$1
		ORDER BY created_at, id
	`, requestID)
	if err != nil {
		return nil, fmt.Errorf("list change items: %w", err)
	}
	defer rows.Close()
	out := []domain.ChangeItem{}
	for rows.Next() {
		var item domain.ChangeItem
		if err := rows.Scan(&item.ID, &item.RequestID, &item.ItemType, &item.TargetTable, &item.TargetID, &item.ActionType, &item.BeforeJSON, &item.AfterJSON, &item.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan change item: %w", err)
		}
		out = append(out, item)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) AddItem(ctx context.Context, tx pgx.Tx, item *domain.ChangeItem) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_change_items (
			id, request_id, item_type, target_table, target_id, action_type, before_json, after_json
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, item.ID, item.RequestID, item.ItemType, nullableString(item.TargetTable), nullableString(item.TargetID),
		item.ActionType, jsonOrEmpty(item.BeforeJSON), jsonOrEmpty(item.AfterJSON))
	if err != nil {
		return fmt.Errorf("insert change item: %w", err)
	}
	return nil
}

func (r *PostgresRepository) UpdateItem(ctx context.Context, tx pgx.Tx, item *domain.ChangeItem) error {
	tag, err := tx.Exec(ctx, `
		UPDATE permission_change_items
		   SET item_type=$3, target_table=$4, target_id=$5, action_type=$6,
		       before_json=$7, after_json=$8
		 WHERE request_id=$1 AND id=$2
	`, item.RequestID, item.ID, item.ItemType, nullableString(item.TargetTable), nullableString(item.TargetID),
		item.ActionType, jsonOrEmpty(item.BeforeJSON), jsonOrEmpty(item.AfterJSON))
	if err != nil {
		return fmt.Errorf("update change item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("permission change item", item.ID)
	}
	return nil
}

func (r *PostgresRepository) DeleteItem(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, itemID uuid.UUID) error {
	tag, err := tx.Exec(ctx, "DELETE FROM permission_change_items WHERE request_id=$1 AND id=$2", requestID, itemID)
	if err != nil {
		return fmt.Errorf("delete change item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("permission change item", itemID)
	}
	return nil
}

func (r *PostgresRepository) CountItems(ctx context.Context, requestID uuid.UUID) (int, error) {
	var count int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM permission_change_items WHERE request_id=$1", requestID).Scan(&count)
	return count, err
}

func (r *PostgresRepository) FindApprovalSetting(ctx context.Context, requestType string, riskLevel string) (*domain.ApprovalSetting, error) {
	setting, err := r.findApprovalSetting(ctx, `
		module='permission' AND request_type=$1 AND risk_level=$2 AND is_active=true
		ORDER BY is_default DESC, setting_code LIMIT 1`, requestType, riskLevel)
	if err != nil || setting != nil {
		return setting, err
	}
	setting, err = r.findApprovalSetting(ctx, `
		module='permission' AND risk_level=$1 AND is_default=true AND is_active=true
		ORDER BY setting_code LIMIT 1`, riskLevel)
	if err != nil || setting != nil {
		return setting, err
	}
	return r.findApprovalSetting(ctx, `
		module='permission' AND setting_code='PERMISSION_LOW_RISK_DEFAULT' AND is_active=true
		LIMIT 1`)
}

func (r *PostgresRepository) ListApprovalSettings(ctx context.Context) ([]domain.ApprovalSetting, error) {
	rows, err := r.pool.Query(ctx, approvalSettingSelect+" ORDER BY risk_level, setting_code")
	if err != nil {
		return nil, fmt.Errorf("list approval settings: %w", err)
	}
	defer rows.Close()
	out := []domain.ApprovalSetting{}
	for rows.Next() {
		s, err := scanApprovalSetting(rows)
		if err != nil {
			return nil, err
		}
		steps, err := r.listApprovalSettingSteps(ctx, s.ID)
		if err != nil {
			return nil, err
		}
		s.Steps = steps
		out = append(out, *s)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) GetApprovalSetting(ctx context.Context, id uuid.UUID) (*domain.ApprovalSetting, error) {
	s, err := r.findApprovalSetting(ctx, "id=$1", id)
	if err != nil || s == nil {
		return s, err
	}
	s.Steps, err = r.listApprovalSettingSteps(ctx, s.ID)
	return s, err
}

func (r *PostgresRepository) CreateRequestStep(ctx context.Context, tx pgx.Tx, step *domain.ApprovalStep) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_request_approval_steps (
			id, request_id, workflow_setting_id, step_no, step_name, approval_mode,
			status, min_approvals_required, approvals_received, started_at, completed_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
	`, step.ID, step.RequestID, step.WorkflowSettingID, step.StepNo, step.StepName,
		step.ApprovalMode, step.Status, step.MinApprovalsRequired, step.ApprovalsReceived,
		step.StartedAt, step.CompletedAt)
	if err != nil {
		return fmt.Errorf("insert request approval step: %w", err)
	}
	return nil
}

func (r *PostgresRepository) CreateStepApprover(ctx context.Context, tx pgx.Tx, approver *domain.StepApprover) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_request_step_approvers (
			id, request_id, approval_step_id, approver_user_id, approver_role_code,
			approval_status, approval_comment, is_delegated, delegated_from_user_id, approved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)
	`, approver.ID, approver.RequestID, approver.ApprovalStepID, approver.ApproverUserID,
		nullableString(approver.ApproverRoleCode), approver.ApprovalStatus,
		nullableString(approver.ApprovalComment), approver.IsDelegated, approver.DelegatedFromUserID,
		approver.ApprovedAt)
	if err != nil {
		return fmt.Errorf("insert request step approver: %w", err)
	}
	return nil
}

func (r *PostgresRepository) DeleteRequestApprovalState(ctx context.Context, tx pgx.Tx, requestID uuid.UUID) error {
	_, err := tx.Exec(ctx, "DELETE FROM permission_request_approval_steps WHERE request_id=$1", requestID)
	return err
}

func (r *PostgresRepository) ListApprovalSteps(ctx context.Context, requestID uuid.UUID) ([]domain.ApprovalStep, error) {
	rows, err := r.pool.Query(ctx, approvalStepSelect+" WHERE request_id=$1 ORDER BY step_no", requestID)
	if err != nil {
		return nil, fmt.Errorf("list approval steps: %w", err)
	}
	defer rows.Close()
	out := []domain.ApprovalStep{}
	for rows.Next() {
		step, err := scanApprovalStep(rows)
		if err != nil {
			return nil, err
		}
		step.Approvers, err = r.listStepApprovers(ctx, step.ID)
		if err != nil {
			return nil, err
		}
		out = append(out, *step)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) GetApprovalStep(ctx context.Context, requestID uuid.UUID, stepID uuid.UUID) (*domain.ApprovalStep, error) {
	step, err := r.getApprovalStep(ctx, requestID, stepID)
	if err != nil || step == nil {
		return step, err
	}
	step.Approvers, err = r.listStepApprovers(ctx, step.ID)
	return step, err
}

func (r *PostgresRepository) GetCurrentPendingStep(ctx context.Context, requestID uuid.UUID) (*domain.ApprovalStep, error) {
	row := r.pool.QueryRow(ctx, approvalStepSelect+" WHERE request_id=$1 AND status='PENDING' ORDER BY step_no LIMIT 1", requestID)
	step, err := scanApprovalStep(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	step.Approvers, _ = r.listStepApprovers(ctx, step.ID)
	return step, nil
}

func (r *PostgresRepository) UpdateApprovalStep(ctx context.Context, tx pgx.Tx, step *domain.ApprovalStep) error {
	_, err := tx.Exec(ctx, `
		UPDATE permission_request_approval_steps
		   SET status=$3, approvals_received=$4, started_at=$5, completed_at=$6
		 WHERE request_id=$1 AND id=$2
	`, step.RequestID, step.ID, step.Status, step.ApprovalsReceived, step.StartedAt, step.CompletedAt)
	return err
}

func (r *PostgresRepository) UpdateStepApproverDecision(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stepID uuid.UUID, actorID uuid.UUID, roleCode string, status string, comment string) error {
	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE permission_request_step_approvers
		   SET approval_status=$5, approval_comment=$6, approved_at=$7
		 WHERE request_id=$1 AND approval_step_id=$2
		   AND (approver_user_id=$3 OR (approver_role_code <> '' AND approver_role_code=$4))
	`, requestID, stepID, actorID, roleCode, status, nullableString(comment), now)
	if err != nil {
		return err
	}
	if tag.RowsAffected() > 0 {
		return nil
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO permission_request_step_approvers (
			id, request_id, approval_step_id, approver_user_id, approver_role_code,
			approval_status, approval_comment, approved_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, uuid.New(), requestID, stepID, actorID, nullableString(roleCode), status, nullableString(comment), now)
	return err
}

func (r *PostgresRepository) StartNextApprovalStep(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, completedStepNo int) (bool, error) {
	var nextID uuid.UUID
	err := tx.QueryRow(ctx, `
		SELECT id
		FROM permission_request_approval_steps
		WHERE request_id=$1 AND step_no > $2 AND status='NOT_STARTED'
		ORDER BY step_no
		LIMIT 1
	`, requestID, completedStepNo).Scan(&nextID)
	if errors.Is(err, pgx.ErrNoRows) {
		return true, nil
	}
	if err != nil {
		return false, err
	}
	_, err = tx.Exec(ctx, `
		UPDATE permission_request_approval_steps
		   SET status='PENDING', started_at=NOW()
		 WHERE id=$1
	`, nextID)
	return false, err
}

func (r *PostgresRepository) ListComments(ctx context.Context, requestID uuid.UUID) ([]domain.Comment, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.request_id, c.user_id, COALESCE(u.display_name, u.username, ''),
		       c.comment, c.created_at, c.updated_at, c.deleted_at
		FROM permission_request_comments c
		LEFT JOIN iam_users u ON u.id = c.user_id
		WHERE c.request_id=$1 AND c.deleted_at IS NULL
		ORDER BY c.created_at
	`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Comment{}
	for rows.Next() {
		var c domain.Comment
		if err := rows.Scan(&c.ID, &c.RequestID, &c.UserID, &c.UserName, &c.Comment, &c.CreatedAt, &c.UpdatedAt, &c.DeletedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) AddComment(ctx context.Context, tx pgx.Tx, comment *domain.Comment) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_request_comments (id, request_id, user_id, comment)
		VALUES ($1,$2,$3,$4)
	`, comment.ID, comment.RequestID, comment.UserID, comment.Comment)
	return err
}

func (r *PostgresRepository) UpdateComment(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, commentID uuid.UUID, actorID uuid.UUID, text string) error {
	tag, err := tx.Exec(ctx, `
		UPDATE permission_request_comments SET comment=$4
		WHERE request_id=$1 AND id=$2 AND user_id=$3 AND deleted_at IS NULL
	`, requestID, commentID, actorID, text)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("permission request comment", commentID)
	}
	return nil
}

func (r *PostgresRepository) DeleteComment(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, commentID uuid.UUID, actorID uuid.UUID) error {
	tag, err := tx.Exec(ctx, `
		UPDATE permission_request_comments SET deleted_at=NOW()
		WHERE request_id=$1 AND id=$2 AND user_id=$3 AND deleted_at IS NULL
	`, requestID, commentID, actorID)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("permission request comment", commentID)
	}
	return nil
}

func (r *PostgresRepository) ReplaceChecks(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, checks []domain.Check) error {
	if _, err := tx.Exec(ctx, "DELETE FROM permission_request_checks WHERE request_id=$1", requestID); err != nil {
		return err
	}
	for _, c := range checks {
		if c.ID == uuid.Nil {
			c.ID = uuid.New()
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO permission_request_checks (
				id, request_id, check_code, check_name, status, severity, message
			) VALUES ($1,$2,$3,$4,$5,$6,$7)
		`, c.ID, requestID, c.CheckCode, c.CheckName, c.Status, c.Severity, c.Message)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *PostgresRepository) ListChecks(ctx context.Context, requestID uuid.UUID) ([]domain.Check, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, request_id, check_code, check_name, status, severity, message, created_at
		FROM permission_request_checks
		WHERE request_id=$1
		ORDER BY status DESC, severity DESC, check_code
	`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Check{}
	for rows.Next() {
		var c domain.Check
		if err := rows.Scan(&c.ID, &c.RequestID, &c.CheckCode, &c.CheckName, &c.Status, &c.Severity, &c.Message, &c.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) HasBlockingChecks(ctx context.Context, requestID uuid.UUID) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM permission_request_checks
			WHERE request_id=$1 AND status='FAILED' AND severity='BLOCKER'
		)
	`, requestID).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ListLabels(ctx context.Context) ([]domain.Label, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, label_code, label_name, label_type, color, COALESCE(description,''),
		       is_system, is_active, created_at, updated_at
		FROM permission_labels
		WHERE is_active=true
		ORDER BY label_type, label_code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLabels(rows)
}

func (r *PostgresRepository) UpsertLabel(ctx context.Context, tx pgx.Tx, label *domain.Label) error {
	if label.ID == uuid.Nil {
		label.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_labels (
			id, label_code, label_name, label_type, color, description,
			is_system, is_active
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
		ON CONFLICT (label_code) DO UPDATE SET
			label_name=EXCLUDED.label_name,
			label_type=EXCLUDED.label_type,
			color=EXCLUDED.color,
			description=EXCLUDED.description,
			is_system=EXCLUDED.is_system,
			is_active=EXCLUDED.is_active,
			updated_at=NOW()
	`, label.ID, label.LabelCode, label.LabelName, label.LabelType, label.Color,
		nullableString(label.Description), label.IsSystem, label.IsActive)
	return err
}

func (r *PostgresRepository) AddRequestLabel(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, labelCode string, actorID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_change_request_labels (request_id, label_id, added_by)
		SELECT $1, id, $3 FROM permission_labels WHERE label_code=$2 AND is_active=true
		ON CONFLICT (request_id, label_id) DO NOTHING
	`, requestID, labelCode, actorID)
	return err
}

func (r *PostgresRepository) RemoveRequestLabel(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, labelID uuid.UUID) error {
	_, err := tx.Exec(ctx, "DELETE FROM permission_change_request_labels WHERE request_id=$1 AND label_id=$2", requestID, labelID)
	return err
}

func (r *PostgresRepository) ListRequestLabels(ctx context.Context, requestID uuid.UUID) ([]domain.Label, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT l.id, l.label_code, l.label_name, l.label_type, l.color, COALESCE(l.description,''),
		       l.is_system, l.is_active, l.created_at, l.updated_at
		FROM permission_change_request_labels rl
		JOIN permission_labels l ON l.id = rl.label_id
		WHERE rl.request_id=$1
		ORDER BY l.label_type, l.label_code
	`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanLabels(rows)
}

func (r *PostgresRepository) AppendEvent(ctx context.Context, tx pgx.Tx, event domain.WorkflowEvent) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval_workflow_events (
			id, request_id, approval_step_id, actor_user_id, event_type,
			before_json, after_json, comment
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8)
	`, event.ID, event.RequestID, event.ApprovalStepID, event.ActorUserID, event.EventType,
		jsonOrEmpty(event.BeforeJSON), jsonOrEmpty(event.AfterJSON), nullableString(event.Comment))
	return err
}

func (r *PostgresRepository) ListEvents(ctx context.Context, requestID uuid.UUID) ([]domain.WorkflowEvent, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT e.id, e.request_id, e.approval_step_id, e.actor_user_id,
		       COALESCE(u.display_name, u.username, ''), e.event_type,
		       e.before_json, e.after_json, COALESCE(e.comment,''), e.created_at
		FROM approval_workflow_events e
		LEFT JOIN iam_users u ON u.id = e.actor_user_id
		WHERE e.request_id=$1
		ORDER BY e.created_at
	`, requestID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.WorkflowEvent{}
	for rows.Next() {
		var e domain.WorkflowEvent
		if err := rows.Scan(&e.ID, &e.RequestID, &e.ApprovalStepID, &e.ActorUserID, &e.ActorName, &e.EventType, &e.BeforeJSON, &e.AfterJSON, &e.Comment, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) AppendAuditLog(ctx context.Context, tx pgx.Tx, log domain.AuditLog) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO audit_logs (
			id, actor_user_id, action, module, entity_type, entity_id,
			before_json, after_json, ip_address, user_agent, correlation_id
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9::inet,$10,$11)
	`, log.ID, log.ActorUserID, log.Action, log.Module, nullableString(log.EntityType), nullableString(log.EntityID),
		jsonOrEmpty(log.BeforeJSON), jsonOrEmpty(log.AfterJSON), nullableString(log.IPAddress), nullableString(log.UserAgent),
		nullableString(log.CorrelationID))
	return err
}

func (r *PostgresRepository) ListRoles(ctx context.Context) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, roleSelect+" ORDER BY priority_rank DESC, role_code")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Role{}
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *role)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) GetRoleByID(ctx context.Context, id uuid.UUID) (*domain.Role, error) {
	role, err := scanRole(r.pool.QueryRow(ctx, roleSelect+" WHERE id=$1", id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return role, err
}

func (r *PostgresRepository) GetRoleByCode(ctx context.Context, code string) (*domain.Role, error) {
	role, err := scanRole(r.pool.QueryRow(ctx, roleSelect+" WHERE role_code=$1", code))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return role, err
}

func (r *PostgresRepository) GetUserApprovedRoles(ctx context.Context, userID uuid.UUID) ([]domain.Role, error) {
	rows, err := r.pool.Query(ctx, roleSelect+`
		JOIN permission_user_role_assignments ura ON ura.role_id = pr.id
		WHERE ura.user_id=$1 AND ura.status='APPROVED'
		  AND (ura.effective_from IS NULL OR ura.effective_from <= NOW())
		  AND (ura.effective_to IS NULL OR ura.effective_to > NOW())
		ORDER BY pr.priority_rank DESC
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.Role{}
	for rows.Next() {
		role, err := scanRole(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *role)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) ListUsers(ctx context.Context, search string, page int, limit int) ([]domain.UserSummary, int, error) {
	where := "u.deleted_at IS NULL"
	args := []any{}
	if search != "" {
		where += " AND (u.username ILIKE $1 OR u.display_name ILIKE $1 OR u.email ILIKE $1)"
		args = append(args, "%"+search+"%")
	}
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM iam_users u WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	l, o := pageLimit(page, limit)
	args = append(args, l, o)
	limitIdx := len(args) - 1
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT u.id, u.username, u.display_name, COALESCE(u.email,''), u.is_active,
		       (u.locked_until IS NOT NULL AND u.locked_until > NOW()) AS is_locked,
		       u.locked_until, u.force_password_change, u.last_login_at,
		       COALESCE(array_agg(DISTINCT g.name) FILTER (WHERE g.name IS NOT NULL), '{}'),
		       COALESCE(array_agg(DISTINCT pr.role_code) FILTER (WHERE pr.role_code IS NOT NULL), '{}'),
		       u.created_at, u.updated_at
		FROM iam_users u
		LEFT JOIN permissions_accounts_groups ag ON ag.user_id = u.id
		LEFT JOIN permissions_groups g ON g.id = ag.group_id AND g.is_active = true AND g.deleted_at IS NULL
		LEFT JOIN permission_user_role_assignments ura ON ura.user_id = u.id AND ura.status='APPROVED'
		LEFT JOIN permission_roles pr ON pr.id = ura.role_id AND pr.is_active = true
		WHERE %s
		GROUP BY u.id
		ORDER BY u.username
		LIMIT $%d OFFSET $%d
	`, where, limitIdx, limitIdx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.UserSummary{}
	for rows.Next() {
		var u domain.UserSummary
		if err := rows.Scan(&u.ID, &u.Username, &u.DisplayName, &u.Email, &u.IsActive, &u.IsLocked, &u.LockedUntil, &u.ForcePasswordChange, &u.LastLoginAt, &u.Groups, &u.Roles, &u.CreatedAt, &u.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, u)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserSummary, error) {
	users, _, err := r.ListUsers(ctx, "", 1, 500)
	if err != nil {
		return nil, err
	}
	for _, user := range users {
		if user.ID == id {
			return &user, nil
		}
	}
	return nil, nil
}

func (r *PostgresRepository) ListGroups(ctx context.Context, search string, page int, limit int) ([]domain.GroupSummary, int, error) {
	where := "g.deleted_at IS NULL"
	args := []any{}
	if search != "" {
		where += " AND g.name ILIKE $1"
		args = append(args, "%"+search+"%")
	}
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM permissions_groups g WHERE "+where, args...).Scan(&total); err != nil {
		return nil, 0, err
	}
	l, o := pageLimit(page, limit)
	args = append(args, l, o)
	limitIdx := len(args) - 1
	rows, err := r.pool.Query(ctx, fmt.Sprintf(`
		SELECT g.id, g.name, COALESCE(g.description,''), g.is_active,
		       COUNT(ag.user_id), g.created_at, g.updated_at
		FROM permissions_groups g
		LEFT JOIN permissions_accounts_groups ag ON ag.group_id = g.id
		WHERE %s
		GROUP BY g.id
		ORDER BY g.name
		LIMIT $%d OFFSET $%d
	`, where, limitIdx, limitIdx+1), args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.GroupSummary{}
	for rows.Next() {
		var g domain.GroupSummary
		if err := rows.Scan(&g.ID, &g.Name, &g.Description, &g.IsActive, &g.MembersCount, &g.CreatedAt, &g.UpdatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, g)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) GetGroup(ctx context.Context, id uuid.UUID) (*domain.GroupSummary, error) {
	rows, _, err := r.ListGroups(ctx, "", 1, 500)
	if err != nil {
		return nil, err
	}
	for _, group := range rows {
		if group.ID == id {
			return &group, nil
		}
	}
	return nil, nil
}

func (r *PostgresRepository) ListFunctionDefinitions(ctx context.Context) ([]domain.FunctionDefinition, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT code, module, COALESCE(screen,''), action, name, COALESCE(description,''), deprecated_at
		FROM permission_function_definitions
		ORDER BY module, screen, action, code
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.FunctionDefinition{}
	for rows.Next() {
		var d domain.FunctionDefinition
		if err := rows.Scan(&d.Code, &d.Module, &d.Screen, &d.Action, &d.Name, &d.Description, &d.DeprecatedAt); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) FunctionDefinitionExists(ctx context.Context, code string) (bool, error) {
	var exists bool
	err := r.pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM permission_function_definitions WHERE code=$1 AND deprecated_at IS NULL
			UNION ALL
			SELECT 1 FROM permissions_function_definitions WHERE code=$1 AND deprecated_at IS NULL
			LIMIT 1
		)
	`, code).Scan(&exists)
	return exists, err
}

func (r *PostgresRepository) ListFunctionRights(ctx context.Context) ([]domain.FunctionRight, error) {
	rows, err := r.pool.Query(ctx, functionRightSelect+" ORDER BY permission_code, subject_type, subject_name")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFunctionRights(rows)
}

func (r *PostgresRepository) UpsertFunctionRight(ctx context.Context, tx pgx.Tx, right domain.FunctionRight) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_function_rights (
			id, subject_type, subject_id, permission_code, can_view, can_search, can_add,
			can_edit, can_delete, can_approve, can_revoke_approval, can_export,
			can_configure, approved_at, approved_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,NOW(),$14)
		ON CONFLICT (subject_type, subject_id, permission_code) DO UPDATE SET
			can_view=EXCLUDED.can_view, can_search=EXCLUDED.can_search,
			can_add=EXCLUDED.can_add, can_edit=EXCLUDED.can_edit,
			can_delete=EXCLUDED.can_delete, can_approve=EXCLUDED.can_approve,
			can_revoke_approval=EXCLUDED.can_revoke_approval,
			can_export=EXCLUDED.can_export, can_configure=EXCLUDED.can_configure,
			approved_at=EXCLUDED.approved_at, approved_by=EXCLUDED.approved_by,
			updated_at=NOW()
	`, newOrExisting(right.ID), right.SubjectType, right.SubjectID, right.PermissionCode, right.CanView,
		right.CanSearch, right.CanAdd, right.CanEdit, right.CanDelete, right.CanApprove,
		right.CanRevokeApproval, right.CanExport, right.CanConfigure, right.ApprovedBy)
	return err
}

func (r *PostgresRepository) ListDataRights(ctx context.Context) ([]domain.DataRight, error) {
	rows, err := r.pool.Query(ctx, dataRightSelect+" ORDER BY subject_type, subject_name, access_level")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDataRights(rows)
}

func (r *PostgresRepository) UpsertDataRight(ctx context.Context, tx pgx.Tx, right domain.DataRight) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_data_rights (
			id, subject_type, subject_id, fund_id, contract_id, portfolio_id,
			access_level, approved_at, approved_by
		) VALUES ($1,$2,$3,$4,$5,$6,$7,NOW(),$8)
		ON CONFLICT (
			subject_type, subject_id,
			COALESCE(fund_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(contract_id, '00000000-0000-0000-0000-000000000000'::uuid),
			COALESCE(portfolio_id, '00000000-0000-0000-0000-000000000000'::uuid),
			access_level
		) DO UPDATE SET
			approved_at=EXCLUDED.approved_at, approved_by=EXCLUDED.approved_by, updated_at=NOW()
	`, newOrExisting(right.ID), right.SubjectType, right.SubjectID, right.FundID, right.ContractID,
		right.PortfolioID, right.AccessLevel, right.ApprovedBy)
	return err
}

func (r *PostgresRepository) InsertRoleAssignment(ctx context.Context, tx pgx.Tx, userID uuid.UUID, roleID uuid.UUID, actorID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permission_user_role_assignments (
			user_id, role_id, status, assigned_by, approved_by, approved_at, effective_from
		) VALUES ($1,$2,'APPROVED',$3,$3,NOW(),NOW())
		ON CONFLICT DO NOTHING
	`, userID, roleID, actorID)
	return err
}

func (r *PostgresRepository) InsertGroupMembership(ctx context.Context, tx pgx.Tx, userID uuid.UUID, groupID uuid.UUID, actorID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO permissions_accounts_groups (user_id, group_id, assigned_by)
		VALUES ($1,$2,$3)
		ON CONFLICT (user_id, group_id) DO NOTHING
	`, userID, groupID, actorID)
	return err
}

func (r *PostgresRepository) DeleteGroupMembership(ctx context.Context, tx pgx.Tx, userID uuid.UUID, groupID uuid.UUID) error {
	_, err := tx.Exec(ctx, "DELETE FROM permissions_accounts_groups WHERE user_id=$1 AND group_id=$2", userID, groupID)
	return err
}

func (r *PostgresRepository) EffectivePermissions(ctx context.Context, userID uuid.UUID) (*domain.EffectivePermissions, error) {
	e := &domain.EffectivePermissions{UserID: userID}
	e.Roles, _ = r.GetUserApprovedRoles(ctx, userID)
	if groups, _, err := r.ListGroups(ctx, "", 1, 500); err == nil {
		e.Groups = groups
	}
	var err error
	e.DirectFunctions, err = r.functionRightsForUser(ctx, userID, "USER")
	if err != nil {
		return nil, err
	}
	e.GroupFunctions, err = r.functionRightsForUser(ctx, userID, "GROUP")
	if err != nil {
		return nil, err
	}
	e.RoleFunctions, err = r.functionRightsForUser(ctx, userID, "ROLE")
	if err != nil {
		return nil, err
	}
	e.DirectDataRights, _ = r.dataRightsForUser(ctx, userID, "USER")
	e.GroupDataRights, _ = r.dataRightsForUser(ctx, userID, "GROUP")
	e.RoleDataRights, _ = r.dataRightsForUser(ctx, userID, "ROLE")
	e.FinalFunctionCodes, _ = r.effectiveFunctionCodes(ctx, userID)
	e.FinalContractScopes, _ = r.effectiveContractScopes(ctx, userID)
	return e, nil
}

func (r *PostgresRepository) ListAuditLogs(ctx context.Context, page int, limit int) ([]domain.AuditLog, int, error) {
	var total int
	if err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM audit_logs").Scan(&total); err != nil {
		return nil, 0, err
	}
	l, o := pageLimit(page, limit)
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.actor_user_id, COALESCE(u.display_name, u.username, ''),
		       a.action, a.module, COALESCE(a.entity_type,''), COALESCE(a.entity_id,''),
		       a.before_json, a.after_json, COALESCE(host(a.ip_address),''),
		       COALESCE(a.user_agent,''), COALESCE(a.correlation_id,''), a.created_at
		FROM audit_logs a
		LEFT JOIN iam_users u ON u.id = a.actor_user_id
		ORDER BY a.created_at DESC
		LIMIT $1 OFFSET $2
	`, l, o)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()
	out := []domain.AuditLog{}
	for rows.Next() {
		var a domain.AuditLog
		if err := rows.Scan(&a.ID, &a.ActorUserID, &a.ActorName, &a.Action, &a.Module, &a.EntityType, &a.EntityID, &a.BeforeJSON, &a.AfterJSON, &a.IPAddress, &a.UserAgent, &a.CorrelationID, &a.CreatedAt); err != nil {
			return nil, 0, err
		}
		out = append(out, a)
	}
	return out, total, rows.Err()
}

func (r *PostgresRepository) ListNotificationSettings(ctx context.Context) ([]domain.NotificationSetting, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, event_code, channel, enabled, target_scope, template_subject,
		       template_body, created_at, updated_at
		FROM notification_settings
		ORDER BY event_code, channel
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.NotificationSetting{}
	for rows.Next() {
		var n domain.NotificationSetting
		if err := rows.Scan(&n.ID, &n.EventCode, &n.Channel, &n.Enabled, &n.TargetScope, &n.TemplateSubject, &n.TemplateBody, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpdateNotificationSetting(ctx context.Context, tx pgx.Tx, setting domain.NotificationSetting) error {
	tag, err := tx.Exec(ctx, `
		UPDATE notification_settings
		   SET enabled=$2, target_scope=$3, template_subject=$4, template_body=$5
		 WHERE id=$1
	`, setting.ID, setting.Enabled, setting.TargetScope, setting.TemplateSubject, setting.TemplateBody)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return domain.NotFound("notification setting", setting.ID)
	}
	return nil
}

const changeRequestSelect = `
	SELECT cr.id, cr.request_no, cr.title, COALESCE(cr.description,''), cr.request_type,
	       cr.status, cr.risk_level, COALESCE(cr.target_entity_type,''),
	       COALESCE(cr.target_entity_id,''), cr.created_by,
	       COALESCE(u.display_name, u.username, ''), cr.assigned_to,
	       cr.submitted_at, cr.approved_at, cr.approved_by, cr.merged_at, cr.merged_by,
	       cr.rejected_at, cr.rejected_by, COALESCE(cr.rejection_reason,''),
	       cr.closed_at, cr.closed_by, cr.created_at, cr.updated_at
	FROM permission_change_requests cr
	LEFT JOIN iam_users u ON u.id = cr.created_by`

func (r *PostgresRepository) getChangeRequest(ctx context.Context, q queryer, id uuid.UUID, forUpdate bool) (*domain.ChangeRequest, error) {
	sql := changeRequestSelect + " WHERE cr.id=$1"
	if forUpdate {
		sql += " FOR UPDATE OF cr"
	}
	cr, err := scanChangeRequest(q.QueryRow(ctx, sql, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return cr, err
}

func scanChangeRequest(row pgx.Row) (*domain.ChangeRequest, error) {
	var cr domain.ChangeRequest
	err := row.Scan(
		&cr.ID, &cr.RequestNo, &cr.Title, &cr.Description, &cr.RequestType,
		&cr.Status, &cr.RiskLevel, &cr.TargetEntityType, &cr.TargetEntityID,
		&cr.CreatedBy, &cr.CreatedByName, &cr.AssignedTo, &cr.SubmittedAt,
		&cr.ApprovedAt, &cr.ApprovedBy, &cr.MergedAt, &cr.MergedBy,
		&cr.RejectedAt, &cr.RejectedBy, &cr.RejectionReason, &cr.ClosedAt,
		&cr.ClosedBy, &cr.CreatedAt, &cr.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &cr, nil
}

const approvalSettingSelect = `
	SELECT id, setting_code, setting_name, module, request_type, risk_level,
	       is_active, is_default, sequential_approval, allow_creator_approval,
	       failed_check_blocks_submit, failed_check_blocks_merge,
	       COALESCE(description,''), created_at, updated_at
	FROM approval_workflow_settings`

func (r *PostgresRepository) findApprovalSetting(ctx context.Context, where string, args ...any) (*domain.ApprovalSetting, error) {
	row := r.pool.QueryRow(ctx, approvalSettingSelect+" WHERE "+where, args...)
	s, err := scanApprovalSetting(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s.Steps, err = r.listApprovalSettingSteps(ctx, s.ID)
	return s, err
}

func scanApprovalSetting(row pgx.Row) (*domain.ApprovalSetting, error) {
	var s domain.ApprovalSetting
	err := row.Scan(&s.ID, &s.SettingCode, &s.SettingName, &s.Module, &s.RequestType, &s.RiskLevel,
		&s.IsActive, &s.IsDefault, &s.SequentialApproval, &s.AllowCreatorApproval,
		&s.FailedCheckBlocksSubmit, &s.FailedCheckBlocksMerge, &s.Description,
		&s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) listApprovalSettingSteps(ctx context.Context, settingID uuid.UUID) ([]domain.ApprovalSettingStep, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, workflow_setting_id, step_no, step_name, COALESCE(step_description,''),
		       approval_mode, approver_type, COALESCE(required_role_code,''),
		       required_group_id, required_user_id, min_approvals_required,
		       allow_delegation, is_required, created_at, updated_at
		FROM approval_workflow_steps
		WHERE workflow_setting_id=$1
		ORDER BY step_no
	`, settingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.ApprovalSettingStep{}
	for rows.Next() {
		var s domain.ApprovalSettingStep
		if err := rows.Scan(&s.ID, &s.WorkflowSettingID, &s.StepNo, &s.StepName, &s.StepDescription,
			&s.ApprovalMode, &s.ApproverType, &s.RequiredRoleCode, &s.RequiredGroupID,
			&s.RequiredUserID, &s.MinApprovalsRequired, &s.AllowDelegation, &s.IsRequired,
			&s.CreatedAt, &s.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

const approvalStepSelect = `
	SELECT id, request_id, workflow_setting_id, step_no, step_name, approval_mode,
	       status, min_approvals_required, approvals_received, started_at,
	       completed_at, created_at, updated_at
	FROM permission_request_approval_steps`

func (r *PostgresRepository) getApprovalStep(ctx context.Context, requestID uuid.UUID, stepID uuid.UUID) (*domain.ApprovalStep, error) {
	step, err := scanApprovalStep(r.pool.QueryRow(ctx, approvalStepSelect+" WHERE request_id=$1 AND id=$2", requestID, stepID))
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return step, err
}

func scanApprovalStep(row pgx.Row) (*domain.ApprovalStep, error) {
	var s domain.ApprovalStep
	err := row.Scan(&s.ID, &s.RequestID, &s.WorkflowSettingID, &s.StepNo, &s.StepName,
		&s.ApprovalMode, &s.Status, &s.MinApprovalsRequired, &s.ApprovalsReceived,
		&s.StartedAt, &s.CompletedAt, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

func (r *PostgresRepository) listStepApprovers(ctx context.Context, stepID uuid.UUID) ([]domain.StepApprover, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT a.id, a.request_id, a.approval_step_id, a.approver_user_id,
		       COALESCE(u.display_name, u.username, ''), COALESCE(a.approver_role_code,''),
		       a.approval_status, COALESCE(a.approval_comment,''), a.is_delegated,
		       a.delegated_from_user_id, a.approved_at, a.created_at
		FROM permission_request_step_approvers a
		LEFT JOIN iam_users u ON u.id = a.approver_user_id
		WHERE a.approval_step_id=$1
		ORDER BY a.created_at
	`, stepID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []domain.StepApprover{}
	for rows.Next() {
		var a domain.StepApprover
		if err := rows.Scan(&a.ID, &a.RequestID, &a.ApprovalStepID, &a.ApproverUserID,
			&a.ApproverDisplayName, &a.ApproverRoleCode, &a.ApprovalStatus,
			&a.ApprovalComment, &a.IsDelegated, &a.DelegatedFromUserID,
			&a.ApprovedAt, &a.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

const roleSelect = `
	SELECT pr.id, pr.role_code, pr.role_name, COALESCE(pr.department,''),
	       COALESCE(pr.role_category,''), pr.priority_rank, pr.assignment_scope,
	       pr.can_request_role_assignment, pr.can_approve_role_assignment,
	       pr.is_high_risk, pr.is_active, COALESCE(pr.description,''),
	       pr.created_at, pr.updated_at
	FROM permission_roles pr`

func scanRole(row pgx.Row) (*domain.Role, error) {
	var role domain.Role
	err := row.Scan(&role.ID, &role.RoleCode, &role.RoleName, &role.Department, &role.RoleCategory,
		&role.PriorityRank, &role.AssignmentScope, &role.CanRequestRoleAssignment,
		&role.CanApproveRoleAssignment, &role.IsHighRisk, &role.IsActive,
		&role.Description, &role.CreatedAt, &role.UpdatedAt)
	if err != nil {
		return nil, err
	}
	return &role, nil
}

const functionRightSelect = `
	SELECT fr.id, fr.subject_type, fr.subject_id,
	       COALESCE(u.display_name, g.name, pr.role_name, fr.subject_id::text) AS subject_name,
	       fr.permission_code, fr.can_view, fr.can_search, fr.can_add, fr.can_edit,
	       fr.can_delete, fr.can_approve, fr.can_revoke_approval, fr.can_export,
	       fr.can_configure, fr.created_at, fr.updated_at, fr.approved_at, fr.approved_by
	FROM permission_function_rights fr
	LEFT JOIN iam_users u ON fr.subject_type='USER' AND u.id=fr.subject_id
	LEFT JOIN permissions_groups g ON fr.subject_type='GROUP' AND g.id=fr.subject_id
	LEFT JOIN permission_roles pr ON fr.subject_type='ROLE' AND pr.id=fr.subject_id`

func scanFunctionRights(rows pgx.Rows) ([]domain.FunctionRight, error) {
	out := []domain.FunctionRight{}
	for rows.Next() {
		var fr domain.FunctionRight
		if err := rows.Scan(&fr.ID, &fr.SubjectType, &fr.SubjectID, &fr.SubjectName, &fr.PermissionCode,
			&fr.CanView, &fr.CanSearch, &fr.CanAdd, &fr.CanEdit, &fr.CanDelete, &fr.CanApprove,
			&fr.CanRevokeApproval, &fr.CanExport, &fr.CanConfigure, &fr.CreatedAt, &fr.UpdatedAt,
			&fr.ApprovedAt, &fr.ApprovedBy); err != nil {
			return nil, err
		}
		out = append(out, fr)
	}
	return out, rows.Err()
}

const dataRightSelect = `
	SELECT dr.id, dr.subject_type, dr.subject_id,
	       COALESCE(u.display_name, g.name, pr.role_name, dr.subject_id::text) AS subject_name,
	       dr.fund_id, dr.contract_id, dr.portfolio_id, dr.access_level,
	       dr.created_at, dr.updated_at, dr.approved_at, dr.approved_by
	FROM permission_data_rights dr
	LEFT JOIN iam_users u ON dr.subject_type='USER' AND u.id=dr.subject_id
	LEFT JOIN permissions_groups g ON dr.subject_type='GROUP' AND g.id=dr.subject_id
	LEFT JOIN permission_roles pr ON dr.subject_type='ROLE' AND pr.id=dr.subject_id`

func scanDataRights(rows pgx.Rows) ([]domain.DataRight, error) {
	out := []domain.DataRight{}
	for rows.Next() {
		var dr domain.DataRight
		if err := rows.Scan(&dr.ID, &dr.SubjectType, &dr.SubjectID, &dr.SubjectName, &dr.FundID,
			&dr.ContractID, &dr.PortfolioID, &dr.AccessLevel, &dr.CreatedAt, &dr.UpdatedAt,
			&dr.ApprovedAt, &dr.ApprovedBy); err != nil {
			return nil, err
		}
		out = append(out, dr)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) functionRightsForUser(ctx context.Context, userID uuid.UUID, subjectType string) ([]domain.FunctionRight, error) {
	var where string
	var args []any
	switch subjectType {
	case "USER":
		where = " WHERE fr.subject_type='USER' AND fr.subject_id=$1"
		args = []any{userID}
	case "GROUP":
		where = ` WHERE fr.subject_type='GROUP' AND EXISTS (
			SELECT 1 FROM permissions_accounts_groups ag
			WHERE ag.group_id=fr.subject_id AND ag.user_id=$1
		)`
		args = []any{userID}
	case "ROLE":
		where = ` WHERE fr.subject_type='ROLE' AND EXISTS (
			SELECT 1 FROM permission_user_role_assignments ura
			WHERE ura.role_id=fr.subject_id AND ura.user_id=$1 AND ura.status='APPROVED'
		)`
		args = []any{userID}
	}
	rows, err := r.pool.Query(ctx, functionRightSelect+where+" ORDER BY fr.permission_code", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanFunctionRights(rows)
}

func (r *PostgresRepository) dataRightsForUser(ctx context.Context, userID uuid.UUID, subjectType string) ([]domain.DataRight, error) {
	var where string
	var args []any
	switch subjectType {
	case "USER":
		where = " WHERE dr.subject_type='USER' AND dr.subject_id=$1"
		args = []any{userID}
	case "GROUP":
		where = ` WHERE dr.subject_type='GROUP' AND EXISTS (
			SELECT 1 FROM permissions_accounts_groups ag
			WHERE ag.group_id=dr.subject_id AND ag.user_id=$1
		)`
		args = []any{userID}
	case "ROLE":
		where = ` WHERE dr.subject_type='ROLE' AND EXISTS (
			SELECT 1 FROM permission_user_role_assignments ura
			WHERE ura.role_id=dr.subject_id AND ura.user_id=$1 AND ura.status='APPROVED'
		)`
		args = []any{userID}
	}
	rows, err := r.pool.Query(ctx, dataRightSelect+where+" ORDER BY dr.access_level", args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanDataRights(rows)
}

func (r *PostgresRepository) effectiveFunctionCodes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT permission_code
		FROM (
			SELECT fr.permission_code
			FROM permission_function_rights fr
			WHERE (fr.can_view OR fr.can_search OR fr.can_add OR fr.can_edit OR fr.can_delete
			       OR fr.can_approve OR fr.can_revoke_approval OR fr.can_export OR fr.can_configure)
			  AND (
			    (fr.subject_type='USER' AND fr.subject_id=$1)
			    OR (fr.subject_type='GROUP' AND EXISTS (
			        SELECT 1 FROM permissions_accounts_groups ag
			        WHERE ag.group_id=fr.subject_id AND ag.user_id=$1
			    ))
			    OR (fr.subject_type='ROLE' AND EXISTS (
			        SELECT 1 FROM permission_user_role_assignments ura
			        WHERE ura.role_id=fr.subject_id AND ura.user_id=$1 AND ura.status='APPROVED'
			    ))
			  )
			UNION
			SELECT legacy.permission_code
			FROM permissions_function_rights legacy
			JOIN permissions_accounts_groups ag ON ag.group_id = legacy.group_id
			WHERE ag.user_id=$1 AND legacy.is_granted=true
		) p
		ORDER BY permission_code
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		out = append(out, code)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) effectiveContractScopes(ctx context.Context, userID uuid.UUID) ([]string, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT DISTINCT scope_id
		FROM (
			SELECT COALESCE(dr.contract_id, dr.fund_id, dr.portfolio_id)::text AS scope_id
			FROM permission_data_rights dr
			WHERE (
			    (dr.subject_type='USER' AND dr.subject_id=$1)
			    OR (dr.subject_type='GROUP' AND EXISTS (
			        SELECT 1 FROM permissions_accounts_groups ag
			        WHERE ag.group_id=dr.subject_id AND ag.user_id=$1
			    ))
			    OR (dr.subject_type='ROLE' AND EXISTS (
			        SELECT 1 FROM permission_user_role_assignments ura
			        WHERE ura.role_id=dr.subject_id AND ura.user_id=$1 AND ura.status='APPROVED'
			    ))
			)
			UNION
			SELECT contract_id
			FROM permissions_data_rights
			WHERE user_id=$1 AND is_granted=true
		) p
		WHERE scope_id IS NOT NULL
		ORDER BY scope_id
	`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []string{}
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, rows.Err()
}

func scanLabels(rows pgx.Rows) ([]domain.Label, error) {
	out := []domain.Label{}
	for rows.Next() {
		var l domain.Label
		if err := rows.Scan(&l.ID, &l.LabelCode, &l.LabelName, &l.LabelType, &l.Color,
			&l.Description, &l.IsSystem, &l.IsActive, &l.CreatedAt, &l.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

type queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

func pageLimit(page int, limit int) (int, int) {
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	return limit, (page - 1) * limit
}

func nullableString(s string) any {
	if strings.TrimSpace(s) == "" {
		return nil
	}
	return s
}

func jsonOrEmpty(raw json.RawMessage) any {
	if len(raw) == 0 {
		return []byte("{}")
	}
	return raw
}

func newOrExisting(id uuid.UUID) uuid.UUID {
	if id == uuid.Nil {
		return uuid.New()
	}
	return id
}

var _ domain.Repository = (*PostgresRepository)(nil)
