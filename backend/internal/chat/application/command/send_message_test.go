package command

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/application/service"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/chat/domain/valueobject"
)

// ── Fakes ───────────────────────────────────────────────────────────────────

type fakeSessionRepo struct{ created *entity.Session }

func (f *fakeSessionRepo) Create(_ context.Context, s *entity.Session) error {
	f.created = s
	return nil
}
func (f *fakeSessionRepo) FindByID(_ context.Context, _ uuid.UUID) (*entity.Session, error) {
	return nil, nil
}
func (f *fakeSessionRepo) Touch(_ context.Context, _ uuid.UUID) error { return nil }
func (f *fakeSessionRepo) UpdateModel(_ context.Context, _ uuid.UUID, _, _ string) error {
	return nil
}
func (f *fakeSessionRepo) ListByUser(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.Session, int, error) {
	return nil, 0, nil
}

type fakeMessageRepo struct{ appended []*entity.Message }

func (f *fakeMessageRepo) Append(_ context.Context, m *entity.Message) error {
	f.appended = append(f.appended, m)
	return nil
}
func (f *fakeMessageRepo) ListBySession(_ context.Context, _ uuid.UUID, _ int) ([]entity.Message, error) {
	out := make([]entity.Message, 0, len(f.appended))
	for _, m := range f.appended {
		out = append(out, *m)
	}
	return out, nil
}
func (f *fakeMessageRepo) ListBySessionPaged(_ context.Context, _ uuid.UUID, _, _ int) ([]entity.Message, int, error) {
	out := make([]entity.Message, 0, len(f.appended))
	for _, m := range f.appended {
		out = append(out, *m)
	}
	return out, len(out), nil
}

type fakeInvocationRepo struct{ rows []*entity.ToolInvocation }

func (f *fakeInvocationRepo) Append(_ context.Context, inv *entity.ToolInvocation) error {
	f.rows = append(f.rows, inv)
	return nil
}

type fakeAudit struct{ events []string }

func (f *fakeAudit) RecordTurnEvent(_ context.Context, eventType string, _ *uuid.UUID, _, _, _ string, _ map[string]interface{}) error {
	f.events = append(f.events, eventType)
	return nil
}

// scriptedProvider returns a queued sequence of stream results, one per
// Generate call — letting us simulate "model asks for a tool, then answers".
type scriptedProvider struct {
	steps []scriptedStep
	call  int
}

type scriptedStep struct {
	text      string
	toolCalls []service.ToolCall
	stop      valueobject.StopReason
}

func (p *scriptedProvider) Name() valueobject.ProviderID { return valueobject.ProviderAnthropic }
func (p *scriptedProvider) Generate(_ context.Context, _ service.ChatRequest) (service.ChatStreamReader, error) {
	if p.call >= len(p.steps) {
		return &scriptedStream{step: scriptedStep{stop: valueobject.StopReasonEndTurn}}, nil
	}
	s := p.steps[p.call]
	p.call++
	return &scriptedStream{step: s}, nil
}

type scriptedStream struct {
	step scriptedStep
	emit bool
}

func (s *scriptedStream) Next(_ context.Context) (service.TextDelta, bool) {
	if !s.emit && s.step.text != "" {
		s.emit = true
		return service.TextDelta{Text: s.step.text}, true
	}
	return service.TextDelta{}, false
}
func (s *scriptedStream) Result() service.ChatResult {
	return service.ChatResult{ToolCalls: s.step.toolCalls, StopReason: s.step.stop}
}
func (s *scriptedStream) Err() error   { return nil }
func (s *scriptedStream) Close() error { return nil }

// recordingGateway records executed calls and returns a canned result.
type recordingGateway struct {
	specs    []service.ToolSpec
	executed []service.ToolCall
	result   service.ToolResult
}

func (g *recordingGateway) ListTools(context.Context) ([]service.ToolSpec, error) {
	return g.specs, nil
}
func (g *recordingGateway) ExecuteTool(_ context.Context, call service.ToolCall, _ string) (service.ToolResult, error) {
	g.executed = append(g.executed, call)
	return g.result, nil
}
func (g *recordingGateway) Close() error { return nil }

type denyGate struct{}

func (denyGate) HasFunctionPermission(context.Context, uuid.UUID, string) (bool, error) {
	return false, nil
}

// drainText consumes the event channel and returns the concatenated assistant
// text plus whether a done event was seen and which tool statuses streamed.
func drainEvents(ch <-chan StreamEvent) (text string, done bool, toolStatuses []string) {
	for ev := range ch {
		switch ev.Kind {
		case StreamText:
			text += ev.Text
		case StreamToolCall:
			toolStatuses = append(toolStatuses, ev.ToolStatus)
		case StreamDone:
			done = true
		}
	}
	return
}

// ── Tests ────────────────────────────────────────────────────────────────────

