// Package chat is the modular monolith's chat module. It owns conversation
// sessions, messages, the provider-agnostic agent loop, the LLM-backend
// adapters, and the MCP client through which ALL business data is reached.
//
// Cross-module integrations are deliberately narrow:
//   - audit/domain.Recorder for the append-only audit trail (correction 2).
//   - a structural IAM permission port for the pre-execution permission gate.
//   - platform/middleware for auth context.
//
// chat MUST NOT import any other internal/<module>/ package directly, and it
// reaches all financial data ONLY via MCP (never an in-process investment
// import). The boundary is enforced by backend/internal/chat/boundary_test.go.
package chat

import (
	"context"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	auditdomain "github.com/neo-kanta/ims-th-solution/backend/internal/audit/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/command"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/query"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/infrastructure/adapter"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/infrastructure/mcp"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/infrastructure/persistence"
	chatprovider "github.com/neo-kanta/ims-th-solution/backend/internal/chat/infrastructure/provider"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/transport"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/transport/handler"
)

// Config is the chat module's external configuration. main.go fills this from
// platform/config so the module doesn't depend on the AppConfig shape.
type Config struct {
	Provider         chatprovider.Config
	MaxTokensPerTurn int

	// WriteEnabled gates mutating tools. Off by default — the assistant is
	// read-only unless explicitly enabled AND the user passes the permission
	// gate.
	WriteEnabled bool

	// MCP wiring.
	MCPConfigPath string // optional mcp-servers.yaml; empty → default single IMS server
	MCPIMSBinPath string // path to the ims-mcp binary for the default config
	IMSAPIBaseURL string // forwarded to ims-mcp as IMS_API_BASE_URL
}

// IAMPermissionPort is the structural slice of IAM the chat module needs for
// its permission gate. *iam.Module satisfies it, so main.go passes the module
// directly without chat importing iam's internal package.
type IAMPermissionPort interface {
	HasFunctionPermission(ctx context.Context, userID string, code string) (bool, error)
}

// Module bundles chat's persistence, command, transport, and MCP client.
type Module struct {
	handler        *handler.ChatHandler
	sessionHandler *handler.SessionHandler
	mcpClient      *mcp.Client
}

// NewModule constructs and wires the chat module. Returns an error only if the
// configured LLM provider fails to initialize (fail fast). MCP connection
// failures are non-fatal: the module degrades to no-tools and logs a warning.
func NewModule(pool *pgxpool.Pool, cfg Config, auditRecorder auditdomain.Recorder, iamPort IAMPermissionPort) (*Module, error) {
	provider, err := chatprovider.Build(cfg.Provider)
	if err != nil {
		return nil, err
	}

	sessionRepo := persistence.NewPostgresSessionRepository(pool)
	messageRepo := persistence.NewPostgresMessageRepository(pool)
	invocationRepo := persistence.NewPostgresToolInvocationRepository(pool)
	audit := adapter.NewRecorderAuditWriter(auditRecorder)

	var permGate service.PermissionGate = service.AllowAllPermissionGate{}
	if iamPort != nil {
		permGate = adapter.NewPermissionGate(iamPort)
	}

	// Connect MCP. A missing/empty config falls back to the bundled IMS
	// stdio server. Connection failures degrade to no-tools.
	var gateway service.ToolGateway = service.NoopToolGateway{}
	var mcpClient *mcp.Client
	mcpCfg, cfgErr := mcp.LoadConfig(cfg.MCPConfigPath)
	if cfgErr != nil {
		slog.Warn("chat: failed to load MCP config; tools disabled", "error", cfgErr)
	} else {
		if len(mcpCfg.Servers) == 0 {
			mcpCfg = mcp.DefaultIMSConfig(cfg.MCPIMSBinPath, cfg.IMSAPIBaseURL)
		}
		client, connErr := mcp.Connect(context.Background(), mcpCfg, slog.Default())
		if connErr != nil {
			slog.Warn("chat: MCP connect failed; tools disabled", "error", connErr)
		} else {
			mcpClient = client
			gateway = client
		}
	}

	cmd := command.NewSendMessage(
		sessionRepo, messageRepo, invocationRepo,
		provider, gateway, permGate, audit,
		cfg.MaxTokensPerTurn, cfg.WriteEnabled,
	)
	h := handler.NewChatHandler(cmd)

	// Session-history reads (owner + IAM_AUDIT_VIEW auditor, strictly audited).
	sessionQueries := query.NewSessionQueries(sessionRepo, messageRepo, permGate, audit)
	sh := handler.NewSessionHandler(sessionQueries)

	return &Module{handler: h, sessionHandler: sh, mcpClient: mcpClient}, nil
}

// Health is a non-secret snapshot of chat/MCP readiness for the health probe.
type Health struct {
	Enabled      bool `json:"enabled"`       // the chat module is mounted
	MCPConnected bool `json:"mcp_connected"` // at least one MCP server connected
	ToolCount    int  `json:"tool_count"`    // allowlisted tools available
}

// HealthSnapshot reports chat/MCP readiness without leaking any secret
// (no API keys, tokens, model ids, or endpoints). Safe to call on a nil module.
func (m *Module) HealthSnapshot(ctx context.Context) Health {
	if m == nil || m.handler == nil {
		return Health{Enabled: false}
	}
	h := Health{Enabled: true}
	if m.mcpClient != nil {
		if specs, err := m.mcpClient.ListTools(ctx); err == nil {
			h.MCPConnected = true
			h.ToolCount = len(specs)
		}
	}
	return h
}

// RegisterRoutes mounts /chat on the given authenticated router.
func (m *Module) RegisterRoutes(r chi.Router) {
	if m == nil || m.handler == nil {
		return
	}
	transport.RegisterRoutes(r, m.handler, m.sessionHandler)
}

// Close shuts down MCP sessions and their child processes. Safe to call on a
// nil module or when MCP never connected.
func (m *Module) Close() error {
	if m == nil || m.mcpClient == nil {
		return nil
	}
	return m.mcpClient.Close()
}
