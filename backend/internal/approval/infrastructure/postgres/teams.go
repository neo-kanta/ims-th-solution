package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

const teamSelect = `
	SELECT id, team_code, team_name, remarks, has_co_manager,
	       min_required_stamps, max_allowed_stamps, is_active,
	       created_by, created_at, updated_by, updated_at
	FROM approval__teams`

func scanTeam(row pgx.Row) (*entity.ApprovalTeam, error) {
	var t entity.ApprovalTeam
	err := row.Scan(&t.ID, &t.TeamCode, &t.TeamName, &t.Remarks, &t.HasCoManager,
		&t.MinRequiredStamps, &t.MaxAllowedStamps, &t.IsActive,
		&t.CreatedBy, &t.CreatedAt, &t.UpdatedBy, &t.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval team: %w", err)
	}
	return &t, nil
}

// CreateTeam inserts an approval team.
func (r *PostgresRepository) CreateTeam(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTeam) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__teams
			(id, team_code, team_name, remarks, has_co_manager, min_required_stamps, max_allowed_stamps, is_active, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		t.ID, t.TeamCode, t.TeamName, t.Remarks, t.HasCoManager,
		t.MinRequiredStamps, t.MaxAllowedStamps, t.IsActive, t.CreatedBy, t.UpdatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Conflict("approval team code already exists")
		}
		return fmt.Errorf("inserting approval team: %w", err)
	}
	return nil
}

// UpdateTeam updates mutable team fields.
func (r *PostgresRepository) UpdateTeam(ctx context.Context, tx pgx.Tx, t *entity.ApprovalTeam) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__teams
		SET team_name=$2, remarks=$3, has_co_manager=$4, min_required_stamps=$5, max_allowed_stamps=$6, is_active=$7, updated_by=$8
		WHERE id=$1`,
		t.ID, t.TeamName, t.Remarks, t.HasCoManager, t.MinRequiredStamps, t.MaxAllowedStamps, t.IsActive, t.UpdatedBy)
	if err != nil {
		return fmt.Errorf("updating approval team: %w", err)
	}
	return nil
}

// GetTeam returns a team by id or nil.
func (r *PostgresRepository) GetTeam(ctx context.Context, id uuid.UUID) (*entity.ApprovalTeam, error) {
	return scanTeam(r.pool.QueryRow(ctx, teamSelect+" WHERE id = $1", id))
}

// GetTeamByCode returns a team by code or nil.
func (r *PostgresRepository) GetTeamByCode(ctx context.Context, code string) (*entity.ApprovalTeam, error) {
	return scanTeam(r.pool.QueryRow(ctx, teamSelect+" WHERE team_code = $1", code))
}

// ListTeams returns all teams ordered by code.
func (r *PostgresRepository) ListTeams(ctx context.Context) ([]*entity.ApprovalTeam, error) {
	rows, err := r.pool.Query(ctx, teamSelect+" ORDER BY team_code")
	if err != nil {
		return nil, fmt.Errorf("listing approval teams: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalTeam
	for rows.Next() {
		t, err := scanTeam(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// AssignContract assigns a contract to a team (active).
func (r *PostgresRepository) AssignContract(ctx context.Context, tx pgx.Tx, c *entity.ApprovalTeamContract) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__team_contracts (id, team_id, contract_id, effective_date, is_active)
		VALUES ($1,$2,$3,$4,$5)`,
		c.ID, c.TeamID, c.ContractID, c.EffectiveDate, c.IsActive)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Conflict("contract is already assigned to an active approval team")
		}
		return fmt.Errorf("assigning contract to team: %w", err)
	}
	return nil
}

