package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// FundRepository persists Fund (contract) master records. Soft-delete is
// represented by a non-nil DeletedAt; List/Get methods MUST filter alive rows
// unless an explicit IncludeDeleted flag says otherwise.
//
// Write methods accept a pgx.Tx so they participate in command-layer
// transactions, matching the project-wide repository convention.
type FundRepository interface {
	Create(ctx context.Context, tx pgx.Tx, f *entity.Fund) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Fund, error)
	GetByCode(ctx context.Context, code string) (*entity.Fund, error)
	List(ctx context.Context, filter FundListFilter) ([]*entity.Fund, int, error)
	Update(ctx context.Context, tx pgx.Tx, f *entity.Fund) error
	SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, expectedVersion int, deletedBy uuid.UUID) error
	CountActivePortfolios(ctx context.Context, fundID uuid.UUID) (int, error)
}

// FundListFilter restricts the result set for FundRepository.List.
type FundListFilter struct {
	Status            *vo.FundStatus
	FundCategoryID    *uuid.UUID
	ManagerUserID     *uuid.UUID
	IDs               []uuid.UUID
	AccessibleFundIDs []uuid.UUID
	Page              int
	Limit             int
}

// PortfolioRepository persists Portfolio master records.
type PortfolioRepository interface {
	Create(ctx context.Context, tx pgx.Tx, p *entity.Portfolio) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Portfolio, error)
	GetByFundCode(ctx context.Context, fundID uuid.UUID, code string) (*entity.Portfolio, error)
	List(ctx context.Context, filter PortfolioListFilter) ([]*entity.Portfolio, int, error)
	Update(ctx context.Context, tx pgx.Tx, p *entity.Portfolio) error
	SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, expectedVersion int, deletedBy uuid.UUID) error
	HasOpenActivity(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (bool, string, error)
}

// PortfolioListFilter restricts PortfolioRepository.List.
type PortfolioListFilter struct {
	FundID            *uuid.UUID
	Status            *vo.PortfolioStatus
	ManagerUserID     *uuid.UUID
	AccessibleFundIDs []uuid.UUID // injected by the data-permission layer
	Page              int
	Limit             int
}

// InstrumentRepository persists Instrument master records.
type InstrumentRepository interface {
	Create(ctx context.Context, tx pgx.Tx, inst *entity.Instrument) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Instrument, error)
	List(ctx context.Context, filter InstrumentListFilter) ([]*entity.Instrument, int, error)
	Update(ctx context.Context, tx pgx.Tx, inst *entity.Instrument) error
	SoftDelete(ctx context.Context, tx pgx.Tx, id uuid.UUID, deletedBy uuid.UUID) error
}

// InstrumentListFilter restricts InstrumentRepository.List.
type InstrumentListFilter struct {
	AssetClassID   *uuid.UUID
	AssetSubtypeID *uuid.UUID
	SectorID       *uuid.UUID
	CountryID      *uuid.UUID
	Status         *vo.InstrumentStatus
	Search         string
	Page           int
	Limit          int
}

// InstrumentIdentifierRepository manages external identifier mappings.
type InstrumentIdentifierRepository interface {
	Create(ctx context.Context, tx pgx.Tx, id *entity.InstrumentIdentifier) error
	ListByInstrument(ctx context.Context, instrumentID uuid.UUID) ([]*entity.InstrumentIdentifier, error)
	Delete(ctx context.Context, tx pgx.Tx, id uuid.UUID) error
}

// PortfolioTransactionRepository persists immutable ledger rows.
//
// Implementations MUST NOT expose UPDATE or DELETE operations. The Postgres
// RULE in migration 20260428000004 enforces this at the storage layer; the
// interface mirrors that intent.
type PortfolioTransactionRepository interface {
	Insert(ctx context.Context, tx pgx.Tx, t *entity.PortfolioTransaction) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.PortfolioTransaction, error)
	List(ctx context.Context, filter TransactionListFilter) ([]*entity.PortfolioTransaction, int, error)
	ListForPositionReplay(ctx context.Context, tx pgx.Tx, portfolioID, instrumentID uuid.UUID) ([]*entity.PortfolioTransaction, error)
	HasReversal(ctx context.Context, originalID uuid.UUID) (bool, error)
	SumRealisedPnLBase(ctx context.Context, portfolioID uuid.UUID, asOf time.Time) (decimal.Decimal, error)
}

