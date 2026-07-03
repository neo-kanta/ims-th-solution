package postgres

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/valueobject"
)

// CreateSignature inserts an approval signature/stamp record.
func (r *PostgresRepository) CreateSignature(ctx context.Context, tx pgx.Tx, s *entity.ApprovalSignatureRecord) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	_, err := tx.Exec(ctx, `
		INSERT INTO approval__signature_records
			(id, approval_request_id, stage_number, signer_user_id, signer_display_name, signer_title,
			 signed_at, is_proxy_signature, proxy_for_user_id, signature_label)
		VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10)`,
		s.ID, s.ApprovalRequestID, s.StageNumber, s.SignerUserID, s.SignerDisplayName, s.SignerTitle,
		s.SignedAt, s.IsProxySignature, s.ProxyForUserID, string(s.SignatureLabel))
	if err != nil {
		return fmt.Errorf("inserting approval signature: %w", err)
	}
	return nil
}

// ListSignatures returns signature records for a request ordered by stage/time.
func (r *PostgresRepository) ListSignatures(ctx context.Context, requestID uuid.UUID) ([]*entity.ApprovalSignatureRecord, error) {
	rows, err := r.pool.Query(ctx, `
		SELECT id, approval_request_id, stage_number, signer_user_id, signer_display_name, signer_title,
		       signed_at, is_proxy_signature, proxy_for_user_id, signature_label, created_at
		FROM approval__signature_records
		WHERE approval_request_id = $1
		ORDER BY stage_number, signed_at`, requestID)
	if err != nil {
		return nil, fmt.Errorf("listing approval signatures: %w", err)
	}
	defer rows.Close()
	var out []*entity.ApprovalSignatureRecord
	for rows.Next() {
		var s entity.ApprovalSignatureRecord
		var label string
		if err := rows.Scan(&s.ID, &s.ApprovalRequestID, &s.StageNumber, &s.SignerUserID, &s.SignerDisplayName,
			&s.SignerTitle, &s.SignedAt, &s.IsProxySignature, &s.ProxyForUserID, &label, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning approval signature: %w", err)
		}
		s.SignatureLabel = vo.SignatureLabel(label)
		out = append(out, &s)
	}
	return out, rows.Err()
}
