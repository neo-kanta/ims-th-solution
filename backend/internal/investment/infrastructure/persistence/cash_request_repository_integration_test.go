package persistence

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// cashReqIntegrationSchema stands up an FK-free copy of
// investment__portfolio_cash_requests (matching the real column set plus the
// idempotency-key and request-fingerprint additions and BOTH partial indexes) in
// a throwaway schema, and returns a pool bound to it. It mirrors the convention
// in cash_repository_concurrency_test.go: focused on the repository's own SQL and
// the real Postgres index/uniqueness behaviour, without dragging in every FK
// parent table.
func cashReqIntegrationPool(t *testing.T) (*pgxpool.Pool, func()) {
	t.Helper()
	ctx := context.Background()
	dsn := integrationDSN()
	admin, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Skipf("Postgres integration database unavailable: %v", err)
	}

	schema := "inv_cashreq_it_" + strings.ReplaceAll(uuid.NewString(), "-", "_")
	_, err = admin.Exec(ctx, fmt.Sprintf("CREATE SCHEMA %s", schema))
	require.NoError(t, err)

	_, err = admin.Exec(ctx, fmt.Sprintf(`
		CREATE TABLE %s.investment__portfolio_cash_requests (
			id                    UUID          PRIMARY KEY,
			portfolio_id          UUID          NOT NULL,
			fund_id               UUID,
			transaction_type      VARCHAR(20)   NOT NULL,
			amount                DECIMAL(28,8) NOT NULL,
			currency              CHAR(3)       NOT NULL,
			fees                  DECIMAL(28,8) NOT NULL DEFAULT 0,
			value_date            DATE          NOT NULL,
			memo                  TEXT,
			status                VARCHAR(20)   NOT NULL DEFAULT 'PENDING',
			approval_request_id   UUID,
			resulting_txn_id      UUID,
			idempotency_key       VARCHAR(255),
			request_fingerprint   VARCHAR(64),
			submitted_by          UUID          NOT NULL,
			submitted_at          TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			decided_by            UUID,
			decided_at            TIMESTAMPTZ,
			version               INTEGER       NOT NULL DEFAULT 1,
			created_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			updated_at            TIMESTAMPTZ   NOT NULL DEFAULT NOW(),
			created_by            UUID          NOT NULL,
			updated_by            UUID          NOT NULL
		);
		CREATE UNIQUE INDEX uq_inv_cash_req_idempotency_%s
			ON %s.investment__portfolio_cash_requests (portfolio_id, idempotency_key)
			WHERE idempotency_key IS NOT NULL;
		CREATE INDEX idx_inv_cash_req_fingerprint_%s
			ON %s.investment__portfolio_cash_requests (portfolio_id, request_fingerprint)
			WHERE request_fingerprint IS NOT NULL;

		-- Mirrors 20260727000004's terminal-state guard (G2 item 7) so
		-- TestIntegrationCashRequestConcurrentApproveVsCancel exercises the real
		-- DB-level backstop, not just the application's row-lock discipline.
		CREATE OR REPLACE FUNCTION %s.inv_cash_req_prevent_terminal_status_change() RETURNS TRIGGER AS $BODY$
		BEGIN
			IF OLD.status IN ('APPROVED', 'REJECTED', 'CANCELLED')
			   AND NEW.status IS DISTINCT FROM OLD.status THEN
				RAISE EXCEPTION 'cash request %% is in terminal status %% and cannot transition to %%',
					OLD.id, OLD.status, NEW.status
					USING ERRCODE = 'check_violation';
			END IF;
			RETURN NEW;
		END;
		$BODY$ LANGUAGE plpgsql;
		CREATE TRIGGER trg_inv_cash_req_terminal_status_immutable
			BEFORE UPDATE ON %s.investment__portfolio_cash_requests
			FOR EACH ROW EXECUTE FUNCTION %s.inv_cash_req_prevent_terminal_status_change();`,
		schema, schema, schema, schema, schema, schema, schema, schema))
	require.NoError(t, err)
	admin.Close(ctx)

	cfg, err := pgxpool.ParseConfig(dsn)
	require.NoError(t, err)
	cfg.MaxConns = 8
	cfg.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		_, err := conn.Exec(ctx, "SET search_path TO "+schema)
		return err
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	require.NoError(t, err)

	cleanup := func() {
		pool.Close()
		a, err := pgx.Connect(ctx, dsn)
		if err == nil {
			_, _ = a.Exec(ctx, fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE", schema))
			a.Close(ctx)
		}
	}
	return pool, cleanup
}

