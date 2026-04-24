package adapter

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/neo-kanta/ims-th-solution/backend/internal/compliance/spi"
)

// PostgresRestrictionListAdapter implements ports.RestrictionListPort backed by
// the compliance_restriction_list_entries table.
//
// The query pulls rows that are effective for the caller's business date and
// whose ticker matches the in-scope set (or issuer if supplied). The result is
// grouped into the four canonical buckets on the RestrictionSnapshot.
type PostgresRestrictionListAdapter struct {
	pool *pgxpool.Pool
	// clock is injectable for tests. When nil, time.Now().UTC() is used so the
	// adapter picks up the caller's date rather than a stale fixture.
	clock func() time.Time
}

// NewPostgresRestrictionListAdapter wires the adapter.
func NewPostgresRestrictionListAdapter(pool *pgxpool.Pool) *PostgresRestrictionListAdapter {
	return &PostgresRestrictionListAdapter{pool: pool}
}

// GetRestrictions returns the restriction snapshot. `tickers` is advisory —
// we always include GLOBAL whitelists and sector/alert entries so the rule
// can evaluate negative lookups too. On read failure the adapter returns an
// error and the caller fails closed upstream.
func (a *PostgresRestrictionListAdapter) GetRestrictions(
	ctx context.Context,
	_ uuid.UUID,
	tickers []string,
) (*spi.RestrictionSnapshot, error) {
	if a == nil || a.pool == nil {
		return nil, fmt.Errorf("restriction list adapter not initialised")
	}

	asOf := a.nowUTC()

	// Effective filter: effective_from <= today AND (effective_to IS NULL OR effective_to >= today).
	// Empty ticker list → return whitelist + all active alerts/disposals so the
	// rule can still detect whitelist misses.
	const q = `
		SELECT ticker, reason, list_type, source
		FROM compliance_restriction_list_entries
		WHERE effective_from <= $1
		  AND (effective_to IS NULL OR effective_to >= $1)
		  AND (
		      cardinality($2::text[]) = 0
		      OR list_type = 'WHITELIST'
		      OR ticker = ANY($2::text[])
		  )
	`

	rows, err := a.pool.Query(ctx, q, asOf, tickers)
	if err != nil {
		return nil, fmt.Errorf("querying restriction list: %w", err)
	}
	defer rows.Close()

	snap := &spi.RestrictionSnapshot{
		Blacklisted: map[string]spi.RestrictionEntry{},
		Whitelisted: map[string]bool{},
		GrayListed:  map[string]spi.RestrictionEntry{},
		Alerted:     map[string]spi.RestrictionEntry{},
		Disposal:    map[string]spi.RestrictionEntry{},
	}

	whitelistRows := 0
	for rows.Next() {
		var entry spi.RestrictionEntry
		if err := rows.Scan(&entry.Ticker, &entry.Reason, &entry.ListType, &entry.Source); err != nil {
			return nil, fmt.Errorf("scanning restriction list row: %w", err)
		}
		switch strings.ToUpper(entry.ListType) {
		case string(spi.RestrictionListTypeBlacklist):
			snap.Blacklisted[entry.Ticker] = entry
		case string(spi.RestrictionListTypeWhitelist):
			snap.Whitelisted[entry.Ticker] = true
			whitelistRows++
		case string(spi.RestrictionListTypeAlert):
			snap.Alerted[entry.Ticker] = entry
		case string(spi.RestrictionListTypeDisposal):
			snap.Disposal[entry.Ticker] = entry
		case string(spi.RestrictionListTypeGray):
			// Legacy alias — populate both GrayListed (for legacy rules) and
			// Alerted (so the unified rule treats it uniformly).
			snap.GrayListed[entry.Ticker] = entry
			snap.Alerted[entry.Ticker] = entry
		}
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating restriction list rows: %w", err)
	}

	snap.HasWhitelist = whitelistRows > 0
	return snap, nil
}

func (a *PostgresRestrictionListAdapter) nowUTC() time.Time {
	if a.clock != nil {
		return a.clock()
	}
	return time.Now().UTC()
}