// DeactivateContractAssignment deactivates any active assignment for a contract.
func (r *PostgresRepository) DeactivateContractAssignment(ctx context.Context, tx pgx.Tx, contractID uuid.UUID) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__team_contracts SET is_active = false
		WHERE contract_id = $1 AND is_active = true`, contractID)
	if err != nil {
		return fmt.Errorf("deactivating contract assignment: %w", err)
	}
	return nil
}

// GetActiveTeamForContract returns the active team for a contract or nil.
func (r *PostgresRepository) GetActiveTeamForContract(ctx context.Context, contractID uuid.UUID) (*entity.ApprovalTeam, error) {
	row := r.pool.QueryRow(ctx, teamSelect+`
		WHERE id = (
			SELECT team_id FROM approval__team_contracts
			WHERE contract_id = $1 AND is_active = true
			LIMIT 1
		) AND is_active = true`, contractID)
	return scanTeam(row)
}

// ListTeamContracts returns all contract assignments for a team.
func (r *PostgresRepository) ListTeamContracts(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamContract, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, team_id, contract_id, effective_date, is_active, created_at, updated_at
		FROM approval__team_contracts WHERE team_id = $1 ORDER BY created_at`, teamID)
	if err != nil {
		return nil, fmt.Errorf("listing team contracts: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalTeamContract
	for rows.Next() {
		var c entity.ApprovalTeamContract
		if err := rows.Scan(&c.ID, &c.TeamID, &c.ContractID, &c.EffectiveDate, &c.IsActive, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning team contract: %w", err)
		}
		out = append(out, &c)
	}
	return out, rows.Err()
}

const teamMemberSelect = `
	SELECT m.id, m.team_id, m.user_id, m.member_type, m.priority_order, m.is_active, m.created_at, m.updated_at,
	       COALESCE(u.display_name, u.username, '')
	FROM approval__team_members m
	LEFT JOIN iam_users u ON u.id = m.user_id`

func scanTeamMember(row pgx.Row) (*entity.ApprovalTeamMember, error) {
	var m entity.ApprovalTeamMember
	var memberType string
	err := row.Scan(&m.ID, &m.TeamID, &m.UserID, &memberType, &m.PriorityOrder, &m.IsActive, &m.CreatedAt, &m.UpdatedAt, &m.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval team member: %w", err)
	}
	m.MemberType = vo.TeamMemberType(memberType)
	return &m, nil
}

// AddTeamMember inserts a team member.
func (r *PostgresRepository) AddTeamMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalTeamMember) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__team_members (id, team_id, user_id, member_type, priority_order, is_active)
		VALUES ($1,$2,$3,$4,$5,$6)`,
		m.ID, m.TeamID, m.UserID, string(m.MemberType), m.PriorityOrder, m.IsActive)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Conflict("user is already a member of this team")
		}
		return fmt.Errorf("inserting approval team member: %w", err)
	}
	return nil
}

// UpdateTeamMember updates a team member.
func (r *PostgresRepository) UpdateTeamMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalTeamMember) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__team_members SET member_type=$2, priority_order=$3, is_active=$4 WHERE id=$1`,
		m.ID, string(m.MemberType), m.PriorityOrder, m.IsActive)
	if err != nil {
		return fmt.Errorf("updating approval team member: %w", err)
	}
	return nil
}

// RemoveTeamMember deletes a team member.
func (r *PostgresRepository) RemoveTeamMember(ctx context.Context, tx pgx.Tx, memberID uuid.UUID) error {
	_, err := tx.Exec(ctx, `DELETE FROM approval__team_members WHERE id=$1`, memberID)
	if err != nil {
		return fmt.Errorf("removing approval team member: %w", err)
	}
	return nil
}

// GetTeamMember returns a team member by id or nil.
func (r *PostgresRepository) GetTeamMember(ctx context.Context, memberID uuid.UUID) (*entity.ApprovalTeamMember, error) {
	return scanTeamMember(r.pool.QueryRow(ctx, teamMemberSelect+" WHERE m.id = $1", memberID))
}

// ListTeamMembers returns all members of a team ordered by priority.
func (r *PostgresRepository) ListTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error) {
	return r.queryTeamMembers(ctx, teamMemberSelect+" WHERE m.team_id = $1 ORDER BY m.priority_order, m.created_at", teamID)
}

// ListEligibleTeamMembers returns active reviewer-agents ordered by priority.
func (r *PostgresRepository) ListEligibleTeamMembers(ctx context.Context, teamID uuid.UUID) ([]*entity.ApprovalTeamMember, error) {
	return r.queryTeamMembers(ctx,
		teamMemberSelect+" WHERE m.team_id = $1 AND m.is_active = true AND m.member_type = 'REVIEWER_AGENT' ORDER BY m.priority_order, m.created_at",
		teamID)
}

func (r *PostgresRepository) queryTeamMembers(ctx context.Context, sql string, args ...any) ([]*entity.ApprovalTeamMember, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listing approval team members: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalTeamMember
	for rows.Next() {
		m, err := scanTeamMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
