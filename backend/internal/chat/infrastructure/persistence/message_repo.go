package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	chatdomain "github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// PostgresMessageRepository persists Message rows in chat_messages.
type PostgresMessageRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresMessageRepository constructs the repository.
func NewPostgresMessageRepository(pool *pgxpool.Pool) *PostgresMessageRepository {
	return &PostgresMessageRepository{pool: pool}
}

var _ chatdomain.MessageRepository = (*PostgresMessageRepository)(nil)

// Append inserts one message. Messages are immutable once written.
func (r *PostgresMessageRepository) Append(ctx context.Context, m *entity.Message) error {
	const q = `
		INSERT INTO chat_messages (
			id, session_id, role, content,
			provenance_map, raw_provider_payload, correlation_id, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	provenance := m.ProvenanceMap
	if len(provenance) == 0 {
		provenance = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, q,
		m.ID,
		m.SessionID,
		string(m.Role),
		m.Content,
		provenance,
		nullableJSON(m.RawProviderPayload),
		nullableString(m.CorrelationID),
		m.CreatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert chat_message: %w", err)
	}
	return nil
}

// ListBySession returns messages in chronological order, oldest first,
// capped at limit (most recent wins when the cap is exceeded — the policy
// trimmer in domain/policy is the one that handles that).
func (r *PostgresMessageRepository) ListBySession(ctx context.Context, sessionID uuid.UUID, limit int) ([]entity.Message, error) {
	if limit <= 0 {
		limit = 200
	}
	const q = `
		SELECT id, session_id, role, content,
		       provenance_map, COALESCE(raw_provider_payload, 'null'::jsonb), created_at
		FROM (
			SELECT id, session_id, role, content,
			       provenance_map, raw_provider_payload, created_at
			FROM chat_messages
			WHERE session_id = $1
			ORDER BY created_at DESC, id DESC
			LIMIT $2
		) recent
		ORDER BY created_at ASC, id ASC
	`
	rows, err := r.pool.Query(ctx, q, sessionID, limit)
	if err != nil {
		return nil, fmt.Errorf("select chat_messages: %w", err)
	}
	defer rows.Close()

	var out []entity.Message
	for rows.Next() {
		var (
			m    entity.Message
			role string
			prov []byte
			raw  []byte
		)
		if err := rows.Scan(&m.ID, &m.SessionID, &role, &m.Content, &prov, &raw, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan chat_message: %w", err)
		}
		m.Role = valueobject.MessageRole(role)
		m.ProvenanceMap = json.RawMessage(prov)
		if len(raw) > 0 && string(raw) != "null" {
			m.RawProviderPayload = json.RawMessage(raw)
		}
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate chat_messages: %w", err)
	}
	return out, nil
}

// ListBySessionPaged returns one page of a session's messages in chronological
// order (oldest first) plus the total count. raw_provider_payload is never
// selected here — history reads must not expose the raw provider payload.
func (r *PostgresMessageRepository) ListBySessionPaged(ctx context.Context, sessionID uuid.UUID, limit, offset int) ([]entity.Message, int, error) {
	if limit <= 0 {
		limit = 20
	}
	if offset < 0 {
		offset = 0
	}

	var total int
	if err := r.pool.QueryRow(ctx, `SELECT COUNT(*) FROM chat_messages WHERE session_id = $1`, sessionID).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count chat_messages: %w", err)
	}

	const q = `
		SELECT id, session_id, role, content, provenance_map, created_at
		FROM chat_messages
		WHERE session_id = $1
		ORDER BY created_at ASC, id ASC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.pool.Query(ctx, q, sessionID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("select chat_messages paged: %w", err)
	}
	defer rows.Close()

	var out []entity.Message
	for rows.Next() {
		var (
			m    entity.Message
			role string
			prov []byte
		)
		if err := rows.Scan(&m.ID, &m.SessionID, &role, &m.Content, &prov, &m.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("scan chat_message: %w", err)
		}
		m.Role = valueobject.MessageRole(role)
		m.ProvenanceMap = json.RawMessage(prov)
		out = append(out, m)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, fmt.Errorf("iterate chat_messages: %w", err)
	}
	return out, total, nil
}

// nullableJSON returns nil for empty inputs so the column receives SQL NULL
// rather than the literal string "null".
func nullableJSON(b json.RawMessage) interface{} {
	if len(b) == 0 {
		return nil
	}
	return []byte(b)
}
