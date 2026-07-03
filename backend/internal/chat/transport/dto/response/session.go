package response

import (
	"encoding/json"
	"time"
)

// SessionSummaryResponse is one chat session in a history listing. It carries
// no message content — only metadata.
type SessionSummaryResponse struct {
	ID        string    `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Provider  string    `json:"provider" example:"anthropic"`
	Model     string    `json:"model" example:"claude-haiku-4-5-20251001"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// SessionListResponse is a paginated list of session summaries.
type SessionListResponse struct {
	Items      []SessionSummaryResponse `json:"items"`
	Total      int64                    `json:"total" example:"42"`
	Page       int                      `json:"page" example:"1"`
	Limit      int                      `json:"limit" example:"20"`
	TotalPages int                      `json:"total_pages" example:"3"`
}

// ChatMessageResponse is one persisted message in a session transcript.
// `provenance` is the safe provenance map (tool-call pointers + numeric
// validation summary). The raw provider payload is NEVER included.
type ChatMessageResponse struct {
	ID         string          `json:"id"`
	Role       string          `json:"role" example:"assistant"`
	Content    string          `json:"content"`
	Provenance json.RawMessage `json:"provenance,omitempty" swaggertype:"object"`
	CreatedAt  time.Time       `json:"created_at"`
}

// SessionMessagesResponse is a paginated transcript for one session.
type SessionMessagesResponse struct {
	SessionID  string                `json:"session_id"`
	Items      []ChatMessageResponse `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	Limit      int                   `json:"limit"`
	TotalPages int                   `json:"total_pages"`
}