func newCashReqRow(portfolioID, actor uuid.UUID, key, fingerprint string) *entity.PortfolioCashRequest {
	now := time.Now().UTC()
	var keyPtr, fpPtr *string
	if key != "" {
		keyPtr = &key
	}
	if fingerprint != "" {
		fpPtr = &fingerprint
	}
	return &entity.PortfolioCashRequest{
		ID:                 uuid.New(),
		PortfolioID:        portfolioID,
		FundID:             nil,
		TransactionType:    vo.TransactionTypeCashIn,
		Amount:             decimal.NewFromInt(1000),
		Currency:           "THB",
		Fees:               decimal.Zero,
		ValueDate:          time.Date(2026, 7, 27, 0, 0, 0, 0, time.UTC),
		Memo:               "integration test",
		Status:             vo.CashRequestStatusPending,
		IdempotencyKey:     keyPtr,
		RequestFingerprint: fpPtr,
		SubmittedBy:        actor,
		SubmittedAt:        now,
		Version:            1,
		CreatedAt:          now,
		UpdatedAt:          now,
		CreatedBy:          actor,
		UpdatedBy:          actor,
	}
}

// TestIntegrationCashRequestRoundTrip proves Create then GetByIdempotencyKey and
// GetByFingerprint round-trip a row carrying a non-nil key AND fingerprint on a
// real Postgres, including that the fingerprint column persists and is queryable.
func TestIntegrationCashRequestRoundTrip(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	pool, cleanup := cashReqIntegrationPool(t)
	defer cleanup()

	repo := NewPostgresCashRequestRepository(pool)
	portfolioID := uuid.New()
	actor := uuid.New()
	row := newCashReqRow(portfolioID, actor, "round-trip-key", "fp-round-trip-1234")

	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error {
		return repo.Create(ctx, tx, row)
	}))

	byKey, err := repo.GetByIdempotencyKey(ctx, portfolioID, "round-trip-key")
	require.NoError(t, err)
	require.NotNil(t, byKey)
	require.Equal(t, row.ID, byKey.ID)
	require.NotNil(t, byKey.IdempotencyKey)
	require.Equal(t, "round-trip-key", *byKey.IdempotencyKey)
	require.NotNil(t, byKey.RequestFingerprint)
	require.Equal(t, "fp-round-trip-1234", *byKey.RequestFingerprint)

	byFp, err := repo.GetByFingerprint(ctx, portfolioID, "fp-round-trip-1234")
	require.NoError(t, err)
	require.NotNil(t, byFp)
	require.Equal(t, row.ID, byFp.ID)

	// Miss cases return (nil, nil), not an error.
	miss, err := repo.GetByIdempotencyKey(ctx, portfolioID, "no-such-key")
	require.NoError(t, err)
	require.Nil(t, miss)
	missFp, err := repo.GetByFingerprint(ctx, portfolioID, "no-such-fp")
	require.NoError(t, err)
	require.Nil(t, missFp)
}

// TestIntegrationCashRequestFingerprintDeterministic proves GetByFingerprint
// returns the FIRST-submitted row when two DISTINCT client keys carry the same
// payload fingerprint (the index is non-unique — this is reconciliation-only).
func TestIntegrationCashRequestFingerprintDeterministic(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	pool, cleanup := cashReqIntegrationPool(t)
	defer cleanup()

	repo := NewPostgresCashRequestRepository(pool)
	portfolioID := uuid.New()
	actor := uuid.New()

	first := newCashReqRow(portfolioID, actor, "key-a", "shared-fp")
	first.SubmittedAt = time.Now().UTC().Add(-time.Hour)
	second := newCashReqRow(portfolioID, actor, "key-b", "shared-fp")
	second.SubmittedAt = time.Now().UTC()

	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error { return repo.Create(ctx, tx, first) }))
	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error { return repo.Create(ctx, tx, second) }))

	got, err := repo.GetByFingerprint(ctx, portfolioID, "shared-fp")
	require.NoError(t, err)
	require.NotNil(t, got)
	require.Equal(t, first.ID, got.ID, "GetByFingerprint returns the earliest-submitted match")
}

