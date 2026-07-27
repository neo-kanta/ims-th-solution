package response

import (
	"time"

	"github.com/neo-kanta/ims-th-solution/backend/internal/approval/domain/entity"
)

// SyncFailureResponse is a durable, operator-visible subject-sync replay
// record.
type SyncFailureResponse struct {
	ID                string     `json:"id"`
	ApprovalRequestID string     `json:"approval_request_id"`
	SubjectType       string     `json:"subject_type"`
	SubjectID         string     `json:"subject_id"`
	Outcome           string     `json:"outcome"`
	Reason            string     `json:"reason,omitempty"`
	AttemptCount      int        `json:"attempt_count"`
	MaxAttempts       int        `json:"max_attempts"`
	LastError         string     `json:"last_error"`
	Status            string     `json:"status"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`
	ResolvedAt        *time.Time `json:"resolved_at,omitempty"`
	ResolvedBy        string     `json:"resolved_by,omitempty"`
}

// SyncFailureListResponse is the paginated operator inbox.
type SyncFailureListResponse struct {
	Items []SyncFailureResponse `json:"items"`
	Total int                   `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
}

// FromSyncFailure maps a domain sync-failure record to its response DTO.
func FromSyncFailure(f *entity.SyncFailure) SyncFailureResponse {
	out := SyncFailureResponse{
		ID:                f.ID.String(),
		ApprovalRequestID: f.ApprovalRequestID.String(),
		SubjectType:       string(f.SubjectType),
		SubjectID:         f.SubjectID.String(),
		Outcome:           string(f.Outcome),
		Reason:            f.Reason,
		AttemptCount:      f.AttemptCount,
		MaxAttempts:       f.MaxAttempts,
		LastError:         f.LastError,
		Status:            string(f.Status),
		CreatedAt:         f.CreatedAt,
		UpdatedAt:         f.UpdatedAt,
		ResolvedAt:        f.ResolvedAt,
	}
	if f.ResolvedBy != nil {
		out.ResolvedBy = f.ResolvedBy.String()
	}
	return out
}

// FromSyncFailures maps a slice of sync-failure records.
func FromSyncFailures(fs []*entity.SyncFailure) []SyncFailureResponse {
	out := make([]SyncFailureResponse, 0, len(fs))
	for _, f := range fs {
		out = append(out, FromSyncFailure(f))
	}
	return out
}
