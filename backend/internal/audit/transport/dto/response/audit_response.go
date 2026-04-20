package response

import "time"

// AuditEventResponse is the HTTP response for a single audit event.
type AuditEventResponse struct {
	ID         string                 `json:"id"`
	ActorID    *string                `json:"actor_id"`
	EventType  string                 `json:"event_type"`
	TargetType string                 `json:"target_type"`
	TargetID   string                 `json:"target_id"`
	IPAddress  string                 `json:"ip_address"`
	UserAgent  string                 `json:"user_agent"`
	Metadata   map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt  time.Time              `json:"created_at"`
}

// AuditListResponse is the paginated audit event list response.
type AuditListResponse struct {
	Events []AuditEventResponse `json:"events"`
	Total  int                  `json:"total"`
	Offset int                  `json:"offset"`
	Limit  int                  `json:"limit"`
}
