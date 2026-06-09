// Package query holds chat read use cases (session history). Authorization is
// enforced here, not in transport: a caller may read their OWN sessions; a
// caller holding IAM_AUDIT_VIEW may read ANY session for audit/investigation,
// and every such auditor read emits a STRICT audit event (fail-closed: if the
// audit write fails, the read fails).
package query

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	chatdomain "github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/types"
)

const (
	// AuditEventSessionViewed is emitted when an auditor (non-owner with
	// IAM_AUDIT_VIEW) reads another user's chat session or messages.
	AuditEventSessionViewed = "CHAT_SESSION_VIEWED"
	// PermissionAuditView is the existing IAM permission that grants cross-user
	// read access for audit/investigation.
	PermissionAuditView = "IAM_AUDIT_VIEW"
)

// ErrForbidden / ErrNotFound are sentinel errors the transport maps to 403/404.
var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("session not found")
)

// AccessContext carries the authenticated caller and request metadata needed
// for authorization and the auditor audit trail.
type AccessContext struct {
	CallerID      uuid.UUID
	IPAddress     string
	UserAgent     string
	CorrelationID string
}

// SessionQueries serves session-history reads.
type SessionQueries struct {
	sessions chatdomain.SessionRepository
	messages chatdomain.MessageRepository
	perms    service.PermissionGate
	audit    service.AuditWriter
}

// NewSessionQueries wires the read use cases.
func NewSessionQueries(
	sessions chatdomain.SessionRepository,
	messages chatdomain.MessageRepository,
	perms service.PermissionGate,
	audit service.AuditWriter,
) *SessionQueries {
	return &SessionQueries{sessions: sessions, messages: messages, perms: perms, audit: audit}
}

// ListSessions returns one page of sessions. By default the caller's own
// sessions; if targetUserID is set and differs from the caller, IAM_AUDIT_VIEW
// is required and the access is strictly audited.
func (q *SessionQueries) ListSessions(ctx context.Context, caller AccessContext, targetUserID uuid.UUID, pag types.Pagination) (types.PagedResult[entity.Session], error) {
	owner := targetUserID
	if owner == uuid.Nil {
		owner = caller.CallerID
	}
	if owner != caller.CallerID {
		allowed, err := q.perms.HasFunctionPermission(ctx, caller.CallerID, PermissionAuditView)
		if err != nil {
			return types.PagedResult[entity.Session]{}, fmt.Errorf("permission check: %w", err)
		}
		if !allowed {
			return types.PagedResult[entity.Session]{}, ErrForbidden
		}
		if err := q.auditAccess(ctx, caller, uuid.Nil, owner, "list_sessions"); err != nil {
			return types.PagedResult[entity.Session]{}, err
		}
	}

	items, total, err := q.sessions.ListByUser(ctx, owner, pag.Limit, pag.Offset)
	if err != nil {
		return types.PagedResult[entity.Session]{}, err
	}
	return types.NewPagedResult(items, int64(total), pag.Page, pag.Limit), nil
}

// GetSession returns one session if the caller may read it.
func (q *SessionQueries) GetSession(ctx context.Context, caller AccessContext, sessionID uuid.UUID) (*entity.Session, error) {
	s, err := q.loadAuthorized(ctx, caller, sessionID, "get_session")
	if err != nil {
		return nil, err
	}
	return s, nil
}

// ListMessages returns one page of a session's messages if the caller may read
// it. raw_provider_payload is never loaded by the underlying repository.
func (q *SessionQueries) ListMessages(ctx context.Context, caller AccessContext, sessionID uuid.UUID, pag types.Pagination) (*entity.Session, types.PagedResult[entity.Message], error) {
	s, err := q.loadAuthorized(ctx, caller, sessionID, "list_messages")
	if err != nil {
		return nil, types.PagedResult[entity.Message]{}, err
	}
	items, total, err := q.messages.ListBySessionPaged(ctx, sessionID, pag.Limit, pag.Offset)
	if err != nil {
		return nil, types.PagedResult[entity.Message]{}, err
	}
	return s, types.NewPagedResult(items, int64(total), pag.Page, pag.Limit), nil
}

// loadAuthorized loads a session and enforces owner-or-auditor access, emitting
// a strict audit event for auditor reads.
func (q *SessionQueries) loadAuthorized(ctx context.Context, caller AccessContext, sessionID uuid.UUID, action string) (*entity.Session, error) {
	s, err := q.sessions.FindByID(ctx, sessionID)
	if err != nil {
		return nil, err
	}
	if s == nil {
		return nil, ErrNotFound
	}
	if s.UserID == caller.CallerID {
		return s, nil // owner
	}
	allowed, err := q.perms.HasFunctionPermission(ctx, caller.CallerID, PermissionAuditView)
	if err != nil {
		return nil, fmt.Errorf("permission check: %w", err)
	}
	if !allowed {
		return nil, ErrForbidden
	}
	if err := q.auditAccess(ctx, caller, s.ID, s.UserID, action); err != nil {
		return nil, err
	}
	return s, nil
}

// auditAccess writes the strict auditor-access event. Returns an error if the
// audit write fails so the caller can fail the read closed.
func (q *SessionQueries) auditAccess(ctx context.Context, caller AccessContext, targetSessionID, targetOwnerID uuid.UUID, action string) error {
	meta := map[string]interface{}{
		"access":          "auditor",
		"action":          action,
		"target_owner_id": targetOwnerID.String(),
		"correlation_id":  caller.CorrelationID,
	}
	target := caller.CallerID.String()
	if targetSessionID != uuid.Nil {
		meta["target_session_id"] = targetSessionID.String()
		target = targetSessionID.String()
	}
	if err := q.audit.RecordTurnEvent(ctx, AuditEventSessionViewed, &caller.CallerID, target, caller.IPAddress, caller.UserAgent, meta); err != nil {
		return fmt.Errorf("audit auditor session access: %w", err)
	}
	return nil
}
