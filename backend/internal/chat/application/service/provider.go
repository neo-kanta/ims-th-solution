// Package service exposes the provider-agnostic chat contract.
//
// ChatProvider is the single seam every LLM backend implements. The agent
// loop in application/command/send_message.go is written against this
// interface and has no vendor knowledge; adapters under
// infrastructure/provider/* are responsible for translating their native
// streaming protocol — including tool calls — into the canonical types here.
package service

import (
	"context"
	"encoding/json"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// ToolSpec is one tool, in the canonical (provider-agnostic) shape. The agent
// loop sends the allowlisted MCP tools; each adapter translates these to its
// own function-calling schema.
type ToolSpec struct {
	Name        string
	Description string
	InputSchema json.RawMessage // JSON Schema (object)
}

// ToolCall is one tool-invocation request emitted by the model.
type ToolCall struct {
	ID    string
	Name  string
	Input json.RawMessage
}

// ToolResultMsg is the outcome of a tool call, fed back to the model on the
// next loop iteration.
type ToolResultMsg struct {
	ToolCallID string
	Content    string
	IsError    bool
}

// TextDelta is one incremental text fragment streamed to the user.
type TextDelta struct {
	Text string
}

// ConversationMessage is one provider-agnostic message in the running
// conversation. Plain user/assistant turns set Content. An assistant turn
// that requested tools sets ToolCalls (and possibly Content). A tool turn
// sets ToolResults. Adapters translate these into vendor message blocks.
type ConversationMessage struct {
	Role        valueobject.MessageRole
	Content     string
	ToolCalls   []ToolCall
	ToolResults []ToolResultMsg
}

// ChatRequest is the provider-agnostic input for one model turn.
//
// The backend owns conversation history and resends it each turn — providers
// are stateless from our perspective. Messages MUST already be trimmed by
// the history-cap policy before reaching the adapter.
type ChatRequest struct {
	SystemPrompt string
	Messages     []ConversationMessage
	Tools        []ToolSpec
	Model        string
	MaxTokens    int
	Temperature  float32
}

// ChatResult is the terminal value of a streamed response.
type ChatResult struct {
	ToolCalls          []ToolCall
	StopReason         valueobject.StopReason
	RawProviderPayload json.RawMessage // for audit; full accumulated message
}

// ChatStreamReader is one provider's incremental output. Consumers iterate
// with Next until it returns false, then inspect Result and Err.
type ChatStreamReader interface {
	Next(ctx context.Context) (TextDelta, bool)
	Result() ChatResult
	Err() error
	Close() error
}

// ChatProvider is the single seam every LLM backend implements.
type ChatProvider interface {
	Name() valueobject.ProviderID
	Generate(ctx context.Context, req ChatRequest) (ChatStreamReader, error)
}
