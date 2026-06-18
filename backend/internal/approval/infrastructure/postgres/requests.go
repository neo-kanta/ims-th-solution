package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

const requestSelect = `
	SELECT r.id, r.request_number, r.process_type, r.process_config_id, r.subject_type, r.subject_id,
	       r.subject_title, r.subject_reference, r.contract_id, r.portfolio_id, r.submitter_id, r.submitted_at,
	       r.current_stage_number, r.status, r.final_decision_by, r.final_decision_at, r.rejection_reason,
	       r.created_at, r.updated_at,
	       COALESCE(u.display_name, u.username, ''),
	       COALESCE(r.config_snapshot, '[]'::jsonb)
	FROM approval__requests r
	LEFT JOIN iam_users u ON u.id = r.submitter_id`

func scanRequest(row pgx.Row) (*entity.ApprovalRequest, error) {
	var r entity.ApprovalRequest
	var processType, subjectType, status string
	var snapshotJSON []byte
	err := row.Scan(&r.ID, &r.RequestNumber, &processType, &r.ProcessConfigID, &subjectType, &r.SubjectID,
		&r.SubjectTitle, &r.SubjectReference, &r.ContractID, &r.PortfolioID, &r.SubmitterID, &r.SubmittedAt,
		&r.CurrentStageNumber, &status, &r.FinalDecisionBy, &r.FinalDecisionAt, &r.RejectionReason,
		&r.CreatedAt, &r.UpdatedAt, &r.SubmitterName, &snapshotJSON)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval request: %w", err)
	}
	r.ProcessType = vo.ProcessType(processType)
	r.SubjectType = vo.SubjectType(subjectType)
	r.Status = vo.RequestStatus(status)
	if len(snapshotJSON) > 0 {
		if err := json.Unmarshal(snapshotJSON, &r.ConfigSnapshot); err != nil {
			return nil, fmt.Errorf("decoding config snapshot: %w", err)
		}
	}
	return &r, nil
}

// NextRequestNumber returns the next APR-###### sequence value.
func (r *PostgresRepository) NextRequestNumber(ctx context.Context, tx pgx.Tx) (string, error) {
	var n int64
	if err := tx.QueryRow(ctx, `SELECT nextval('approval__request_no_seq')`).Scan(&n); err != nil {
		return "", fmt.Errorf("allocating request number: %w", err)
	}
	return fmt.Sprintf("APR-%06d", n), nil
}

// CreateRequest inserts an approval request.
func (r *PostgresRepository) CreateRequest(ctx context.Context, tx pgx.Tx, req *entity.ApprovalRequest) error {
	snapshotJSON, err := json.Marshal(req.ConfigSnapshot)
	if err != nil {
		return fmt.Errorf("encoding config snapshot: %w", err)
	}
	_, err = tx.Exec(ctx, `
		INSERT INTO approval__requests
			(id, request_number, process_type, process_config_id, subject_type, subject_id, subject_title,
			 subject_reference, contract_id, portfolio_id, submitter_id, submitted_at, current_stage_number,
			 status, final_decision_by, final_decision_at, rejection_reason, config_snapshot)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14,$15,$16,$17,$18)`,
		req.ID, req.RequestNumber, string(req.ProcessType), req.ProcessConfigID, string(req.SubjectType),
		req.SubjectID, req.SubjectTitle, req.SubjectReference, req.ContractID, req.PortfolioID, req.SubmitterID,
		req.SubmittedAt, req.CurrentStageNumber, string(req.Status), req.FinalDecisionBy, req.FinalDecisionAt,
		req.RejectionReason, snapshotJSON)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.NewError(domain.ErrDuplicateActiveRequest, domain.ErrDuplicateActiveRequest.Error())
		}
		return fmt.Errorf("inserting approval request: %w", err)
	}
	return nil
}

// GetRequest returns a request by id or nil.
func (r *PostgresRepository) GetRequest(ctx context.Context, id uuid.UUID) (*entity.ApprovalRequest, error) {
	return scanRequest(r.pool.QueryRow(ctx, requestSelect+" WHERE r.id = $1", id))
}

