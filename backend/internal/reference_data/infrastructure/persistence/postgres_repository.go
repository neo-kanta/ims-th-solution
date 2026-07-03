package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/reference_data/domain"
)

// PostgresRepository implements domain.SecurityRepository over PostgreSQL.
type PostgresRepository struct {
	pool *pgxpool.Pool
}

// NewPostgresRepository constructs the repository.
func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

// ----- Securities -----

func (r *PostgresRepository) CreateSecurity(ctx context.Context, sec domain.Security) (*domain.Security, error) {
	if r == nil || r.pool == nil {
		return nil, fmt.Errorf("postgres pool not initialised")
	}
	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO securities_master (
			ims_symbol, display_symbol, primary_identifier, name, asset_type,
			currency, country_code, exchange_mic, isin, cusip, figi, status
		) VALUES (
			$1, $2, NULLIF($3,''), $4, $5,
			NULLIF($6,''), NULLIF($7,''), NULLIF($8,''),
			NULLIF($9,''), NULLIF($10,''), NULLIF($11,''),
			$12
		)
		RETURNING id`,
		sec.IMSSymbol, sec.DisplaySymbol, sec.PrimaryIdentifier, sec.Name, string(sec.AssetType),
		sec.Currency, sec.CountryCode, sec.ExchangeMIC, sec.ISIN, sec.CUSIP, sec.FIGI, string(sec.Status),
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("inserting securities_master: %w", err)
	}
	return r.GetSecurityByID(ctx, id)
}

func (r *PostgresRepository) UpdateSecurity(ctx context.Context, sec domain.Security) (*domain.Security, error) {
	if r == nil || r.pool == nil {
		return nil, fmt.Errorf("postgres pool not initialised")
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE securities_master
		   SET display_symbol = $2,
		       primary_identifier = NULLIF($3,''),
		       name = $4,
		       asset_type = $5,
		       currency = NULLIF($6,''),
		       country_code = NULLIF($7,''),
		       exchange_mic = NULLIF($8,''),
		       isin = NULLIF($9,''),
		       cusip = NULLIF($10,''),
		       figi = NULLIF($11,''),
		       status = $12,
		       updated_at = NOW()
		 WHERE id = $1`,
		sec.ID, sec.DisplaySymbol, sec.PrimaryIdentifier, sec.Name, string(sec.AssetType),
		sec.Currency, sec.CountryCode, sec.ExchangeMIC, sec.ISIN, sec.CUSIP, sec.FIGI, string(sec.Status),
	)
	if err != nil {
		return nil, fmt.Errorf("updating securities_master: %w", err)
	}
	return r.GetSecurityByID(ctx, sec.ID)
}

func (r *PostgresRepository) GetSecurityByID(ctx context.Context, id string) (*domain.Security, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectSecurityColumns+`
		  FROM securities_master
		 WHERE id = $1`, id)
	sec, err := scanSecurity(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mappings, err := r.ListProviderMappingsBySecurity(ctx, sec.ID)
	if err == nil {
		sec.ProviderMappings = mappings
	}
	return sec, nil
}

func (r *PostgresRepository) GetSecurityByIMSSymbol(ctx context.Context, imsSymbol string) (*domain.Security, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectSecurityColumns+`
		  FROM securities_master
		 WHERE ims_symbol = $1`, imsSymbol)
	sec, err := scanSecurity(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mappings, err := r.ListProviderMappingsBySecurity(ctx, sec.ID)
	if err == nil {
		sec.ProviderMappings = mappings
	}
	return sec, nil
}

func (r *PostgresRepository) GetSecurityByDisplaySymbol(ctx context.Context, displaySymbol string) (*domain.Security, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectSecurityColumns+`
		  FROM securities_master
		 WHERE display_symbol = $1
		 ORDER BY (status = 'ACTIVE') DESC, created_at ASC
		 LIMIT 1`, displaySymbol)
	sec, err := scanSecurity(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	mappings, err := r.ListProviderMappingsBySecurity(ctx, sec.ID)
	if err == nil {
		sec.ProviderMappings = mappings
	}
	return sec, nil
}

