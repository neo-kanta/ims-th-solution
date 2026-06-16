package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

const taskSelect = `
	SELECT t.id, t.approval_request_id, t.stage_number, t.assigned_user_id, t.assigned_group_id,
	       t.assigned_team_id, t.delegated_from_user_id, t.status, t.acted_by, t.acted_at,
	       t.action_comment, t.is_delegated_action, t.due_at, t.created_at, t.updated_at,
	       COALESCE(u.display_name, u.username, '')
	FROM approval__tasks t
	LEFT JOIN iam_users u ON u.id = t.assigned_user_id`

func scanTask(row pgx.Row) (*entity.ApprovalTask, error) {
	var t entity.ApprovalTask
	var status string
	err := row.Scan(&t.ID, &t.ApprovalRequestID, &t.StageNumber, &t.AssignedUserID, &t.AssignedGroupID,
		&t.AssignedTeamID, &t.DelegatedFromUserID, &status, &t.ActedBy, &t.ActedAt,
		&t.ActionComment, &t.IsDelegatedAction, &t.DueAt, &t.CreatedAt, &t.UpdatedAt, &t.AssignedUserName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval task: %w", err)
	}
	t.Status = vo.TaskStatus(status)
	return &t, nil
}

// CreateTask inserts an approval task.
func (r *PostgresRepository) CreateTask(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTask) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__tasks
			(id, approval_request_id, stage_number, assigned_user_id, assigned_group_id, assigned_team_id,
			 delegated_from_user_id, status, acted_by, acted_at, action_comment, is_delegated_action, due_at)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13)`,
		t.ID, t.ApprovalRequestID, t.StageNumber, t.AssignedUserID, t.AssignedGroupID, t.AssignedTeamID,
		t.DelegatedFromUserID, string(t.Status), t.ActedBy, t.ActedAt, t.ActionComment, t.IsDelegatedAction, t.DueAt)
	if err != nil {
		return fmt.Errorf("inserting approval task: %w", err)
	}
	return nil
}

// GetTask returns a task by id or nil.
func (r *PostgresRepository) GetTask(ctx context.Context, id uuid.UUID) (*entity.ApprovalTask, error) {
	return scanTask(r.pool.QueryRow(ctx, taskSelect+" WHERE t.id = $1", id))
}

// GetTaskForUpdate locks and returns a task inside the transaction.
func (r *PostgresRepository) GetTaskForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.ApprovalTask, error) {
	var locked uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM approval__tasks WHERE id = $1 FOR UPDATE`, id).Scan(&locked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("locking approval task: %w", err)
	}
	return scanTask(tx.QueryRow(ctx, taskSelect+" WHERE t.id = $1", id))
}

// UpdateTask updates a task's mutable action fields.
func (r *PostgresRepository) UpdateTask(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTask) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__tasks
		SET status=$2, acted_by=$3, acted_at=$4, action_comment=$5, is_delegated_action=$6, delegated_from_user_id=$7
		WHERE id=$1`,
		t.ID, string(t.Status), t.ActedBy, t.ActedAt, t.ActionComment, t.IsDelegatedAction, t.DelegatedFromUserID)
	if err != nil {
		return fmt.Errorf("updating approval task: %w", err)
	}
	return nil
}

// ListTasksByRequest returns all tasks for a request ordered by stage/created.
func (r *PostgresRepository) ListTasksByRequest(ctx context.Context, requestID uuid.UUID) ([]*entity.ApprovalTask, error) {
	return r.queryTasks(ctx, taskSelect+" WHERE t.approval_request_id = $1 ORDER BY t.stage_number, t.created_at", requestID)
}

// ListPendingByStage returns pending tasks for a request stage (FOR UPDATE).
func (r *PostgresRepository) ListPendingByStage(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stage int) ([]*entity.ApprovalTask, error) {
	rows, err := tx.Query(ctx, taskSelect+`
		WHERE t.approval_request_id = $1 AND t.stage_number = $2 AND t.status = 'PENDING'
		ORDER BY t.created_at`, requestID, stage)
	if err != nil {
		return nil, fmt.Errorf("listing pending stage tasks: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// CountApprovedInStage counts APPROVED tasks in a request stage.
func (r *PostgresRepository) CountApprovedInStage(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stage int) (int, error) {
	var n int
	err := tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM approval__tasks
		WHERE approval_request_id = $1 AND stage_number = $2 AND status = 'APPROVED'`, requestID, stage).Scan(&n)
	if err != nil {
		return 0, fmt.Errorf("counting approved stage tasks: %w", err)
	}
	return n, nil
}

