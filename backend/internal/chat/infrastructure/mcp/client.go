package mcp

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"os/exec"
	"sync"

	sdk "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
)

// authMetaKey MUST match the constant the ims-mcp server reads from _meta.
// The caller's JWT is forwarded here, out-of-band from the model.
const authMetaKey = "ims_auth_token"

// serverConn is a live connection to one MCP server.
type serverConn struct {
	cfg     ServerConfig
	session *sdk.ClientSession
	mu      sync.Mutex // serializes CallTool on this session
}

// Client connects to the configured MCP servers and implements
// service.ToolGateway. Tool names are assumed unique across servers; on a
// collision the first server wins and the duplicate is logged and skipped.
type Client struct {
	conns     []*serverConn
	toolIndex map[string]*serverConn // toolName → owning server
	logger    *slog.Logger
}

var _ service.ToolGateway = (*Client)(nil)

// Connect dials every server in cfg. Servers that fail to connect are logged
// and skipped so chat degrades to no-tools rather than failing outright.
func Connect(ctx context.Context, cfg *Config, logger *slog.Logger) (*Client, error) {
	if logger == nil {
		logger = slog.Default()
	}
	c := &Client{toolIndex: map[string]*serverConn{}, logger: logger}

	for _, sc := range cfg.Servers {
		session, err := dial(ctx, sc)
		if err != nil {
			logger.Warn("mcp: server connect failed; skipping", "server", sc.Name, "error", err)
			continue
		}
		conn := &serverConn{cfg: sc, session: session}
		c.conns = append(c.conns, conn)
		c.indexTools(ctx, conn)
	}
	return c, nil
}

// dial establishes a session for one server config.
func dial(ctx context.Context, sc ServerConfig) (*sdk.ClientSession, error) {
	client := sdk.NewClient(&sdk.Implementation{Name: "ims-chat", Version: "0.1.0"}, nil)

	switch sc.Transport {
	case TransportStdio:
		if sc.Command == "" {
			return nil, fmt.Errorf("stdio server %q has no command", sc.Name)
		}
		cmd := exec.CommandContext(ctx, sc.Command, sc.Args...)
		cmd.Env = os.Environ()
		for k, v := range sc.Env {
			cmd.Env = append(cmd.Env, k+"="+v)
		}
		return client.Connect(ctx, &sdk.CommandTransport{Command: cmd}, nil)
	case TransportHTTP:
		// Structure in place; the user selected stdio for this slice. Wiring
		// the Streamable HTTP transport here is a localized follow-up.
		return nil, fmt.Errorf("http transport not yet enabled for server %q", sc.Name)
	default:
		return nil, fmt.Errorf("server %q has unknown transport %q", sc.Name, sc.Transport)
	}
}

// indexTools records which server owns each allowlisted tool.
func (c *Client) indexTools(ctx context.Context, conn *serverConn) {
	res, err := conn.session.ListTools(ctx, nil)
	if err != nil {
		c.logger.Warn("mcp: ListTools failed", "server", conn.cfg.Name, "error", err)
		return
	}
	for _, t := range res.Tools {
		if !toolAllowed(conn.cfg, t.Name) {
			c.logger.Info("mcp: tool filtered by allow/deny policy", "server", conn.cfg.Name, "tool", t.Name)
			continue
		}
		if _, dup := c.toolIndex[t.Name]; dup {
			c.logger.Warn("mcp: duplicate tool name across servers; keeping first", "tool", t.Name, "server", conn.cfg.Name)
			continue
		}
		c.toolIndex[t.Name] = conn
	}
}

// ListTools returns the allowlisted tools across all servers as canonical
// ToolSpecs for provider translation.
func (c *Client) ListTools(ctx context.Context) ([]service.ToolSpec, error) {
	var specs []service.ToolSpec
	for _, conn := range c.conns {
		res, err := conn.session.ListTools(ctx, nil)
		if err != nil {
			c.logger.Warn("mcp: ListTools failed", "server", conn.cfg.Name, "error", err)
			continue
		}
		for _, t := range res.Tools {
			if !toolAllowed(conn.cfg, t.Name) {
				continue
			}
			schema := json.RawMessage(`{"type":"object"}`)
			if t.InputSchema != nil {
				if raw, err := json.Marshal(t.InputSchema); err == nil {
					schema = raw
				}
			}
			specs = append(specs, service.ToolSpec{
				Name:        t.Name,
				Description: t.Description,
				InputSchema: schema,
			})
		}
	}
	return specs, nil
}

// ExecuteTool routes a tool call to its owning server, forwarding the caller's
// JWT via _meta. A non-nil error means a transport failure; tool-level errors
// surface via ToolResult.IsError.
func (c *Client) ExecuteTool(ctx context.Context, call service.ToolCall, authToken string) (service.ToolResult, error) {
	conn, ok := c.toolIndex[call.Name]
	if !ok {
		return service.ToolResult{
			ToolName: call.Name,
			IsError:  true,
			Content:  fmt.Sprintf(`{"error":"unknown tool %q"}`, call.Name),
		}, nil
	}

	// Decode the model-supplied arguments. Empty → empty object.
	var args any
	if len(call.Input) > 0 {
		if err := json.Unmarshal(call.Input, &args); err != nil {
			return service.ToolResult{
				ToolName:   call.Name,
				ServerName: conn.cfg.Name,
				IsError:    true,
				Content:    `{"error":"tool arguments were not valid JSON"}`,
			}, nil
		}
	}

	params := &sdk.CallToolParams{
		Name:      call.Name,
		Arguments: args,
		Meta:      sdk.Meta{authMetaKey: authToken},
	}

	conn.mu.Lock()
	res, err := conn.session.CallTool(ctx, params)
	conn.mu.Unlock()
	if err != nil {
		return service.ToolResult{ToolName: call.Name, ServerName: conn.cfg.Name}, fmt.Errorf("call tool %q: %w", call.Name, err)
	}

	return service.ToolResult{
		ToolName:   call.Name,
		ServerName: conn.cfg.Name,
		Content:    extractText(res),
		IsError:    res.IsError,
	}, nil
}

// extractText concatenates the text content blocks of a tool result. Our IMS
// tools always return a single JSON text block.
func extractText(res *sdk.CallToolResult) string {
	if res == nil {
		return ""
	}
	var out string
	for _, ct := range res.Content {
		if tc, ok := ct.(*sdk.TextContent); ok {
			out += tc.Text
		}
	}
	return out
}

// Close shuts down every session (and its child process for stdio).
func (c *Client) Close() error {
	var firstErr error
	for _, conn := range c.conns {
		if err := conn.session.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