func (r *PostgresRepository) GetSecurityByISIN(ctx context.Context, isin string) (*domain.Security, error) {
	if r == nil || r.pool == nil || strings.TrimSpace(isin) == "" {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectSecurityColumns+`
		  FROM securities_master
		 WHERE isin = $1
		 LIMIT 1`, isin)
	sec, err := scanSecurity(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return sec, nil
}

func (r *PostgresRepository) SearchSecurities(ctx context.Context, filter domain.SecuritySearchFilter) ([]domain.Security, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	args := []any{}
	clauses := []string{}
	if filter.Query != "" {
		args = append(args, "%"+strings.ToLower(filter.Query)+"%")
		clauses = append(clauses, fmt.Sprintf(
			"(LOWER(ims_symbol) LIKE $%d OR LOWER(display_symbol) LIKE $%d OR LOWER(name) LIKE $%d OR LOWER(COALESCE(isin,'')) LIKE $%d)",
			len(args), len(args), len(args), len(args),
		))
	}
	if filter.AssetType != "" {
		args = append(args, string(filter.AssetType))
		clauses = append(clauses, fmt.Sprintf("asset_type = $%d", len(args)))
	}
	if filter.Status != "" {
		args = append(args, string(filter.Status))
		clauses = append(clauses, fmt.Sprintf("status = $%d", len(args)))
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, filter.Limit)
	query := selectSecurityColumns + " FROM securities_master" + where + " ORDER BY ims_symbol ASC LIMIT $" + fmt.Sprint(len(args))

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("searching securities: %w", err)
	}
	defer rows.Close()
	securities := []domain.Security{}
	for rows.Next() {
		sec, err := scanSecurity(rows)
		if err != nil {
			return nil, err
		}
		securities = append(securities, *sec)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	// Provider filter applied post-load to keep the query simple. Also fetches
	// mapping rows so the response includes them.
	out := make([]domain.Security, 0, len(securities))
	for i := range securities {
		mappings, err := r.ListProviderMappingsBySecurity(ctx, securities[i].ID)
		if err == nil {
			securities[i].ProviderMappings = mappings
		}
		if filter.Provider != "" && !hasActiveMappingForProvider(securities[i].ProviderMappings, filter.Provider) {
			continue
		}
		out = append(out, securities[i])
	}
	return out, nil
}

// ----- Provider Mappings -----

func (r *PostgresRepository) AddProviderMapping(ctx context.Context, m domain.ProviderMapping) (*domain.ProviderMapping, error) {
	if r == nil || r.pool == nil {
		return nil, fmt.Errorf("postgres pool not initialised")
	}
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("starting mapping tx: %w", err)
	}
	defer tx.Rollback(ctx)

	if m.IsPrimary {
		// Demote any existing primary mapping for this security+provider.
		_, _ = tx.Exec(ctx, `
			UPDATE security_provider_mappings
			   SET is_primary = false, updated_at = NOW()
			 WHERE security_id = $1 AND provider_code = $2 AND is_primary = true`,
			m.SecurityID, m.ProviderCode,
		)
	}
	var id string
	err = tx.QueryRow(ctx, `
		INSERT INTO security_provider_mappings (
			security_id, provider_code, provider_symbol, provider_exchange,
			provider_asset_type, provider_currency, priority, confidence_score,
			mapping_status, is_primary
		) VALUES (
			$1, $2, $3, NULLIF($4,''),
			NULLIF($5,''), NULLIF($6,''), $7, $8,
			$9, $10
		)
		ON CONFLICT (provider_code, provider_symbol) DO UPDATE
		   SET security_id = EXCLUDED.security_id,
		       provider_exchange = COALESCE(EXCLUDED.provider_exchange, security_provider_mappings.provider_exchange),
		       provider_asset_type = COALESCE(EXCLUDED.provider_asset_type, security_provider_mappings.provider_asset_type),
		       provider_currency = COALESCE(EXCLUDED.provider_currency, security_provider_mappings.provider_currency),
		       priority = EXCLUDED.priority,
		       confidence_score = EXCLUDED.confidence_score,
		       mapping_status = EXCLUDED.mapping_status,
		       is_primary = EXCLUDED.is_primary,
		       updated_at = NOW()
		RETURNING id`,
		m.SecurityID, m.ProviderCode, m.ProviderSymbol, m.ProviderExchange,
		m.ProviderAssetType, m.ProviderCurrency, m.Priority, m.ConfidenceScore,
		string(m.MappingStatus), m.IsPrimary,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("upserting provider mapping: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("committing mapping tx: %w", err)
	}
	return r.GetProviderMapping(ctx, id)
}

func (r *PostgresRepository) GetProviderMapping(ctx context.Context, id string) (*domain.ProviderMapping, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectMappingColumns+`
		  FROM security_provider_mappings
		 WHERE id = $1`, id)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *PostgresRepository) ListProviderMappingsBySecurity(ctx context.Context, securityID string) ([]domain.ProviderMapping, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	rows, err := r.pool.Query(ctx, selectMappingColumns+`
		  FROM security_provider_mappings
		 WHERE security_id = $1
		 ORDER BY is_primary DESC, priority ASC, created_at ASC`, securityID)
	if err != nil {
		return nil, fmt.Errorf("listing mappings: %w", err)
	}
	defer rows.Close()
	out := []domain.ProviderMapping{}
	for rows.Next() {
		m, err := scanMapping(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) GetProviderMappingByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*domain.ProviderMapping, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectMappingColumns+`
		  FROM security_provider_mappings
		 WHERE provider_code = $1 AND provider_symbol = $2`, providerCode, providerSymbol)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *PostgresRepository) GetActiveProviderMappingForSecurity(ctx context.Context, securityID, providerCode string) (*domain.ProviderMapping, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectMappingColumns+`
		  FROM security_provider_mappings
		 WHERE security_id = $1 AND provider_code = $2 AND mapping_status = 'ACTIVE'
		 ORDER BY is_primary DESC, priority ASC
		 LIMIT 1`, securityID, providerCode)
	m, err := scanMapping(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return m, err
}