// CancelPendingByRequest cancels all pending tasks for a request.
func (r *PostgresRepository) CancelPendingByRequest(ctx context.Context, tx pgx.Tx, requestID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__tasks SET status='CANCELLED'
		WHERE approval_request_id = $1 AND status = 'PENDING'`, requestID)
	if err != nil {
		return fmt.Errorf("cancelling pending tasks: %w", err)
	}
	return nil
}

// SkipOtherPendingInStage marks all OTHER pending tasks in a stage as SKIPPED.
func (r *PostgresRepository) SkipOtherPendingInStage(ctx context.Context, tx pgx.Tx, requestID uuid.UUID, stage int, exceptTaskID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__tasks SET status='SKIPPED'
		WHERE approval_request_id = $1 AND stage_number = $2 AND status = 'PENDING' AND id <> $3`,
		requestID, stage, exceptTaskID)
	if err != nil {
		return fmt.Errorf("skipping other stage tasks: %w", err)
	}
	return nil
}

// Inbox returns approver work items joined with their requests.
func (r *PostgresRepository) Inbox(ctx context.Context, f domain.InboxFilter) ([]*domain.InboxItem, int, error) {
	status := f.Status
	if status == "" {
		status = vo.TaskStatusPending
	}
	limit := f.Limit
	if limit <= 0 || limit > 200 {
		limit = 50
	}
	page := f.Page
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	var total int
	if err := r.pool.QueryRow(ctx, `
		SELECT COUNT(*) FROM approval__tasks t
		WHERE t.assigned_user_id = $1 AND t.status = $2`, f.UserID, string(status)).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting inbox: %w", err)
	}

	rows, err := r.pool.Query(ctx, `
		SELECT
			t.id, t.approval_request_id, t.stage_number, t.assigned_user_id, t.assigned_group_id,
			t.assigned_team_id, t.delegated_from_user_id, t.status, t.acted_by, t.acted_at,
			t.action_comment, t.is_delegated_action, t.due_at, t.created_at, t.updated_at,
			r.id, r.request_number, r.process_type, r.subject_type, r.subject_id, r.subject_title,
			r.subject_reference, r.contract_id, r.portfolio_id, r.submitter_id, r.submitted_at,
			r.current_stage_number, r.status, r.created_at,
			COALESCE(su.display_name, su.username, '')
		FROM approval__tasks t
		JOIN approval__requests r ON r.id = t.approval_request_id
		LEFT JOIN iam_users su ON su.id = r.submitter_id
		WHERE t.assigned_user_id = $1 AND t.status = $2
		ORDER BY r.created_at DESC
		LIMIT $3 OFFSET $4`, f.UserID, string(status), limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("listing inbox: %w", err)
	}
	defer rows.Close()

	var out []*domain.InboxItem
	for rows.Next() {
		var t entity.ApprovalTask
		var req entity.ApprovalRequest
		var taskStatus, processType, subjectType, reqStatus string
		if err := rows.Scan(
			&t.ID, &t.ApprovalRequestID, &t.StageNumber, &t.AssignedUserID, &t.AssignedGroupID,
			&t.AssignedTeamID, &t.DelegatedFromUserID, &taskStatus, &t.ActedBy, &t.ActedAt,
			&t.ActionComment, &t.IsDelegatedAction, &t.DueAt, &t.CreatedAt, &t.UpdatedAt,
			&req.ID, &req.RequestNumber, &processType, &subjectType, &req.SubjectID, &req.SubjectTitle,
			&req.SubjectReference, &req.ContractID, &req.PortfolioID, &req.SubmitterID, &req.SubmittedAt,
			&req.CurrentStageNumber, &reqStatus, &req.CreatedAt, &req.SubmitterName,
		); err != nil {
			return nil, 0, fmt.Errorf("scanning inbox item: %w", err)
		}
		t.Status = vo.TaskStatus(taskStatus)
		req.ProcessType = vo.ProcessType(processType)
		req.SubjectType = vo.SubjectType(subjectType)
		req.Status = vo.RequestStatus(reqStatus)
		out = append(out, &domain.InboxItem{Task: t, Request: req})
	}
	return out, total, rows.Err()
}

// FindPendingTaskForActor returns the PENDING task assigned to actorID on a
// specific request, or nil when no such task exists. Used by batch-approve
// endpoints that supply a request ID rather than a task ID.
func (r *PostgresRepository) FindPendingTaskForActor(ctx context.Context, requestID uuid.UUID, actorID uuid.UUID) (*entity.ApprovalTask, error) {
	return scanTask(r.pool.QueryRow(ctx,
		taskSelect+` WHERE t.approval_request_id = $1 AND t.assigned_user_id = $2 AND t.status = 'PENDING'`,
		requestID, actorID))
}

func (r *PostgresRepository) queryTasks(ctx context.Context, sql string, args ...any) ([]*entity.ApprovalTask, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listing approval tasks: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalTask
	for rows.Next() {
		t, err := scanTask(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}
