package metrics

import "github.com/prometheus/client_golang/prometheus"

// Outcome label values for chat_turn_total.
const (
	ChatTurnCompleted = "completed"
	ChatTurnBlocked   = "blocked"
	ChatTurnFailed    = "failed"
)

// Tool status label values for chat_tool_call_total.
const (
	ChatToolOK     = "ok"
	ChatToolError  = "error"
	ChatToolDenied = "denied"
)

// Chat metric handles. Call sites:
//
//	metrics.ChatTurnTotal.WithLabelValues(metrics.ChatTurnCompleted).Inc()
//	metrics.ChatToolCallTotal.WithLabelValues("get_fund_nav", metrics.ChatToolOK).Inc()
//	metrics.ChatProviderLatencySecs.Observe(elapsed.Seconds())
var (
	ChatTurnTotal               *prometheus.CounterVec
	ChatToolCallTotal           *prometheus.CounterVec
	ChatToolDeniedTotal         *prometheus.CounterVec
	ChatNumericValidationFailed prometheus.Counter
	ChatProviderLatencySecs     prometheus.Histogram
	ChatMCPLatencySecs          prometheus.Histogram
)

func init() { registerChatMetrics() }

// registerChatMetrics constructs every chat metric and registers it with the
// IMS registry. Called from init() and from resetAllForTest().
func registerChatMetrics() {
	ChatTurnTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_turn_total",
			Help: "Number of chat turns by outcome (completed|blocked|failed).",
		},
		[]string{"outcome"},
	)
	ChatToolCallTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_tool_call_total",
			Help: "Number of chat MCP tool calls by tool and status (ok|error|denied).",
		},
		[]string{"tool", "status"},
	)
	ChatToolDeniedTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "chat_tool_denied_total",
			Help: "Number of chat MCP tool calls denied by the permission/write gate, by tool.",
		},
		[]string{"tool"},
	)
	ChatNumericValidationFailed = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "chat_numeric_validation_failed_total",
			Help: "Number of chat turns blocked because a financial figure could not be verified.",
		},
	)
	ChatProviderLatencySecs = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "chat_provider_latency_seconds",
			Help:    "Latency of LLM provider Generate calls (time to response start).",
			Buckets: prometheus.ExponentialBuckets(0.05, 2, 10),
		},
	)
	ChatMCPLatencySecs = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "chat_mcp_latency_seconds",
			Help:    "Latency of chat MCP tool executions.",
			Buckets: prometheus.ExponentialBuckets(0.01, 2, 12),
		},
	)

	mustRegister(ChatTurnTotal)
	mustRegister(ChatToolCallTotal)
	mustRegister(ChatToolDeniedTotal)
	mustRegister(ChatNumericValidationFailed)
	mustRegister(ChatProviderLatencySecs)
	mustRegister(ChatMCPLatencySecs)
}