func (r *PostgresRepository) SetProviderMappingStatus(ctx context.Context, mappingID string, status domain.MappingStatus) error {
	if r == nil || r.pool == nil {
		return nil
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE security_provider_mappings
		   SET mapping_status = $2, updated_at = NOW()
		 WHERE id = $1`,
		mappingID, string(status),
	)
	if err != nil {
		return fmt.Errorf("setting mapping status: %w", err)
	}
	return nil
}

// ----- Unmapped Candidates -----

func (r *PostgresRepository) CreateUnmappedCandidate(ctx context.Context, c domain.UnmappedSecurityCandidate) (*domain.UnmappedSecurityCandidate, error) {
	if r == nil || r.pool == nil {
		return nil, fmt.Errorf("postgres pool not initialised")
	}
	raw, _ := json.Marshal(c.RawPayload)
	if len(raw) == 0 {
		raw = []byte(`{}`)
	}
	var (
		batchID    any
		suggested  any
		confidence any
	)
	if c.BatchID != nil && *c.BatchID != "" {
		batchID = *c.BatchID
	}
	if c.SuggestedSecurityID != nil && *c.SuggestedSecurityID != "" {
		suggested = *c.SuggestedSecurityID
	}
	if c.ConfidenceScore != nil {
		confidence = *c.ConfidenceScore
	}

	var id string
	err := r.pool.QueryRow(ctx, `
		INSERT INTO reference_data_unmapped_security_candidates (
			batch_id, provider_code, provider_symbol, provider_name,
			provider_asset_type, provider_exchange, provider_currency,
			isin, raw_payload, candidate_status, suggested_security_id, confidence_score
		) VALUES (
			$1, $2, $3, NULLIF($4,''),
			NULLIF($5,''), NULLIF($6,''), NULLIF($7,''),
			NULLIF($8,''), $9, $10, $11, $12
		)
		RETURNING id`,
		batchID, c.ProviderCode, c.ProviderSymbol, c.ProviderName,
		c.ProviderAssetType, c.ProviderExchange, c.ProviderCurrency,
		c.ISIN, raw, c.CandidateStatus, suggested, confidence,
	).Scan(&id)
	if err != nil {
		return nil, fmt.Errorf("inserting unmapped candidate: %w", err)
	}
	return r.GetUnmappedCandidate(ctx, id)
}

func (r *PostgresRepository) GetUnmappedCandidate(ctx context.Context, id string) (*domain.UnmappedSecurityCandidate, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectCandidateColumns+`
		  FROM reference_data_unmapped_security_candidates
		 WHERE id = $1`, id)
	c, err := scanCandidate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

func (r *PostgresRepository) ListUnmappedCandidates(ctx context.Context, filter domain.UnmappedCandidateFilter) ([]domain.UnmappedSecurityCandidate, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	args := []any{}
	clauses := []string{}
	if filter.Status != "" {
		args = append(args, filter.Status)
		clauses = append(clauses, fmt.Sprintf("candidate_status = $%d", len(args)))
	}
	if filter.ProviderCode != "" {
		args = append(args, strings.ToLower(filter.ProviderCode))
		clauses = append(clauses, fmt.Sprintf("provider_code = $%d", len(args)))
	}
	if filter.BatchID != "" {
		args = append(args, filter.BatchID)
		clauses = append(clauses, fmt.Sprintf("batch_id = $%d", len(args)))
	}
	where := ""
	if len(clauses) > 0 {
		where = " WHERE " + strings.Join(clauses, " AND ")
	}
	args = append(args, filter.Limit)
	q := selectCandidateColumns + " FROM reference_data_unmapped_security_candidates" + where +
		" ORDER BY created_at DESC LIMIT $" + fmt.Sprint(len(args))

	rows, err := r.pool.Query(ctx, q, args...)
	if err != nil {
		return nil, fmt.Errorf("listing candidates: %w", err)
	}
	defer rows.Close()
	out := []domain.UnmappedSecurityCandidate{}
	for rows.Next() {
		c, err := scanCandidate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, *c)
	}
	return out, rows.Err()
}

func (r *PostgresRepository) UpdateUnmappedCandidateStatus(ctx context.Context, id string, status string, rejectedReason string) error {
	if r == nil || r.pool == nil {
		return nil
	}
	var reason any
	if rejectedReason != "" {
		reason = rejectedReason
	}
	_, err := r.pool.Exec(ctx, `
		UPDATE reference_data_unmapped_security_candidates
		   SET candidate_status = $2,
		       rejected_reason  = COALESCE($3, rejected_reason),
		       resolved_at      = CASE WHEN $2 IN ('MAPPED','REJECTED','DUPLICATE','CONFLICTED') THEN NOW() ELSE resolved_at END
		 WHERE id = $1`,
		id, status, reason,
	)
	if err != nil {
		return fmt.Errorf("updating candidate status: %w", err)
	}
	return nil
}

func (r *PostgresRepository) GetOpenCandidateByProviderSymbol(ctx context.Context, providerCode, providerSymbol string) (*domain.UnmappedSecurityCandidate, error) {
	if r == nil || r.pool == nil {
		return nil, nil
	}
	row := r.pool.QueryRow(ctx, selectCandidateColumns+`
		  FROM reference_data_unmapped_security_candidates
		 WHERE provider_code = $1 AND provider_symbol = $2 AND candidate_status = 'REVIEW_REQUIRED'
		 LIMIT 1`, providerCode, providerSymbol)
	c, err := scanCandidate(row)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	return c, err
}

// ----- Helpers -----

const selectSecurityColumns = `
	SELECT id, ims_symbol, display_symbol, COALESCE(primary_identifier, ''),
	       name, asset_type, COALESCE(currency, ''), COALESCE(country_code, ''),
	       COALESCE(exchange_mic, ''), COALESCE(isin, ''), COALESCE(cusip, ''),
	       COALESCE(figi, ''), status`

const selectMappingColumns = `
	SELECT id, security_id, provider_code, provider_symbol,
	       COALESCE(provider_exchange, ''), COALESCE(provider_asset_type, ''),
	       COALESCE(provider_currency, ''), priority, confidence_score,
	       mapping_status, is_primary`

const selectCandidateColumns = `
	SELECT id, batch_id, provider_code, provider_symbol, COALESCE(provider_name, ''),
	       COALESCE(provider_asset_type, ''), COALESCE(provider_exchange, ''),
	       COALESCE(provider_currency, ''), COALESCE(isin, ''),
	       raw_payload, candidate_status, suggested_security_id, confidence_score,
	       COALESCE(rejected_reason, ''), created_at, resolved_at`

func scanSecurity(scanner pgx.Row) (*domain.Security, error) {
	var (
		sec       domain.Security
		assetType string
		status    string
	)
	err := scanner.Scan(
		&sec.ID, &sec.IMSSymbol, &sec.DisplaySymbol, &sec.PrimaryIdentifier,
		&sec.Name, &assetType, &sec.Currency, &sec.CountryCode,
		&sec.ExchangeMIC, &sec.ISIN, &sec.CUSIP, &sec.FIGI, &status,
	)
	if err != nil {
		return nil, err
	}
	sec.AssetType = domain.AssetType(assetType)
	sec.Status = domain.SecurityStatus(status)
	return &sec, nil
}

func scanMapping(scanner pgx.Row) (*domain.ProviderMapping, error) {
	var (
		m      domain.ProviderMapping
		status string
	)
	err := scanner.Scan(
		&m.ID, &m.SecurityID, &m.ProviderCode, &m.ProviderSymbol,
		&m.ProviderExchange, &m.ProviderAssetType, &m.ProviderCurrency,
		&m.Priority, &m.ConfidenceScore, &status, &m.IsPrimary,
	)
	if err != nil {
		return nil, err
	}
	m.MappingStatus = domain.MappingStatus(status)
	return &m, nil
}

func scanCandidate(scanner pgx.Row) (*domain.UnmappedSecurityCandidate, error) {
	var (
		c                 domain.UnmappedSecurityCandidate
		batchID           *string
		suggestedSecurity *string
		confidence        *decimal.Decimal
		resolvedAt        *time.Time
		rawPayloadBytes   []byte
	)
	err := scanner.Scan(
		&c.ID, &batchID, &c.ProviderCode, &c.ProviderSymbol, &c.ProviderName,
		&c.ProviderAssetType, &c.ProviderExchange, &c.ProviderCurrency, &c.ISIN,
		&rawPayloadBytes, &c.CandidateStatus, &suggestedSecurity, &confidence,
		&c.RejectedReason, new(time.Time), &resolvedAt,
	)
	if err != nil {
		return nil, err
	}
	c.BatchID = batchID
	c.SuggestedSecurityID = suggestedSecurity
	c.ConfidenceScore = confidence
	if len(rawPayloadBytes) > 0 {
		_ = json.Unmarshal(rawPayloadBytes, &c.RawPayload)
	}
	return &c, nil
}

func hasActiveMappingForProvider(mappings []domain.ProviderMapping, providerCode string) bool {
	providerCode = strings.ToLower(strings.TrimSpace(providerCode))
	for _, m := range mappings {
		if m.MappingStatus == domain.MappingStatusActive && strings.EqualFold(m.ProviderCode, providerCode) {
			return true
		}
	}
	return false
}

// Compile-time interface assertion.
var _ domain.SecurityRepository = (*PostgresRepository)(nil)
