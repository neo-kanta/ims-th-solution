package anthropic

import "encoding/json"

// anthropicRequest is the Messages API request body.
type anthropicRequest struct {
	Model       string             `json:"model"`
	MaxTokens   int                `json:"max_tokens"`
	Messages    []anthropicMessage `json:"messages"`
	System      string             `json:"system,omitempty"`
	Temperature float32            `json:"temperature,omitempty"`
	Stream      bool               `json:"stream"`
	Tools       []anthropicTool    `json:"tools,omitempty"`
	ToolChoice  any                `json:"tool_choice,omitempty"`
}

// anthropicMessage carries either a plain string content or an array of
// content blocks (used for tool_use / tool_result). Content is typed as any
// so it marshals to whichever shape we set.
type anthropicMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// anthropicTool is the tool definition sent to the model.
type anthropicTool struct {
	Name        string          `json:"name"`
	Description string          `json:"description,omitempty"`
	InputSchema json.RawMessage `json:"input_schema"`
}

// Content blocks (request side).

type textBlock struct {
	Type string `json:"type"` // "text"
	Text string `json:"text"`
}

type toolUseBlock struct {
	Type  string          `json:"type"` // "tool_use"
	ID    string          `json:"id"`
	Name  string          `json:"name"`
	Input json.RawMessage `json:"input"`
}

type toolResultBlock struct {
	Type      string `json:"type"` // "tool_result"
	ToolUseID string `json:"tool_use_id"`
	Content   string `json:"content"`
	IsError   bool   `json:"is_error,omitempty"`
}

// ── Streaming event payloads (response side) ────────────────────────────────

// contentBlockStart announces a new content block. For tool_use blocks it
// carries the id and name; input arrives later as input_json_delta.
type contentBlockStart struct {
	Type         string `json:"type"`
	Index        int    `json:"index"`
	ContentBlock struct {
		Type string `json:"type"` // "text" | "tool_use"
		ID   string `json:"id"`
		Name string `json:"name"`
	} `json:"content_block"`
}

// contentBlockDelta carries either a text delta or a partial-JSON delta for a
// tool_use block's input.
type contentBlockDelta struct {
	Type  string `json:"type"`
	Index int    `json:"index"`
	Delta struct {
		Type        string `json:"type"`
		Text        string `json:"text"`
		PartialJSON string `json:"partial_json"`
	} `json:"delta"`
}

// messageDelta carries the stop_reason just before message_stop.
type messageDelta struct {
	Type  string `json:"type"`
	Delta struct {
		StopReason   string `json:"stop_reason"`
		StopSequence string `json:"stop_sequence"`
	} `json:"delta"`
}

// errorPayload is the payload for event: error.
type errorPayload struct {
	Type  string `json:"type"`
	Error struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error"`
}
