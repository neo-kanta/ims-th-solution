package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/iam/domain/entity"
)

// PostgresMFARepository implements domain.MFARepository.
type PostgresMFARepository struct {
	pool *pgxpool.Pool
}

// NewPostgresMFARepository creates a new MFA repository.
func NewPostgresMFARepository(pool *pgxpool.Pool) *PostgresMFARepository {
	return &PostgresMFARepository{pool: pool}
}

// FindByUserID retrieves the MFA enrollment for a user.
func (r *PostgresMFARepository) FindByUserID(ctx context.Context, userID uuid.UUID) (*entity.MFAEnrollment, error) {
	query := `
		SELECT id, user_id, mfa_type, secret_encrypted, is_verified, is_enabled,
		       created_at, verified_at, disabled_at
		FROM iam_mfa_enrollments
		WHERE user_id = $1
	`
	var m entity.MFAEnrollment
	err := r.pool.QueryRow(ctx, query, userID).Scan(
		&m.ID, &m.UserID, &m.MFAType, &m.SecretEncrypted,
		&m.IsVerified, &m.IsEnabled,
		&m.CreatedAt, &m.VerifiedAt, &m.DisabledAt,
	)
	if err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, fmt.Errorf("querying MFA enrollment: %w", err)
	}
	return &m, nil
}

// CreateEnrollment inserts a new MFA enrollment.
func (r *PostgresMFARepository) CreateEnrollment(ctx context.Context, enrollment *entity.MFAEnrollment) error {
	query := `
		INSERT INTO iam_mfa_enrollments (id, user_id, mfa_type, secret_encrypted)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (user_id, mfa_type)
		DO UPDATE SET secret_encrypted = EXCLUDED.secret_encrypted,
		             is_verified = false,
		             is_enabled = false,
		             verified_at = NULL,
		             disabled_at = NULL
	`
	_, err := r.pool.Exec(ctx, query,
		enrollment.ID, enrollment.UserID, enrollment.MFAType, enrollment.SecretEncrypted,
	)
	if err != nil {
		return fmt.Errorf("inserting MFA enrollment: %w", err)
	}
	return nil
}

// EnableEnrollment marks the enrollment as verified and enabled.
func (r *PostgresMFARepository) EnableEnrollment(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE iam_mfa_enrollments
		SET is_verified = true, is_enabled = true, verified_at = NOW(), disabled_at = NULL
		WHERE user_id = $1
	`
	result, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("enabling MFA enrollment: %w", err)
	}
	if result.RowsAffected() == 0 {
		return fmt.Errorf("MFA enrollment not found")
	}
	return nil
}

// DisableEnrollment disables MFA for a user.
func (r *PostgresMFARepository) DisableEnrollment(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE iam_mfa_enrollments
		SET is_enabled = false, disabled_at = NOW()
		WHERE user_id = $1
	`
	_, err := r.pool.Exec(ctx, query, userID)
	if err != nil {
		return fmt.Errorf("disabling MFA enrollment: %w", err)
	}
	return nil
}

// DeleteEnrollment removes the MFA enrollment entirely.
func (r *PostgresMFARepository) DeleteEnrollment(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM iam_mfa_enrollments WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("deleting MFA enrollment: %w", err)
	}
	return nil
}

// StoreRecoveryCodes inserts hashed recovery codes (replaces existing ones).
func (r *PostgresMFARepository) StoreRecoveryCodes(ctx context.Context, codes []entity.MFARecoveryCode) error {
	if len(codes) == 0 {
		return nil
	}

	// Delete existing codes for the user first
	_, err := r.pool.Exec(ctx,
		`DELETE FROM iam_mfa_recovery_codes WHERE user_id = $1`,
		codes[0].UserID,
	)
	if err != nil {
		return fmt.Errorf("clearing old recovery codes: %w", err)
	}

	// Insert new codes
	for _, code := range codes {
		_, err := r.pool.Exec(ctx,
			`INSERT INTO iam_mfa_recovery_codes (id, user_id, code_hash) VALUES ($1, $2, $3)`,
			code.ID, code.UserID, code.CodeHash,
		)
		if err != nil {
			return fmt.Errorf("inserting recovery code: %w", err)
		}
	}
	return nil
}

// FindUnusedRecoveryCodes retrieves all unused recovery codes for a user.
func (r *PostgresMFARepository) FindUnusedRecoveryCodes(ctx context.Context, userID uuid.UUID) ([]entity.MFARecoveryCode, error) {
	query := `
		SELECT id, user_id, code_hash, is_used, used_at, created_at
		FROM iam_mfa_recovery_codes
		WHERE user_id = $1 AND is_used = false
		ORDER BY created_at
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("querying recovery codes: %w", err)
	}
	defer rows.Close()

	var codes []entity.MFARecoveryCode
	for rows.Next() {
		var c entity.MFARecoveryCode
		if err := rows.Scan(&c.ID, &c.UserID, &c.CodeHash, &c.IsUsed, &c.UsedAt, &c.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning recovery code: %w", err)
		}
		codes = append(codes, c)
	}
	if codes == nil {
		codes = []entity.MFARecoveryCode{}
	}
	return codes, nil
}

// UseRecoveryCode marks a specific recovery code as used.
func (r *PostgresMFARepository) UseRecoveryCode(ctx context.Context, codeID uuid.UUID) error {
	query := `UPDATE iam_mfa_recovery_codes SET is_used = true, used_at = NOW() WHERE id = $1`
	_, err := r.pool.Exec(ctx, query, codeID)
	if err != nil {
		return fmt.Errorf("using recovery code: %w", err)
	}
	return nil
}

// DeleteRecoveryCodes removes all recovery codes for a user.
func (r *PostgresMFARepository) DeleteRecoveryCodes(ctx context.Context, userID uuid.UUID) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM iam_mfa_recovery_codes WHERE user_id = $1`, userID)
	if err != nil {
		return fmt.Errorf("deleting recovery codes: %w", err)
	}
	return nil
}
