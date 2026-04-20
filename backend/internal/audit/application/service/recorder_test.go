package service

import (
	"context"
	"testing"

	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain/entity"
)

type recorderTestRepo struct {
	recorded []*entity.AuditEvent
}

func (r *recorderTestRepo) Record(_ context.Context, event *entity.AuditEvent) error {
	r.recorded = append(r.recorded, event)
	return nil
}

func (r *recorderTestRepo) List(context.Context, domain.AuditFilter) ([]entity.AuditEvent, int, error) {
	return nil, 0, nil
}

func TestRecorderAddsRequestIDWithoutMutatingInputMetadata(t *testing.T) {
	repo := &recorderTestRepo{}
	recorder := NewRecorder(repo)

	actorID := uuid.New()
	metadata := map[string]interface{}{"reason": "test"}
	ctx := context.WithValue(context.Background(), chimw.RequestIDKey, "req-123")

	recorder.Record(ctx, &actorID, entity.AuditLoginSuccess, "user", actorID.String(), "127.0.0.1", "test-agent", metadata)

	require.Len(t, repo.recorded, 1)
	require.Equal(t, "req-123", repo.recorded[0].Metadata["request_id"])
	require.Equal(t, "test", repo.recorded[0].Metadata["reason"])
	_, exists := metadata["request_id"]
	require.False(t, exists)
}

func TestRecorderCreatesMetadataWhenOnlyRequestIDIsPresent(t *testing.T) {
	repo := &recorderTestRepo{}
	recorder := NewRecorder(repo)

	ctx := context.WithValue(context.Background(), chimw.RequestIDKey, "req-456")
	recorder.Record(ctx, nil, entity.AuditLogout, "user", "", "", "", nil)

	require.Len(t, repo.recorded, 1)
	require.Equal(t, "req-456", repo.recorded[0].Metadata["request_id"])
}
