package logging

import (
	"context"
	"testing"
	"time"

	"github.com/go-chi/chi/v5/middleware"
	"github.com/google/uuid"
)

func TestLogContext_AccumulatesCanonicalFields(t *testing.T) {
	t.Parallel()

	pid := uuid.New()
	fid := uuid.New()
	cid := uuid.New()
	tid := uuid.New()
	aid := uuid.New()
	iid := uuid.New()
	bd := time.Date(2026, 5, 6, 0, 0, 0, 0, time.UTC)

	args := New().
		Actor(aid).
		Portfolio(pid).
		Fund(fid).
		Contract(cid).
		Instrument(iid).
		Transaction(tid).
		BusinessDate(bd).
		String("custom", "value").
		Args()

	if len(args) != 16 {
		t.Fatalf("expected 16 args (8 pairs), got %d", len(args))
	}

	asMap := make(map[string]any)
	for i := 0; i < len(args); i += 2 {
		asMap[args[i].(string)] = args[i+1]
	}

	cases := map[string]any{
		FieldActorID:       aid.String(),
		FieldPortfolioID:   pid.String(),
		FieldFundID:        fid.String(),
		FieldContractID:    cid.String(),
		FieldInstrumentID:  iid.String(),
		FieldTransactionID: tid.String(),
		FieldBusinessDate:  "2026-05-06",
		"custom":           "value",
	}
	for k, want := range cases {
		if asMap[k] != want {
			t.Errorf("field %s = %v, want %v", k, asMap[k], want)
		}
	}
}

func TestLogContext_IgnoresZeroIDs(t *testing.T) {
	t.Parallel()
	args := New().
		Actor(uuid.Nil).
		Portfolio(uuid.Nil).
		Fund(uuid.Nil).
		BusinessDate(time.Time{}).
		Args()
	if len(args) != 0 {
		t.Errorf("zero IDs / zero time must not produce attributes; got %d", len(args))
	}
}

func TestFromRequest_PullsRequestID(t *testing.T) {
	t.Parallel()
	ctx := context.WithValue(context.Background(), middleware.RequestIDKey, "abc-123")
	args := FromRequest(ctx).Args()
	if len(args) != 2 || args[0] != FieldRequestID || args[1] != "abc-123" {
		t.Errorf("FromRequest: got %v", args)
	}
}

func TestWithRequestID_Empty(t *testing.T) {
	t.Parallel()
	if args := WithRequestID("").Args(); len(args) != 0 {
		t.Errorf("empty request id must produce no attrs, got %v", args)
	}
}

func TestNilLogContext_IsSafe(t *testing.T) {
	t.Parallel()
	var lc *LogContext
	// All builder methods on a nil receiver should not panic and must
	// return the same nil pointer so chains stay consistent.
	got := lc.Actor(uuid.New()).Portfolio(uuid.New()).String("k", "v").Args()
	if got != nil {
		t.Errorf("nil LogContext must yield nil Args, got %v", got)
	}
}
