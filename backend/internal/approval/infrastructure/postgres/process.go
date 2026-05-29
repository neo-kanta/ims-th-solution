package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

const processConfigSelect = `
	SELECT id, process_code, process_name, process_type, contract_type, contract_id,
	       effective_date, is_active, group_approval_enabled, require_team_approval,
	       created_by, created_at, updated_by, updated_at
	FROM approval__process_configs`

func scanProcessConfig(row pgx.Row) (*entity.ApprovalProcessConfig, error) {
	var c entity.ApprovalProcessConfig
	var processType, contractType string
	err := row.Scan(&c.ID, &c.ProcessCode, &c.ProcessName, &processType, &contractType, &c.ContractID,
		&c.EffectiveDate, &c.IsActive, &c.GroupApprovalEnabled, &c.RequireTeamApproval,
		&c.CreatedBy, &c.CreatedAt, &c.UpdatedBy, &c.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval process config: %w", err)
	}
	c.ProcessType = vo.ProcessType(processType)
	c.ContractType = vo.ContractType(contractType)
	return &c, nil
}

// CreateConfig inserts a process config.
func (r *PostgresRepository) CreateConfig(ctx context.Context, tx pgx.Tx, c *entity.ApprovalProcessConfig) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__process_configs
			(id, process_code, process_name, process_type, contract_type, contract_id, effective_date,
			 is_active, group_approval_enabled, require_team_approval, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`,
		c.ID, c.ProcessCode, c.ProcessName, string(c.ProcessType), string(c.ContractType), c.ContractID,
		c.EffectiveDate, c.IsActive, c.GroupApprovalEnabled, c.RequireTeamApproval, c.CreatedBy, c.UpdatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Conflict("approval process code already exists")
		}
		return fmt.Errorf("inserting approval process config: %w", err)
	}
	return nil
}

// UpdateConfig updates mutable config fields.
func (r *PostgresRepository) UpdateConfig(ctx context.Context, tx pgx.Tx, c *entity.ApprovalProcessConfig) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__process_configs
		SET process_name=$2, process_type=$3, contract_type=$4, contract_id=$5, effective_date=$6,
		    is_active=$7, group_approval_enabled=$8, require_team_approval=$9, updated_by=$10
		WHERE id=$1`,
		c.ID, c.ProcessName, string(c.ProcessType), string(c.ContractType), c.ContractID, c.EffectiveDate,
		c.IsActive, c.GroupApprovalEnabled, c.RequireTeamApproval, c.UpdatedBy)
	if err != nil {
		return fmt.Errorf("updating approval process config: %w", err)
	}
	return nil
}

// SetConfigActive flips the active flag.
func (r *PostgresRepository) SetConfigActive(ctx context.Context, tx pgx.Tx, id uuid.UUID, active bool, updatedBy *uuid.UUID) error {
	_, err := tx.Exec(ctx, `UPDATE approval__process_configs SET is_active=$2, updated_by=$3 WHERE id=$1`, id, active, updatedBy)
	if err != nil {
		return fmt.Errorf("setting approval process active: %w", err)
	}
	return nil
}