// GetRequestForUpdate locks and returns a request inside the transaction.
func (r *PostgresRepository) GetRequestForUpdate(ctx context.Context, tx pgx.Tx, id uuid.UUID) (*entity.ApprovalRequest, error) {
	// FOR UPDATE cannot be combined with the LEFT JOIN aggregate cleanly in all
	// planners, so lock the base row first, then re-read with the display join.
	var locked uuid.UUID
	err := tx.QueryRow(ctx, `SELECT id FROM approval__requests WHERE id = $1 FOR UPDATE`, id).Scan(&locked)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("locking approval request: %w", err)
	}
	return scanRequest(tx.QueryRow(ctx, requestSelect+" WHERE r.id = $1", id))
}

// UpdateRequestState updates the mutable lifecycle fields of a request.
func (r *PostgresRepository) UpdateRequestState(ctx context.Context, tx pgx.Tx, req *entity.ApprovalRequest) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__requests
		SET current_stage_number=$2, status=$3, submitted_at=$4, final_decision_by=$5,
		    final_decision_at=$6, rejection_reason=$7
		WHERE id=$1`,
		req.ID, req.CurrentStageNumber, string(req.Status), req.SubmittedAt, req.FinalDecisionBy,
		req.FinalDecisionAt, req.RejectionReason)
	if err != nil {
		return fmt.Errorf("updating approval request state: %w", err)
	}
	return nil
}

// ListRequests lists requests with optional filters + pagination.
func (r *PostgresRepository) ListRequests(ctx context.Context, f domain.RequestListFilter) ([]*entity.ApprovalRequest, int, error) {
	conds := []string{"1=1"}
	args := []any{}
	idx := 1
	if f.Status != "" {
		conds = append(conds, fmt.Sprintf("r.status = $%d", idx))
		args = append(args, string(f.Status))
		idx++
	}
	if f.ProcessType != "" {
		conds = append(conds, fmt.Sprintf("r.process_type = $%d", idx))
		args = append(args, string(f.ProcessType))
		idx++
	}
	if f.SubmitterID != nil {
		conds = append(conds, fmt.Sprintf("r.submitter_id = $%d", idx))
		args = append(args, *f.SubmitterID)
		idx++
	}
	if f.ContractID != nil {
		conds = append(conds, fmt.Sprintf("r.contract_id = $%d", idx))
		args = append(args, *f.ContractID)
		idx++
	}
	where := ""
	for i, c := range conds {
		if i == 0 {
			where = " WHERE " + c
		} else {
			where += " AND " + c
		}
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM approval__requests r`+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("counting approval requests: %w", err)
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
	args = append(args, limit, offset)
	sql := requestSelect + where + fmt.Sprintf(" ORDER BY r.created_at DESC LIMIT $%d OFFSET $%d", idx, idx+1)

	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing approval requests: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalRequest
	for rows.Next() {
		req, err := scanRequest(rows)
		if err != nil {
			return nil, 0, err
		}
		out = append(out, req)
	}
	return out, total, rows.Err()
}

// GetActiveBySubject returns the active (non-terminal) request for a subject or nil.
func (r *PostgresRepository) GetActiveBySubject(ctx context.Context, st vo.SubjectType, subjectID uuid.UUID) (*entity.ApprovalRequest, error) {
	return scanRequest(r.pool.QueryRow(ctx, requestSelect+`
		WHERE r.subject_type = $1 AND r.subject_id = $2
		  AND r.status IN ('DRAFT','SUBMITTED','PENDING_APPROVAL')
		ORDER BY r.created_at DESC LIMIT 1`, string(st), subjectID))
}

// GetLatestBySubject returns the most recent request for a subject or nil.
func (r *PostgresRepository) GetLatestBySubject(ctx context.Context, st vo.SubjectType, subjectID uuid.UUID) (*entity.ApprovalRequest, error) {
	return scanRequest(r.pool.QueryRow(ctx, requestSelect+`
		WHERE r.subject_type = $1 AND r.subject_id = $2
		ORDER BY r.created_at DESC LIMIT 1`, string(st), subjectID))
}