// TestAgentLoopExecutesToolThenAnswers verifies the core loop: the model asks
// for a tool, the loop executes it via the gateway, feeds the result back, and
// the model's second turn is the final answer. The tool result must be
// persisted as provenance.
func TestAgentLoopExecutesToolThenAnswers(t *testing.T) {
	gw := &recordingGateway{
		specs:  []service.ToolSpec{{Name: "list_portfolios"}},
		result: service.ToolResult{ToolName: "list_portfolios", ServerName: "ims", Content: `{"data":[{"id":"p1"}]}`},
	}
	prov := &scriptedProvider{steps: []scriptedStep{
		{text: "Let me check.", toolCalls: []service.ToolCall{{ID: "t1", Name: "list_portfolios", Input: json.RawMessage(`{}`)}}, stop: valueobject.StopReasonToolUse},
		{text: "You have 1 portfolio.", stop: valueobject.StopReasonEndTurn},
	}}
	msgs := &fakeMessageRepo{}
	invs := &fakeInvocationRepo{}

	cmd := NewSendMessage(&fakeSessionRepo{}, msgs, invs, prov, gw, service.AllowAllPermissionGate{}, &fakeAudit{}, 1024, false)

	ch, err := cmd.Run(context.Background(), SendMessageInput{UserID: uuid.New(), Content: "list my portfolios"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	text, done, statuses := drainEvents(ch)

	if !done {
		t.Fatal("expected a done event")
	}
	if len(gw.executed) != 1 || gw.executed[0].Name != "list_portfolios" {
		t.Fatalf("expected list_portfolios executed once, got %#v", gw.executed)
	}
	if want := "You have 1 portfolio."; !contains(text, want) {
		t.Fatalf("final text %q missing %q", text, want)
	}
	if len(statuses) == 0 || statuses[len(statuses)-1] != "ok" {
		t.Fatalf("expected an 'ok' tool status, got %#v", statuses)
	}
	// Provenance: the tool invocation must be persisted with the raw result.
	if len(invs.rows) != 1 {
		t.Fatalf("expected 1 tool invocation persisted, got %d", len(invs.rows))
	}
	if invs.rows[0].State != valueobject.ToolStateSucceeded {
		t.Fatalf("expected succeeded state, got %s", invs.rows[0].State)
	}
	if string(invs.rows[0].RawResult) == "" {
		t.Fatal("expected raw result persisted as provenance")
	}
	if invs.rows[0].MessageID == nil {
		t.Fatal("expected tool invocation linked to assistant message")
	}
}

// TestPermissionGateBlocksTool verifies that a denied permission stops the tool
// from executing and records a denied invocation.
func TestPermissionGateBlocksTool(t *testing.T) {
	gw := &recordingGateway{
		specs:  []service.ToolSpec{{Name: "get_portfolio_holdings"}},
		result: service.ToolResult{Content: `{"data":{}}`},
	}
	prov := &scriptedProvider{steps: []scriptedStep{
		{toolCalls: []service.ToolCall{{ID: "t1", Name: "get_portfolio_holdings", Input: json.RawMessage(`{"portfolio_id":"p1"}`)}}, stop: valueobject.StopReasonToolUse},
		{text: "I don't have access to that.", stop: valueobject.StopReasonEndTurn},
	}}
	invs := &fakeInvocationRepo{}

	cmd := NewSendMessage(&fakeSessionRepo{}, &fakeMessageRepo{}, invs, prov, gw, denyGate{}, &fakeAudit{}, 1024, false)

	ch, err := cmd.Run(context.Background(), SendMessageInput{UserID: uuid.New(), Content: "show holdings"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	_, done, statuses := drainEvents(ch)

	if !done {
		t.Fatal("expected a done event")
	}
	if len(gw.executed) != 0 {
		t.Fatalf("permission-denied tool must NOT execute, but it ran: %#v", gw.executed)
	}
	if len(invs.rows) != 1 || invs.rows[0].State != valueobject.ToolStateDenied {
		t.Fatalf("expected one denied invocation, got %#v", invs.rows)
	}
	if len(statuses) == 0 || statuses[len(statuses)-1] != "denied" {
		t.Fatalf("expected a 'denied' tool status, got %#v", statuses)
	}
}

// TestWriteGateBlocksMutatingTool verifies that a mutating tool is denied when
// write is disabled, even with an allow-all permission gate.
func TestWriteGateBlocksMutatingTool(t *testing.T) {
	gw := &recordingGateway{
		specs:  []service.ToolSpec{{Name: "create_order"}},
		result: service.ToolResult{Content: `{}`},
	}
	prov := &scriptedProvider{steps: []scriptedStep{
		{toolCalls: []service.ToolCall{{ID: "t1", Name: "create_order", Input: json.RawMessage(`{}`)}}, stop: valueobject.StopReasonToolUse},
		{text: "Writes are disabled.", stop: valueobject.StopReasonEndTurn},
	}}
	invs := &fakeInvocationRepo{}

	// writeEnabled=false
	cmd := NewSendMessage(&fakeSessionRepo{}, &fakeMessageRepo{}, invs, prov, gw, service.AllowAllPermissionGate{}, &fakeAudit{}, 1024, false)

	ch, err := cmd.Run(context.Background(), SendMessageInput{UserID: uuid.New(), Content: "place an order"})
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	drainEvents(ch)

	if len(gw.executed) != 0 {
		t.Fatalf("mutating tool must NOT execute when write disabled, ran: %#v", gw.executed)
	}
	if len(invs.rows) != 1 || invs.rows[0].State != valueobject.ToolStateDenied {
		t.Fatalf("expected one denied invocation, got %#v", invs.rows)
	}
}

func contains(haystack, needle string) bool {
	return len(haystack) >= len(needle) && (haystack == needle || indexOf(haystack, needle) >= 0)
}
func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

var _ = errors.New
