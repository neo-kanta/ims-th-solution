package service

import "context"

// ToolResult is the canonical outcome of one tool execution, independent of
// MCP. Content is the verbatim text the tool returned (for our IMS tools, a
// JSON envelope with provenance + raw upstream data). The agent loop feeds
// Content back to the model unchanged and persists it for provenance.
type ToolResult struct {
	ToolName   string
	ServerName string
	Content    string // verbatim tool output (JSON text)
	IsError    bool
}

// ToolGateway is the chat module's port onto MCP. The agent loop depends on
// this interface only — it has no knowledge of the MCP SDK, transports, or
// which server backs a given tool.
//
// authToken is the end user's JWT, forwarded to the MCP server out-of-band
// (via MCP _meta) so the server's REST calls run with the user's own
// permissions. It is NEVER exposed to the model and MUST be redacted from
// audit metadata.
type ToolGateway interface {
	// ListTools returns the allowlisted tools across all connected MCP
	// servers, already in canonical ToolSpec form for provider translation.
	ListTools(ctx context.Context) ([]ToolSpec, error)

	// ExecuteTool runs one tool call and returns its verbatim result. A
	// returned error is a transport/infra failure; a tool-level failure is
	// reported via ToolResult.IsError with a safe message in Content.
	ExecuteTool(ctx context.Context, call ToolCall, authToken string) (ToolResult, error)

	// Close shuts down MCP sessions / child processes.
	Close() error
}

// NoopToolGateway is used when MCP is disabled. It advertises no tools and
// refuses execution, so the agent loop runs exactly like Slice A.
type NoopToolGateway struct{}

func (NoopToolGateway) ListTools(context.Context) ([]ToolSpec, error) { return nil, nil }

func (NoopToolGateway) ExecuteTool(context.Context, ToolCall, string) (ToolResult, error) {
	return ToolResult{IsError: true, Content: `{"error":"no tools are available"}`}, nil
}

func (NoopToolGateway) Close() error { return nil }
