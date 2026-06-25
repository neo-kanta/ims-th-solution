package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/watchlist/domain/repository"
)

// PostgresWatchlistItemRepository implements repository.WatchlistItemRepository.
type PostgresWatchlistItemRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresWatchlistItemRepository(pool *pgxpool.Pool) *PostgresWatchlistItemRepository {
	return &PostgresWatchlistItemRepository{pool: pool}
}

const itemSelect = `
	SELECT id, scope_type, owner_user_id, portfolio_id, security_id,
	       display_order, pinned, note, status,
	       created_by, updated_by, deleted_by,
	       created_at, updated_at, deleted_at
	FROM watchlist_items`

func scanItem(row pgx.Row) (*entity.WatchlistItem, error) {
	var i entity.WatchlistItem
	var scopeType, status string
	err := row.Scan(
		&i.ID, &scopeType, &i.OwnerUserID, &i.PortfolioID, &i.SecurityID,
		&i.DisplayOrder, &i.Pinned, &i.Note, &status,
		&i.CreatedBy, &i.UpdatedBy, &i.DeletedBy,
		&i.CreatedAt, &i.UpdatedAt, &i.DeletedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	i.ScopeType = entity.ScopeType(scopeType)
	i.Status = entity.ItemStatus(status)
	return &i, nil
}

func (r *PostgresWatchlistItemRepository) Create(ctx context.Context, tx pgx.Tx, item *entity.WatchlistItem) error {
	now := time.Now().UTC()
	item.CreatedAt = now
	item.UpdatedAt = now
	_, err := tx.Exec(ctx, `
		INSERT INTO watchlist_items (
			id, scope_type, owner_user_id, portfolio_id, security_id,
			display_order, pinned, note, status,
			created_by, updated_by,
			created_at, updated_at
		) VALUES (
			$1, $2, $3, $4, $5,
			$6, $7, $8, $9,
			$10, $11,
			$12, $13
		)`,
		item.ID, string(item.ScopeType), item.OwnerUserID, item.PortfolioID, item.SecurityID,
		item.DisplayOrder, item.Pinned, item.Note, string(item.Status),
		item.CreatedBy, item.UpdatedBy,
		item.CreatedAt, item.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domain.ErrDuplicateWatchlistItem
		}
		return fmt.Errorf("creating watchlist item: %w", err)
	}
	return nil
}

func (r *PostgresWatchlistItemRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.WatchlistItem, error) {
	row := r.pool.QueryRow(ctx, itemSelect+` WHERE id = $1 AND deleted_at IS NULL`, id)
	return scanItem(row)
}

func (r *PostgresWatchlistItemRepository) Update(ctx context.Context, tx pgx.Tx, item *entity.WatchlistItem) error {
	item.UpdatedAt = time.Now().UTC()
	_, err := tx.Exec(ctx, `
		UPDATE watchlist_items SET
			pinned = $2, note = $3, status = $4, display_order = $5,
			updated_by = $6, updated_at = $7
		WHERE id = $1 AND deleted_at IS NULL`,
		item.ID, item.Pinned, item.Note, string(item.Status), item.DisplayOrder,
		item.UpdatedBy, item.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("updating watchlist item: %w", err)
	}
	return nil
}

func (r *PostgresWatchlistItemRepository) SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error {
	now := time.Now().UTC()
	tag, err := tx.Exec(ctx, `
		UPDATE watchlist_items SET
			status = 'DISABLED', deleted_by = $2, deleted_at = $3, updated_at = $3
		WHERE id = $1 AND deleted_at IS NULL`,
		id, deletedBy, now,
	)
	if err != nil {
		return fmt.Errorf("soft-deleting watchlist item: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrItemNotFound
	}
	return nil
}

func (r *PostgresWatchlistItemRepository) List(ctx context.Context, filter repository.WatchlistItemFilter) ([]*entity.WatchlistItem, int, error) {
	conds := []string{"deleted_at IS NULL"}
	args := []interface{}{}
	idx := 1

	if filter.UnionPersonalOwnerID != nil {
		// Unscoped union: personal rows owned by actor OR portfolio rows accessible via fund IDs.
		if filter.UnionPortfolioAll {
			conds = append(conds, fmt.Sprintf(
				"(scope_type = 'PERSONAL' AND owner_user_id = $%d OR scope_type = 'PORTFOLIO')", idx,
			))
			args = append(args, *filter.UnionPersonalOwnerID)
			idx++
		} else if len(filter.FundIDs) > 0 {
			conds = append(conds, fmt.Sprintf(
				"(scope_type = 'PERSONAL' AND owner_user_id = $%d OR scope_type = 'PORTFOLIO' AND EXISTS (SELECT 1 FROM investment__portfolios p WHERE p.id = watchlist_items.portfolio_id AND p.fund_id = ANY($%d)))",
				idx, idx+1,
			))
			args = append(args, *filter.UnionPersonalOwnerID, filter.FundIDs)
			idx += 2
		} else {
			conds = append(conds, fmt.Sprintf("scope_type = 'PERSONAL' AND owner_user_id = $%d", idx))
			args = append(args, *filter.UnionPersonalOwnerID)
			idx++
		}
	} else {
		if filter.OwnerUserID != nil {
			conds = append(conds, fmt.Sprintf("owner_user_id = $%d", idx))
			args = append(args, *filter.OwnerUserID)
			idx++
		}
		if len(filter.PortfolioIDs) > 0 {
			conds = append(conds, fmt.Sprintf("portfolio_id = ANY($%d)", idx))
			args = append(args, filter.PortfolioIDs)
			idx++
		}
		if len(filter.FundIDs) > 0 {
			conds = append(conds, fmt.Sprintf(
				"EXISTS (SELECT 1 FROM investment__portfolios p WHERE p.id = watchlist_items.portfolio_id AND p.fund_id = ANY($%d))", idx,
			))
			args = append(args, filter.FundIDs)
			idx++
		}
		if filter.ScopeType != nil {
			conds = append(conds, fmt.Sprintf("scope_type = $%d", idx))
			args = append(args, string(*filter.ScopeType))
			idx++
		}
	}
	if filter.SecurityID != nil {
		conds = append(conds, fmt.Sprintf("security_id = $%d", idx))
		args = append(args, *filter.SecurityID)
		idx++
	}
	if !filter.IncludeDisabled {
		conds = append(conds, "status = 'ACTIVE'")
	}

	where := "WHERE " + strings.Join(conds, " AND ")

	// Count
	var total int
	err := r.pool.QueryRow(ctx, "SELECT COUNT(*) FROM watchlist_items "+where, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("counting watchlist items: %w", err)
	}
	if total == 0 {
		return []*entity.WatchlistItem{}, 0, nil
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 50
	}
	offset := filter.Offset

	q := itemSelect + " " + where +
		" ORDER BY pinned DESC, display_order ASC NULLS LAST, updated_at DESC, id ASC" +
		fmt.Sprintf(" LIMIT $%d OFFSET $%d", idx, idx+1)
	args = append(args, limit, offset)

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("listing watchlist items: %w", err)
	}
	defer rows.Close()

	var items []*entity.WatchlistItem
	for rows.Next() {
		item, err := scanItem(rows)
		if err != nil {
			return nil, 0, err
		}
		if item != nil {
			items = append(items, item)
		}
	}
	return items, total, rows.Err()
}