// TestIntegrationCashRequestConcurrentSameKey proves the (portfolio_id,
// idempotency_key) partial unique index makes duplicate submits race-safe on a
// real Postgres: N goroutines insert the same (portfolio, key); exactly one row
// survives and the losers observe a 23505 unique violation.
func TestIntegrationCashRequestConcurrentSameKey(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	pool, cleanup := cashReqIntegrationPool(t)
	defer cleanup()

	repo := NewPostgresCashRequestRepository(pool)
	portfolioID := uuid.New()
	actor := uuid.New()
	const key = "concurrent-key"
	const n = 8

	var wg sync.WaitGroup
	results := make(chan error, n)
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			row := newCashReqRow(portfolioID, actor, key, "concurrent-fp")
			results <- withTestTx(ctx, pool, func(tx pgx.Tx) error {
				return repo.Create(ctx, tx, row)
			})
		}()
	}
	wg.Wait()
	close(results)

	var success, uniqueViolations int
	for err := range results {
		switch {
		case err == nil:
			success++
		case isUniqueViolationErr(err):
			uniqueViolations++
		default:
			t.Fatalf("unexpected error (want nil or 23505): %v", err)
		}
	}
	require.Equal(t, 1, success, "exactly one insert survives the race")
	require.Equal(t, n-1, uniqueViolations, "every loser observed the unique-violation path")

	// Exactly one physical row exists for the key.
	var count int
	require.NoError(t, pool.QueryRow(ctx,
		"SELECT COUNT(*) FROM investment__portfolio_cash_requests WHERE portfolio_id = $1 AND idempotency_key = $2",
		portfolioID, key).Scan(&count))
	require.Equal(t, 1, count)

	survivor, err := repo.GetByIdempotencyKey(ctx, portfolioID, key)
	require.NoError(t, err)
	require.NotNil(t, survivor)
}

// TestIntegrationCashRequestConcurrentApproveVsCancel proves the terminal-
// state DB backstop (G2 item 7, migration 20260727000004) on a real Postgres:
// two goroutines race GetForUpdate + Update on the SAME PENDING row toward
// DIFFERENT terminal statuses (APPROVED vs CANCELLED), each WITHOUT
// re-checking the row's status after acquiring the lock — i.e. simulating a
// hypothetical bug in the application-level "recheck under lock" discipline
// that both ApplyCashRequestApproval and CancelCashRequest normally apply.
// Postgres' FOR UPDATE row lock serialises the two transactions, so exactly
// one of them observes the row as still PENDING and succeeds; the other
// observes it as already terminal and — because it skips the recheck this
// test is deliberately designed to bypass — attempts the Update anyway, which
// the trg_inv_cash_req_terminal_status_immutable trigger must then refuse.
func TestIntegrationCashRequestConcurrentApproveVsCancel(t *testing.T) {
	if testing.Short() {
		t.Skip("integration test")
	}
	ctx := context.Background()
	pool, cleanup := cashReqIntegrationPool(t)
	defer cleanup()

	repo := NewPostgresCashRequestRepository(pool)
	portfolioID := uuid.New()
	actor := uuid.New()
	row := newCashReqRow(portfolioID, actor, "", "")
	require.NoError(t, withTestTx(ctx, pool, func(tx pgx.Tx) error { return repo.Create(ctx, tx, row) }))

	attempt := func(targetStatus vo.CashRequestStatus) error {
		return withTestTx(ctx, pool, func(tx pgx.Tx) error {
			locked, err := repo.GetForUpdate(ctx, tx, row.ID)
			if err != nil {
				return err
			}
			// Deliberately NOT checking locked.IsPending() here — this test
			// exists specifically to prove the DB trigger catches the case
			// where that application-level recheck is bypassed.
			locked.Status = targetStatus
			now := time.Now().UTC()
			locked.DecidedBy = &actor
			locked.DecidedAt = &now
			locked.UpdatedBy = actor
			locked.UpdatedAt = now
			return repo.Update(ctx, tx, locked)
		})
	}

	var wg sync.WaitGroup
	results := make(chan error, 2)
	for _, status := range []vo.CashRequestStatus{vo.CashRequestStatusApproved, vo.CashRequestStatusCancelled} {
		wg.Add(1)
		go func(s vo.CashRequestStatus) {
			defer wg.Done()
			results <- attempt(s)
		}(status)
	}
	wg.Wait()
	close(results)

	var successes, blocked int
	for err := range results {
		switch {
		case err == nil:
			successes++
		case err != nil && strings.Contains(err.Error(), "terminal status"):
			blocked++
		default:
			t.Fatalf("unexpected error (want nil or a terminal-status trigger error): %v", err)
		}
	}
	require.Equal(t, 1, successes, "exactly one of approve/cancel must win the race")
	require.Equal(t, 1, blocked, "the loser must be refused by the terminal-state trigger, not silently overwrite the winner")

	final, err := repo.GetByID(ctx, row.ID)
	require.NoError(t, err)
	require.NotNil(t, final)
	require.True(t, final.Status == vo.CashRequestStatusApproved || final.Status == vo.CashRequestStatusCancelled)
	require.NotEqual(t, vo.CashRequestStatusPending, final.Status, "the row must have transitioned exactly once, to exactly one terminal status")
}

// isUniqueViolationErr mirrors command.isUniqueViolation (SQLSTATE 23505) without
// importing that package.
func isUniqueViolationErr(err error) bool {
	if err == nil {
		return false
	}
	type sqlStateProvider interface{ SQLState() string }
	var pe sqlStateProvider
	if errors.As(err, &pe) {
		return pe.SQLState() == "23505"
	}
	return false
}
