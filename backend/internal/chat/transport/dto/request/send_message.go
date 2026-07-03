// Package request holds chat HTTP request DTOs.
package request

// SendMessageRequest is the body of POST /api/v1/chat.
//
// SessionID is optional — omit it to start a new conversation. Content is
// required. The server enforces a 16 KiB content cap to keep audit rows
// reasonable; longer inputs should be uploaded as attachments (Slice D).
type SendMessageRequest struct {
	// SessionID is the existing session UUID. Empty starts a new session.
	SessionID string `json:"session_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`

	// Content is the user's message. Required.
	Content string `json:"content" validate:"required,min=1,max=16384" example:"Summarize last quarter's NAV trend for Fund A."`

	// Model is the specific LLM model ID. Optional.
	Model string `json:"model,omitempty" example:"claude-haiku-4-5-20251001"`

	// Provider is the LLM provider ID. Optional.
	Provider string `json:"provider,omitempty" example:"anthropic"`
}