// TransactionListFilter restricts PortfolioTransactionRepository.List.
type TransactionListFilter struct {
	PortfolioID  *uuid.UUID
	FundID       *uuid.UUID
	InstrumentID *uuid.UUID
	Type         *vo.TransactionType
	BusinessFrom *time.Time
	BusinessTo   *time.Time
	Page         int
	Limit        int
}

// PortfolioPositionRepository manages the position projection.
//
// Upsert is version-aware: callers supply ExpectedVersion to detect
// concurrent posters. A mismatch returns *ErrPortfolioVersionMismatch.
type PortfolioPositionRepository interface {
	GetForUpdate(ctx context.Context, tx pgx.Tx, portfolioID, instrumentID uuid.UUID) (*entity.PortfolioPosition, error)
	Upsert(ctx context.Context, tx pgx.Tx, pos *entity.PortfolioPosition, expectedVersion int) error
	ListByPortfolio(ctx context.Context, portfolioID uuid.UUID) ([]*entity.PortfolioPosition, error)
}

// CashLedgerRepository manages cash movements (immutable) and balances
// (projection). Insert/UpsertBalance are expected to be called within the
// same DB transaction as the parent transaction insert.
type CashLedgerRepository interface {
	InsertMovement(ctx context.Context, tx pgx.Tx, m *entity.CashMovement) error
	GetBalanceForUpdate(ctx context.Context, tx pgx.Tx, portfolioID uuid.UUID, currency string) (*entity.CashBalance, error)
	UpsertBalance(ctx context.Context, tx pgx.Tx, b *entity.CashBalance, expectedVersion int) error
	ListBalances(ctx context.Context, portfolioID uuid.UUID) ([]*entity.CashBalance, error)
}

// PriceSnapshotRepository persists immutable price observations.
type PriceSnapshotRepository interface {
	Insert(ctx context.Context, tx pgx.Tx, p *entity.PriceSnapshot) error
	GetLatest(ctx context.Context, instrumentID uuid.UUID, asOf time.Time) (*entity.PriceSnapshot, error)
	ListByInstrument(ctx context.Context, instrumentID uuid.UUID, from, to time.Time, limit int) ([]*entity.PriceSnapshot, error)
}

// ValuationRepository persists portfolio valuation snapshots and child lines.
type ValuationRepository interface {
	Insert(ctx context.Context, tx pgx.Tx, snap *entity.ValuationSnapshot) error
	GetLatest(ctx context.Context, portfolioID uuid.UUID, source vo.ValuationSource) (*entity.ValuationSnapshot, error)
	GetByID(ctx context.Context, id uuid.UUID) (*entity.ValuationSnapshot, error)
	GetByPortfolioBusinessDate(ctx context.Context, portfolioID uuid.UUID, businessDate time.Time, source vo.ValuationSource) (*entity.ValuationSnapshot, error)
	List(ctx context.Context, portfolioID uuid.UUID, from, to time.Time, page, limit int) ([]*entity.ValuationSnapshot, int, error)

	InsertNAV(ctx context.Context, tx pgx.Tx, nav *entity.NAVSnapshot) error
	ListNAV(ctx context.Context, portfolioID uuid.UUID, from, to time.Time, page, limit int) ([]*entity.NAVSnapshot, int, error)

	InsertAUM(ctx context.Context, tx pgx.Tx, aum *entity.AUMSnapshot) error
	GetAUM(ctx context.Context, scopeType vo.AumScopeType, scopeID uuid.UUID, businessDate time.Time, source vo.ValuationSource) (*entity.AUMSnapshot, error)
	ListAUM(ctx context.Context, scopeType vo.AumScopeType, scopeID uuid.UUID, from, to time.Time, page, limit int) ([]*entity.AUMSnapshot, int, error)
}

// AssetTaxonomyRepository serves the read-side classification reference data.
type AssetTaxonomyRepository interface {
	ListAssetClasses(ctx context.Context, includeInactive bool) ([]*entity.AssetClass, error)
	ListAssetSubtypes(ctx context.Context, assetClassID *uuid.UUID, includeInactive bool) ([]*entity.AssetSubtype, error)
	ListRegions(ctx context.Context) ([]*entity.Region, error)
	ListCountries(ctx context.Context, regionID *uuid.UUID) ([]*entity.Country, error)
	ListSectors(ctx context.Context, parentID *uuid.UUID, level *int) ([]*entity.Sector, error)
	ListFundCategories(ctx context.Context) ([]*entity.FundCategory, error)
	ListInvestmentStyles(ctx context.Context) ([]*entity.InvestmentStyle, error)
}
