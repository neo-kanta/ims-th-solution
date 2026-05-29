package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

const pgUniqueViolation = "23505"

const groupSelect = `
	SELECT id, group_code, group_name, remarks, is_active,
	       created_by, created_at, updated_by, updated_at
	FROM approval__groups`

func scanGroup(row pgx.Row) (*entity.ApprovalGroup, error) {
	var g entity.ApprovalGroup
	err := row.Scan(&g.ID, &g.GroupCode, &g.GroupName, &g.Remarks, &g.IsActive,
		&g.CreatedBy, &g.CreatedAt, &g.UpdatedBy, &g.UpdatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval group: %w", err)
	}
	return &g, nil
}

// CreateGroup inserts an approval group.
func (r *PostgresRepository) CreateGroup(ctx context.Context, tx pgx.Tx, g *entity.ApprovalGroup) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__groups (id, group_code, group_name, remarks, is_active, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7)`,
		g.ID, g.GroupCode, g.GroupName, g.Remarks, g.IsActive, g.CreatedBy, g.UpdatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Conflict("approval group code already exists")
		}
		return fmt.Errorf("inserting approval group: %w", err)
	}
	return nil
}

// UpdateGroup updates mutable group fields.
func (r *PostgresRepository) UpdateGroup(ctx context.Context, tx pgx.Tx, g *entity.ApprovalGroup) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__groups
		SET group_name = $2, remarks = $3, is_active = $4, updated_by = $5
		WHERE id = $1`,
		g.ID, g.GroupName, g.Remarks, g.IsActive, g.UpdatedBy)
	if err != nil {
		return fmt.Errorf("updating approval group: %w", err)
	}
	return nil
}

// GetGroup returns a group by id or nil.
func (r *PostgresRepository) GetGroup(ctx context.Context, id uuid.UUID) (*entity.ApprovalGroup, error) {
	return scanGroup(r.pool.QueryRow(ctx, groupSelect+" WHERE id = $1", id))
}

// GetGroupByCode returns a group by code or nil.
func (r *PostgresRepository) GetGroupByCode(ctx context.Context, code string) (*entity.ApprovalGroup, error) {
	return scanGroup(r.pool.QueryRow(ctx, groupSelect+" WHERE group_code = $1", code))
}

// ListGroups lists groups with optional filters.
func (r *PostgresRepository) ListGroups(ctx context.Context, f domain.GroupListFilter) ([]*entity.ApprovalGroup, error) {
	conds := []string{"1=1"}
	args := []any{}
	idx := 1
	if f.ActiveOnly {
		conds = append(conds, "is_active = true")
	}
	if s := strings.TrimSpace(f.Search); s != "" {
		conds = append(conds, fmt.Sprintf("(group_code ILIKE $%d OR group_name ILIKE $%d)", idx, idx))
		args = append(args, "%"+s+"%")
		idx++
	}
	rows, err := r.pool.Query(ctx, groupSelect+" WHERE "+strings.Join(conds, " AND ")+" ORDER BY group_code", args...)
	if err != nil {
		return nil, fmt.Errorf("listing approval groups: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalGroup
	for rows.Next() {
		g, err := scanGroup(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, g)
	}
	return out, rows.Err()
}

const groupMemberSelect = `
	SELECT m.id, m.group_id, m.user_id, m.priority_order, m.member_type, m.status, m.is_active,
	       m.created_by, m.created_at, m.updated_by, m.updated_at,
	       COALESCE(u.display_name, u.username, '')
	FROM approval__group_members m
	LEFT JOIN iam_users u ON u.id = m.user_id`

func scanGroupMember(row pgx.Row) (*entity.ApprovalGroupMember, error) {
	var m entity.ApprovalGroupMember
	var memberType, status string
	err := row.Scan(&m.ID, &m.GroupID, &m.UserID, &m.PriorityOrder, &memberType, &status, &m.IsActive,
		&m.CreatedBy, &m.CreatedAt, &m.UpdatedBy, &m.UpdatedAt, &m.DisplayName)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("scanning approval group member: %w", err)
	}
	m.MemberType = vo.GroupMemberType(memberType)
	m.Status = vo.GroupMemberStatus(status)
	return &m, nil
}

// AddGroupMember inserts a group member.
func (r *PostgresRepository) AddGroupMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalGroupMember) error {
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__group_members
			(id, group_id, user_id, priority_order, member_type, status, is_active, created_by, updated_by)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		m.ID, m.GroupID, m.UserID, m.PriorityOrder, string(m.MemberType), string(m.Status),
		m.IsActive, m.CreatedBy, m.UpdatedBy)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
			return domain.Conflict("user is already a member of this group")
		}
		return fmt.Errorf("inserting approval group member: %w", err)
	}
	return nil
}

// UpdateGroupMember updates a member's priority, type, status and active flag.
func (r *PostgresRepository) UpdateGroupMember(ctx context.Context, tx pgx.Tx, m *entity.ApprovalGroupMember) error {
	_, err := tx.Exec(ctx, `
		UPDATE approval__group_members
		SET priority_order = $2, member_type = $3, status = $4, is_active = $5, updated_by = $6
		WHERE id = $1`,
		m.ID, m.PriorityOrder, string(m.MemberType), string(m.Status), m.IsActive, m.UpdatedBy)
	if err != nil {
		return fmt.Errorf("updating approval group member: %w", err)
	}
	return nil
}

// GetGroupMember returns a member by id or nil.
func (r *PostgresRepository) GetGroupMember(ctx context.Context, memberID uuid.UUID) (*entity.ApprovalGroupMember, error) {
	return scanGroupMember(r.pool.QueryRow(ctx, groupMemberSelect+" WHERE m.id = $1", memberID))
}

// ListGroupMembers returns all members of a group ordered by priority.
func (r *PostgresRepository) ListGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error) {
	return r.queryGroupMembers(ctx, groupMemberSelect+" WHERE m.group_id = $1 ORDER BY m.priority_order, m.created_at", groupID)
}

// ListEligibleGroupMembers returns active + APPROVED members ordered by priority.
func (r *PostgresRepository) ListEligibleGroupMembers(ctx context.Context, groupID uuid.UUID) ([]*entity.ApprovalGroupMember, error) {
	return r.queryGroupMembers(ctx,
		groupMemberSelect+" WHERE m.group_id = $1 AND m.is_active = true AND m.status = 'APPROVED' ORDER BY m.priority_order, m.created_at",
		groupID)
}

func (r *PostgresRepository) queryGroupMembers(ctx context.Context, sql string, args ...any) ([]*entity.ApprovalGroupMember, error) {
	rows, err := r.pool.Query(ctx, sql, args...)
	if err != nil {
		return nil, fmt.Errorf("listing approval group members: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalGroupMember
	for rows.Next() {
		m, err := scanGroupMember(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
