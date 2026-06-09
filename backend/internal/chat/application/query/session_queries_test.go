package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/types"
)

// ── Fakes ───────────────────────────────────────────────────────────────────

type fakeSessions struct {
	byID map[uuid.UUID]*entity.Session
}

func (f *fakeSessions) Create(context.Context, *entity.Session) error { return nil }
func (f *fakeSessions) FindByID(_ context.Context, id uuid.UUID) (*entity.Session, error) {
	return f.byID[id], nil
}
func (f *fakeSessions) Touch(context.Context, uuid.UUID) error                       { return nil }
func (f *fakeSessions) UpdateModel(context.Context, uuid.UUID, string, string) error { return nil }
func (f *fakeSessions) ListByUser(_ context.Context, userID uuid.UUID, _, _ int) ([]entity.Session, int, error) {
	var out []entity.Session
	for _, s := range f.byID {
		if s.UserID == userID {
			out = append(out, *s)
		}
	}
	return out, len(out), nil
}

type fakeMessages struct{ rows []entity.Message }

func (f *fakeMessages) Append(context.Context, *entity.Message) error { return nil }
func (f *fakeMessages) ListBySession(context.Context, uuid.UUID, int) ([]entity.Message, error) {
	return f.rows, nil
}
func (f *fakeMessages) ListBySessionPaged(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.Message, int, error) {
	return f.rows, len(f.rows), nil
}

// permGate returns a fixed allow/deny for HasFunctionPermission.
type permGate struct{ allow bool }

func (p permGate) HasFunctionPermission(context.Context, uuid.UUID, string) (bool, error) {
	return p.allow, nil
}

type recAudit struct {
	events []string
	fail   bool
}

func (a *recAudit) RecordTurnEvent(_ context.Context, eventType string, _ *uuid.UUID, _, _, _ string, _ map[string]interface{}) error {
	a.events = append(a.events, eventType)
	if a.fail {
		return errors.New("audit sink down")
	}
	return nil
}

func pag() types.Pagination { return types.Pagination{Page: 1, Limit: 20, Offset: 0} }

func newSession(owner uuid.UUID) *entity.Session {
	return &entity.Session{
		ID: uuid.New(), UserID: owner, Provider: valueobject.ProviderAnthropic,
		Model: "claude-haiku-4-5-20251001", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC(),
	}
}

// ── Tests ───────────────────────────────────────────────────────────────────

// Owner can read their own session; no audit event is emitted.
func TestGetSession_OwnerAllowed(t *testing.T) {
	owner := uuid.New()
	s := newSession(owner)
	audit := &recAudit{}
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, &fakeMessages{}, permGate{allow: false}, audit)

	got, err := q.GetSession(context.Background(), AccessContext{CallerID: owner}, s.ID)
	if err != nil {
		t.Fatalf("owner read should succeed: %v", err)
	}
	if got.ID != s.ID {
		t.Fatal("wrong session returned")
	}
	if len(audit.events) != 0 {
		t.Fatalf("owner read must not emit auditor audit, got %v", audit.events)
	}
}

// A non-owner WITHOUT IAM_AUDIT_VIEW is forbidden.
func TestGetSession_NonOwnerForbidden(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	s := newSession(owner)
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, &fakeMessages{}, permGate{allow: false}, &recAudit{})

	_, err := q.GetSession(context.Background(), AccessContext{CallerID: other}, s.ID)
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// A non-owner WITH IAM_AUDIT_VIEW may read, and a strict audit event fires.
func TestGetSession_AuditorAllowedAndAudited(t *testing.T) {
	owner := uuid.New()
	auditor := uuid.New()
	s := newSession(owner)
	audit := &recAudit{}
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, &fakeMessages{}, permGate{allow: true}, audit)

	_, err := q.GetSession(context.Background(), AccessContext{CallerID: auditor, CorrelationID: "corr-1"}, s.ID)
	if err != nil {
		t.Fatalf("auditor read should succeed: %v", err)
	}
	if len(audit.events) != 1 || audit.events[0] != AuditEventSessionViewed {
		t.Fatalf("expected one %s event, got %v", AuditEventSessionViewed, audit.events)
	}
}

