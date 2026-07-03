// Package response holds chat HTTP response DTOs.
//
// The chat endpoint responds with Content-Type: text/event-stream. Each SSE
// frame uses an `event:` line for the discriminator and a `data:` line
// containing one ChatStreamEvent JSON object. The frontend reads the stream
// from fetch().body and parses each frame.
//
// The Kind discriminator matches the SSE event name verbatim. Optional
// fields populate per Kind:
//
//	session_started → SessionID, MessageID (user message id)
//	text            → Text
//	done            → SessionID, MessageID (assistant message id), StopReason
//	error           → SessionID, Error
package response

// ChatStreamEvent is one SSE frame's `data:` payload. The same envelope is
// reused for every event kind so the frontend has exactly one type to parse.
type ChatStreamEvent struct {
	// Kind is the SSE event name. One of: session_started, text, done, error.
	Kind string `json:"kind" example:"text"`

	// SessionID is the conversation id. Present on session_started, done, and error.
	SessionID string `json:"session_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174000"`

	// MessageID is the user message id on session_started, or the assistant
	// message id on done.
	MessageID string `json:"message_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174001"`

	// Text is an incremental token. Present only on kind=text.
	Text string `json:"text,omitempty" example:"Hello"`

	// ToolName is the MCP tool being called. Present only on kind=tool_call.
	ToolName string `json:"tool_name,omitempty" example:"get_portfolio_holdings"`

	// ToolStatus is the tool-call lifecycle marker. Present only on
	// kind=tool_call. One of: running, ok, error, denied.
	ToolStatus string `json:"tool_status,omitempty" example:"running"`

	// StopReason is the canonical termination reason. Present only on kind=done.
	// One of: end_turn, tool_use, max_tokens, stop_sequence, error, unknown.
	StopReason string `json:"stop_reason,omitempty" example:"end_turn"`

	// Error is a safe, user-facing error message. Present only on kind=error.
	Error string `json:"error,omitempty" example:"the assistant could not complete this turn"`

	// Sources lists the turn-level provenance (which tools ran). Present only on
	// kind=sources.
	Sources []ChatSourceRef `json:"sources,omitempty"`

	// Bindings links cited figures to the tool that produced them. Present only
	// on kind=sources, and only when the answer passed numeric validation.
	Bindings []ChatFigureBinding `json:"bindings,omitempty"`

	// Unverified lists the figures that failed numeric validation. Present only
	// on kind=validation with tool_status=blocked.
	Unverified []string `json:"unverified,omitempty"`
}

// ChatSourceRef is one tool that ran during the turn. It carries no secrets.
type ChatSourceRef struct {
	ToolName string `json:"tool_name" example:"get_fund_nav"`
	Server   string `json:"server,omitempty" example:"ims"`
	State    string `json:"state" example:"succeeded"`
	AsOf     string `json:"as_of,omitempty" example:"2026-06-07T08:30:00Z"`
}

// ChatFigureBinding links one figure in the answer to its grounding tool.
type ChatFigureBinding struct {
	Figure   string `json:"figure" example:"10.25"`
	Source   string `json:"source" example:"tool_raw"`
	ToolName string `json:"tool_name,omitempty" example:"get_fund_nav"`
}
