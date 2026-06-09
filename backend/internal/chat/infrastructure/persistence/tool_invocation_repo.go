package persistence

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	chatdomain "github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
)

// PostgresToolInvocationRepository persists ToolInvocation rows.
type PostgresToolInvocationRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresToolInvocationRepository constructs the repository.
func NewPostgresToolInvocationRepository(pool *pgxpool.Pool) *PostgresToolInvocationRepository {
	return &PostgresToolInvocationRepository{pool: pool}
}

var _ chatdomain.ToolInvocationRepository = (*PostgresToolInvocationRepository)(nil)

// Append inserts one tool invocation record.
func (r *PostgresToolInvocationRepository) Append(ctx context.Context, inv *entity.ToolInvocation) error {
	const q = `
		INSERT INTO chat_tool_invocations (
			id, session_id, message_id, tool_call_id, tool_name, mcp_server,
			arguments, raw_result, is_error, state, error, correlation_id, started_at, finished_at
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
	`
	args := inv.Arguments
	if len(args) == 0 {
		args = json.RawMessage(`{}`)
	}
	_, err := r.pool.Exec(ctx, q,
		inv.ID,
		inv.SessionID,
		inv.MessageID,
		inv.ToolCallID,
		inv.ToolName,
		inv.MCPServer,
		args,
		nullableJSON(inv.RawResult),
		inv.IsError,
		string(inv.State),
		nullableString(inv.Error),
		nullableString(inv.CorrelationID),
		inv.StartedAt,
		inv.FinishedAt,
	)
	if err != nil {
		return fmt.Errorf("insert chat_tool_invocation: %w", err)
	}
	return nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