// Auditor access fails closed if the strict audit write fails.
func TestGetSession_AuditorFailsClosedWhenAuditFails(t *testing.T) {
	owner := uuid.New()
	auditor := uuid.New()
	s := newSession(owner)
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, &fakeMessages{}, permGate{allow: true}, &recAudit{fail: true})

	if _, err := q.GetSession(context.Background(), AccessContext{CallerID: auditor}, s.ID); err == nil {
		t.Fatal("auditor read must fail when strict audit write fails")
	}
}

// Unknown session -> ErrNotFound.
func TestGetSession_NotFound(t *testing.T) {
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{}}, &fakeMessages{}, permGate{allow: true}, &recAudit{})
	if _, err := q.GetSession(context.Background(), AccessContext{CallerID: uuid.New()}, uuid.New()); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

// ListMessages enforces the same ownership rule and returns the transcript.
func TestListMessages_OwnerAllowed(t *testing.T) {
	owner := uuid.New()
	s := newSession(owner)
	msgs := &fakeMessages{rows: []entity.Message{
		{ID: uuid.New(), SessionID: s.ID, Role: valueobject.RoleUser, Content: "hi", CreatedAt: time.Now().UTC()},
	}}
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, msgs, permGate{allow: false}, &recAudit{})

	_, page, err := q.ListMessages(context.Background(), AccessContext{CallerID: owner}, s.ID, pag())
	if err != nil {
		t.Fatalf("owner read should succeed: %v", err)
	}
	if page.Total != 1 || len(page.Items) != 1 {
		t.Fatalf("expected 1 message, got total=%d items=%d", page.Total, len(page.Items))
	}
}

func TestListMessages_NonOwnerForbidden(t *testing.T) {
	owner := uuid.New()
	other := uuid.New()
	s := newSession(owner)
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, &fakeMessages{}, permGate{allow: false}, &recAudit{})

	if _, _, err := q.ListMessages(context.Background(), AccessContext{CallerID: other}, s.ID, pag()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden, got %v", err)
	}
}

// Listing your own sessions needs no special permission and no audit.
func TestListSessions_OwnNoAudit(t *testing.T) {
	owner := uuid.New()
	s := newSession(owner)
	audit := &recAudit{}
	q := NewSessionQueries(&fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}, &fakeMessages{}, permGate{allow: false}, audit)

	page, err := q.ListSessions(context.Background(), AccessContext{CallerID: owner}, uuid.Nil, pag())
	if err != nil {
		t.Fatalf("own list should succeed: %v", err)
	}
	if page.Total != 1 {
		t.Fatalf("expected 1 session, got %d", page.Total)
	}
	if len(audit.events) != 0 {
		t.Fatalf("own list must not audit, got %v", audit.events)
	}
}

// Listing another user's sessions requires IAM_AUDIT_VIEW and is audited.
func TestListSessions_OtherUserRequiresAuditor(t *testing.T) {
	caller := uuid.New()
	target := uuid.New()
	s := newSession(target)
	store := &fakeSessions{byID: map[uuid.UUID]*entity.Session{s.ID: s}}

	// Without permission -> forbidden.
	q1 := NewSessionQueries(store, &fakeMessages{}, permGate{allow: false}, &recAudit{})
	if _, err := q1.ListSessions(context.Background(), AccessContext{CallerID: caller}, target, pag()); !errors.Is(err, ErrForbidden) {
		t.Fatalf("expected ErrForbidden without IAM_AUDIT_VIEW, got %v", err)
	}

	// With permission -> allowed and audited.
	audit := &recAudit{}
	q2 := NewSessionQueries(store, &fakeMessages{}, permGate{allow: true}, audit)
	if _, err := q2.ListSessions(context.Background(), AccessContext{CallerID: caller}, target, pag()); err != nil {
		t.Fatalf("auditor list should succeed: %v", err)
	}
	if len(audit.events) != 1 || audit.events[0] != AuditEventSessionViewed {
		t.Fatalf("expected one %s event, got %v", AuditEventSessionViewed, audit.events)
	}
}