// GetConfig returns a config (with stages) by id or nil.
func (r *PostgresRepository) GetConfig(ctx context.Context, id uuid.UUID) (*entity.ApprovalProcessConfig, error) {
	c, err := scanProcessConfig(r.pool.QueryRow(ctx, processConfigSelect+" WHERE id = $1", id))
	if err != nil || c == nil {
		return c, err
	}
	stages, err := r.ListStages(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Stages = stages
	return c, nil
}

// GetConfigByCode returns a config by code or nil.
func (r *PostgresRepository) GetConfigByCode(ctx context.Context, code string) (*entity.ApprovalProcessConfig, error) {
	return scanProcessConfig(r.pool.QueryRow(ctx, processConfigSelect+" WHERE process_code = $1", code))
}

// ListConfigs lists process configs with optional filters (stages included).
func (r *PostgresRepository) ListConfigs(ctx context.Context, f domain.ProcessListFilter) ([]*entity.ApprovalProcessConfig, error) {
	conds := []string{"1=1"}
	args := []any{}
	idx := 1
	if f.ProcessType != "" {
		conds = append(conds, fmt.Sprintf("process_type = $%d", idx))
		args = append(args, string(f.ProcessType))
		idx++
	}
	if f.ActiveOnly {
		conds = append(conds, "is_active = true")
	}
	sql := processConfigSelect
	for i, c := range conds {
		if i == 0 {
			sql += " WHERE " + c
		} else {
			sql += " AND " + c
		}
	}
	sql += " ORDER BY process_type, effective_date DESC"
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listing approval process configs: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalProcessConfig
	for rows.Next() {
		c, err := scanProcessConfig(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for _, c := range out {
		stages, err := r.ListStages(ctx, c.ID)
		if err != nil {
			return nil, err
		}
		c.Stages = stages
	}
	return out, nil
}

const stageSelect = `
	SELECT id, process_config_id, stage_number, stage_name, approver_mode, approver_user_id,
	       approval_group_id, required_approval_count, is_final_stage, reject_policy, created_at, updated_at
	FROM approval__process_stages`

func scanStage(row pgx.Row) (entity.ApprovalProcessStage, error) {
	var s entity.ApprovalProcessStage
	var mode string
	err := row.Scan(&s.ID, &s.ProcessConfigID, &s.StageNumber, &s.StageName, &mode, &s.ApproverUserID,
		&s.ApprovalGroupID, &s.RequiredApprovalCount, &s.IsFinalStage, &s.RejectPolicy, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return s, err
	}
	s.ApproverMode = vo.ApproverMode(mode)
	return s, nil
}

// ListStages returns all stages of a config ordered by stage number.
func (r *PostgresRepository) ListStages(ctx context.Context, configID uuid.UUID) ([]entity.ApprovalProcessStage, error) {
	rows, err := r.pool.Query(ctx, stageSelect+" WHERE process_config_id = $1 ORDER BY stage_number", configID)
	if err != nil {
		return nil, fmt.Errorf("listing approval process stages: %w", err)
	}
	defer rows.Close()
	var out []entity.ApprovalProcessStage
	for rows.Next() {
		s, err := scanStage(rows)
		if err != nil {
			return nil, fmt.Errorf("scanning approval process stage: %w", err)
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

// ReplaceStages deletes existing stages and inserts the provided set.
func (r *PostgresRepository) ReplaceStages(ctx context.Context, tx pgx.Tx, configID uuid.UUID, stages []entity.ApprovalProcessStage) error {
	if _, err := tx.Exec(ctx, `DELETE FROM approval__process_stages WHERE process_config_id = $1`, configID); err != nil {
		return fmt.Errorf("clearing approval process stages: %w", err)
	}
	for _, s := range stages {
		id := s.ID
		if id == uuid.Nil {
			id = uuid.New()
		}
		_, err := tx.Exec(ctx, `
			INSERT INTO approval__process_stages
				(id, process_config_id, stage_number, stage_name, approver_mode, approver_user_id,
				 approval_group_id, required_approval_count, is_final_stage, reject_policy)
			VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
			id, configID, s.StageNumber, s.StageName, string(s.ApproverMode), s.ApproverUserID,
			s.ApprovalGroupID, s.RequiredApprovalCount, s.IsFinalStage, s.RejectPolicy)
		if err != nil {
			return fmt.Errorf("inserting approval process stage: %w", err)
		}
	}
	return nil
}

// Resolve returns the best-matching active config (with stages) for the process
// type and contract scope, honouring effective_date <= asOf. Match priority:
//  1. exact contract_id
//  2. contract_id IS NULL + matching contract_type
//  3. contract_id IS NULL + COMPANY (global)
func (r *PostgresRepository) Resolve(ctx context.Context, pt vo.ProcessType, contractID *uuid.UUID, ct vo.ContractType, asOf time.Time) (*entity.ApprovalProcessConfig, error) {
	row := r.pool.QueryRow(ctx, `
		SELECT id, process_code, process_name, process_type, contract_type, contract_id,
		       effective_date, is_active, group_approval_enabled, require_team_approval,
		       created_by, created_at, updated_by, updated_at
		FROM approval__process_configs
		WHERE is_active = true
		  AND process_type = $1
		  AND effective_date <= $4
		  AND (
		        ($2::uuid IS NOT NULL AND contract_id = $2)
		     OR (contract_id IS NULL AND contract_type = $3)
		     OR (contract_id IS NULL AND contract_type = 'COMPANY')
		  )
		ORDER BY
		  CASE
		    WHEN $2::uuid IS NOT NULL AND contract_id = $2 THEN 1
		    WHEN contract_id IS NULL AND contract_type = $3 THEN 2
		    ELSE 3
		  END ASC,
		  effective_date DESC
		LIMIT 1`,
		string(pt), contractID, string(ct), asOf)
	c, err := scanProcessConfig(row)
	if err != nil || c == nil {
		return c, err
	}
	stages, err := r.ListStages(ctx, c.ID)
	if err != nil {
		return nil, err
	}
	c.Stages = stages
	return c, nil
}
