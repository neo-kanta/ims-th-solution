package service

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

func TestPortfolioProjector_RecomputesPositionAfterFirstInsertRace(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	txn := securityTxn(portfolioID, instrumentID, vo.TransactionTypeBuy, "5", "10", "-50", time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC), uuid.New())

	positions := &scriptedPositionRepo{
		gets: []*entity.PortfolioPosition{
			nil,
			{
				PortfolioID:  portfolioID,
				InstrumentID: instrumentID,
				Quantity:     dec("7"),
				AverageCost:  dec("10"),
				CostBasis:    dec("70"),
				Version:      1,
			},
		},
		upsertErrs: []error{
			&domain.ErrPortfolioVersionMismatch{PortfolioID: portfolioID.String(), ExpectedVersion: 0, ActualVersion: 1},
			nil,
		},
	}
	cash := &scriptedCashRepo{gets: []*entity.CashBalance{nil}}

	err := NewPortfolioProjector(positions, cash).Apply(context.Background(), nil, txn, 1)
	require.NoError(t, err)
	require.Len(t, positions.upserts, 2)

	final := positions.upserts[1]
	require.True(t, dec("12").Equal(final.Quantity), "quantity = %s", final.Quantity)
	require.True(t, dec("10").Equal(final.AverageCost), "average cost = %s", final.AverageCost)
	require.True(t, dec("120").Equal(final.CostBasis), "cost basis = %s", final.CostBasis)
	require.Equal(t, 2, final.Version)
}

func TestPortfolioProjector_RecomputesCashAfterFirstInsertRace(t *testing.T) {
	portfolioID := uuid.New()
	txn := cashTxn(portfolioID, vo.TransactionTypeCashIn, "100", time.Date(2026, 4, 28, 0, 0, 0, 0, time.UTC), uuid.New())

	cash := &scriptedCashRepo{
		gets: []*entity.CashBalance{
			nil,
			{
				PortfolioID: portfolioID,
				Currency:    "THB",
				Balance:     dec("40"),
				Version:     1,
			},
		},
		upsertErrs: []error{
			&domain.ErrPortfolioVersionMismatch{PortfolioID: portfolioID.String(), ExpectedVersion: 0, ActualVersion: 1},
			nil,
		},
	}

	err := NewPortfolioProjector(&scriptedPositionRepo{}, cash).Apply(context.Background(), nil, txn, 1)
	require.NoError(t, err)
	require.Len(t, cash.upserts, 2)

	final := cash.upserts[1]
	require.True(t, dec("140").Equal(final.Balance), "balance = %s", final.Balance)
	require.Equal(t, 2, final.Version)
	require.Len(t, cash.movements, 1)
}

func TestPortfolioProjector_ApplyReversalUsesReversalCashMetadataAndReducesBuyPosition(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	originalActor := uuid.New()
	reversalActor := uuid.New()
	originalDate := time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC)
	reversalDate := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)

	original := securityTxn(portfolioID, instrumentID, vo.TransactionTypeBuy, "100", "50", "-5000", originalDate, originalActor)
	reversal := reversalTxn(original, "5000", reversalDate, reversalActor)

	positions := &scriptedPositionRepo{gets: []*entity.PortfolioPosition{{
		PortfolioID:  portfolioID,
		InstrumentID: instrumentID,
		Quantity:     dec("100"),
		AverageCost:  dec("50"),
		CostBasis:    dec("5000"),
		Version:      1,
	}}}
	cash := &scriptedCashRepo{gets: []*entity.CashBalance{{
		PortfolioID: portfolioID,
		Currency:    "THB",
		Balance:     dec("-5000"),
		Version:     1,
	}}}

	err := NewPortfolioProjector(positions, cash).ApplyReversal(context.Background(), nil, reversal, original)
	require.NoError(t, err)

	require.Len(t, positions.upserts, 1)
	pos := positions.upserts[0]
	require.True(t, decimal.Zero.Equal(pos.Quantity), "quantity = %s", pos.Quantity)
	require.Equal(t, reversal.ID, *pos.LastTransactionID)
	require.Equal(t, reversalDate, *pos.LastBusinessDate)

	require.Len(t, cash.movements, 1)
	movement := cash.movements[0]
	require.Equal(t, reversal.ID, *movement.TransactionID)
	require.Equal(t, reversalDate, movement.BusinessDate)
	require.Equal(t, reversalActor, movement.CreatedBy)
	require.Equal(t, "REVERSAL", movement.MovementType)
	require.True(t, dec("5000").Equal(movement.Amount), "movement amount = %s", movement.Amount)

	require.Len(t, cash.upserts, 1)
	require.True(t, decimal.Zero.Equal(cash.upserts[0].Balance), "balance = %s", cash.upserts[0].Balance)
}

func TestPortfolioProjector_ApplyReversalRestoresSellPositionAtAverageCost(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	original := securityTxn(portfolioID, instrumentID, vo.TransactionTypeSell, "20", "12", "240", time.Date(2026, 4, 24, 0, 0, 0, 0, time.UTC), uuid.New())
	reversal := reversalTxn(original, "-240", time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC), uuid.New())

	positions := &scriptedPositionRepo{gets: []*entity.PortfolioPosition{{
		PortfolioID:  portfolioID,
		InstrumentID: instrumentID,
		Quantity:     dec("80"),
		AverageCost:  dec("10"),
		CostBasis:    dec("800"),
		Version:      1,
	}}}
	cash := &scriptedCashRepo{gets: []*entity.CashBalance{{
		PortfolioID: portfolioID,
		Currency:    "THB",
		Balance:     dec("240"),
		Version:     1,
	}}}

	err := NewPortfolioProjector(positions, cash).ApplyReversal(context.Background(), nil, reversal, original)
	require.NoError(t, err)

	pos := positions.upserts[0]
	require.True(t, dec("100").Equal(pos.Quantity), "quantity = %s", pos.Quantity)
	require.True(t, dec("10").Equal(pos.AverageCost), "average cost = %s", pos.AverageCost)
	require.True(t, dec("1000").Equal(pos.CostBasis), "cost basis = %s", pos.CostBasis)
	require.True(t, dec("-240").Equal(cash.movements[0].Amount), "movement amount = %s", cash.movements[0].Amount)
	require.True(t, decimal.Zero.Equal(cash.upserts[0].Balance), "balance = %s", cash.upserts[0].Balance)
}

func TestPortfolioProjector_ApplyReversalReplaysBuyHistory(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	biz := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	first := securityTxn(portfolioID, instrumentID, vo.TransactionTypeBuy, "100", "10", "-1000", biz.AddDate(0, 0, -2), uuid.New())
	second := securityTxn(portfolioID, instrumentID, vo.TransactionTypeBuy, "100", "20", "-2000", biz.AddDate(0, 0, -1), uuid.New())
	reversal := reversalTxn(first, "1000", biz, uuid.New())

	positions := &scriptedPositionRepo{gets: []*entity.PortfolioPosition{{
		PortfolioID: portfolioID, InstrumentID: instrumentID,
		Quantity: dec("200"), AverageCost: dec("15"), CostBasis: dec("3000"), Version: 2,
	}}}
	cash := &scriptedCashRepo{gets: []*entity.CashBalance{{PortfolioID: portfolioID, Currency: "THB", Balance: dec("-3000"), Version: 2}}}
	txns := &scriptedTxnRepo{replay: []*entity.PortfolioTransaction{first, second, reversal}}

	err := NewPortfolioProjectorWithTransactions(positions, cash, txns).ApplyReversal(context.Background(), nil, reversal, first)
	require.NoError(t, err)

	pos := positions.upserts[0]
	require.True(t, dec("100").Equal(pos.Quantity), "quantity = %s", pos.Quantity)
	require.True(t, dec("20").Equal(pos.AverageCost), "average cost = %s", pos.AverageCost)
	require.True(t, dec("2000").Equal(pos.CostBasis), "cost basis = %s", pos.CostBasis)
}

func TestPortfolioProjector_ApplyReversalReplaysSellHistory(t *testing.T) {
	portfolioID := uuid.New()
	instrumentID := uuid.New()
	biz := time.Date(2026, 4, 29, 0, 0, 0, 0, time.UTC)
	buy := securityTxn(portfolioID, instrumentID, vo.TransactionTypeBuy, "100", "10", "-1000", biz.AddDate(0, 0, -2), uuid.New())
	sell := securityTxn(portfolioID, instrumentID, vo.TransactionTypeSell, "100", "20", "2000", biz.AddDate(0, 0, -1), uuid.New())
	reversal := reversalTxn(sell, "-2000", biz, uuid.New())

	positions := &scriptedPositionRepo{gets: []*entity.PortfolioPosition{{
		PortfolioID: portfolioID, InstrumentID: instrumentID,
		Quantity: decimal.Zero, AverageCost: decimal.Zero, CostBasis: decimal.Zero, Version: 2,
	}}}
	cash := &scriptedCashRepo{gets: []*entity.CashBalance{{PortfolioID: portfolioID, Currency: "THB", Balance: dec("1000"), Version: 2}}}
	txns := &scriptedTxnRepo{replay: []*entity.PortfolioTransaction{buy, sell, reversal}}

	err := NewPortfolioProjectorWithTransactions(positions, cash, txns).ApplyReversal(context.Background(), nil, reversal, sell)
	require.NoError(t, err)

	pos := positions.upserts[0]
	require.True(t, dec("100").Equal(pos.Quantity), "quantity = %s", pos.Quantity)
	require.True(t, dec("10").Equal(pos.AverageCost), "average cost = %s", pos.AverageCost)
	require.True(t, dec("1000").Equal(pos.CostBasis), "cost basis = %s", pos.CostBasis)
}

type scriptedPositionRepo struct {
	gets       []*entity.PortfolioPosition
	upsertErrs []error
	upserts    []*entity.PortfolioPosition
}

func (r *scriptedPositionRepo) GetForUpdate(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (*entity.PortfolioPosition, error) {
	if len(r.gets) == 0 {
		return nil, nil
	}
	next := clonePosition(r.gets[0])
	r.gets = r.gets[1:]
	return next, nil
}

func (r *scriptedPositionRepo) Upsert(_ context.Context, _ pgx.Tx, pos *entity.PortfolioPosition, _ int) error {
	r.upserts = append(r.upserts, clonePosition(pos))
	if len(r.upsertErrs) == 0 {
		return nil
	}
	err := r.upsertErrs[0]
	r.upsertErrs = r.upsertErrs[1:]
	return err
}

func (r *scriptedPositionRepo) ListByPortfolio(context.Context, uuid.UUID) ([]*entity.PortfolioPosition, error) {
	return nil, nil
}

type scriptedCashRepo struct {
	gets       []*entity.CashBalance
	upsertErrs []error
	movements  []*entity.CashMovement
	upserts    []*entity.CashBalance
}

type scriptedTxnRepo struct {
	replay []*entity.PortfolioTransaction
}

func (r *scriptedTxnRepo) Insert(context.Context, pgx.Tx, *entity.PortfolioTransaction) error {
	return nil
}

func (r *scriptedTxnRepo) GetByID(context.Context, uuid.UUID) (*entity.PortfolioTransaction, error) {
	return nil, nil
}

func (r *scriptedTxnRepo) List(context.Context, domain.TransactionListFilter) ([]*entity.PortfolioTransaction, int, error) {
	return nil, 0, nil
}

func (r *scriptedTxnRepo) ListForPositionReplay(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) ([]*entity.PortfolioTransaction, error) {
	return r.replay, nil
}

func (r *scriptedTxnRepo) HasReversal(context.Context, uuid.UUID) (bool, error) {
	return false, nil
}

func (r *scriptedTxnRepo) SumRealisedPnLBase(context.Context, uuid.UUID, time.Time) (decimal.Decimal, error) {
	return decimal.Zero, nil
}

func (r *scriptedCashRepo) InsertMovement(_ context.Context, _ pgx.Tx, m *entity.CashMovement) error {
	cp := *m
	r.movements = append(r.movements, &cp)
	return nil
}

func (r *scriptedCashRepo) GetBalanceForUpdate(context.Context, pgx.Tx, uuid.UUID, string) (*entity.CashBalance, error) {
	if len(r.gets) == 0 {
		return nil, nil
	}
	next := cloneCashBalance(r.gets[0])
	r.gets = r.gets[1:]
	return next, nil
}

func (r *scriptedCashRepo) UpsertBalance(_ context.Context, _ pgx.Tx, bal *entity.CashBalance, _ int) error {
	r.upserts = append(r.upserts, cloneCashBalance(bal))
	if len(r.upsertErrs) == 0 {
		return nil
	}
	err := r.upsertErrs[0]
	r.upsertErrs = r.upsertErrs[1:]
	return err
}

func (r *scriptedCashRepo) ListBalances(context.Context, uuid.UUID) ([]*entity.CashBalance, error) {
	return nil, nil
}

func securityTxn(portfolioID, instrumentID uuid.UUID, typ vo.TransactionType, qty, price, net string, businessDate time.Time, actor uuid.UUID) *entity.PortfolioTransaction {
	q := dec(qty)
	p := dec(price)
	return &entity.PortfolioTransaction{
		ID:              uuid.New(),
		PortfolioID:     portfolioID,
		FundID:          func() *uuid.UUID { v := uuid.New(); return &v }(),
		InstrumentID:    &instrumentID,
		TransactionType: typ,
		Quantity:        &q,
		Price:           &p,
		Currency:        "THB",
		GrossAmount:     q.Mul(p),
		Fees:            decimal.Zero,
		NetAmount:       dec(net),
		BusinessDate:    businessDate,
		CreatedAt:       businessDate.Add(10 * time.Hour),
		CreatedBy:       actor,
	}
}

func cashTxn(portfolioID uuid.UUID, typ vo.TransactionType, net string, businessDate time.Time, actor uuid.UUID) *entity.PortfolioTransaction {
	return &entity.PortfolioTransaction{
		ID:              uuid.New(),
		PortfolioID:     portfolioID,
		FundID:          func() *uuid.UUID { v := uuid.New(); return &v }(),
		TransactionType: typ,
		Currency:        "THB",
		GrossAmount:     dec(net),
		Fees:            decimal.Zero,
		NetAmount:       dec(net),
		BusinessDate:    businessDate,
		CreatedAt:       businessDate.Add(10 * time.Hour),
		CreatedBy:       actor,
	}
}

func reversalTxn(original *entity.PortfolioTransaction, net string, businessDate time.Time, actor uuid.UUID) *entity.PortfolioTransaction {
	return &entity.PortfolioTransaction{
		ID:                    uuid.New(),
		PortfolioID:           original.PortfolioID,
		FundID:                original.FundID,
		InstrumentID:          original.InstrumentID,
		TransactionType:       vo.TransactionTypeReversal,
		Side:                  original.Side,
		Quantity:              original.Quantity,
		Price:                 original.Price,
		Currency:              original.Currency,
		GrossAmount:           original.GrossAmount,
		Fees:                  original.Fees,
		NetAmount:             dec(net),
		FxRateToBase:          original.FxRateToBase,
		BusinessDate:          businessDate,
		ReversesTransactionID: &original.ID,
		Status:                vo.TransactionStatusPosted,
		CreatedAt:             businessDate.Add(11 * time.Hour),
		CreatedBy:             actor,
	}
}

func clonePosition(pos *entity.PortfolioPosition) *entity.PortfolioPosition {
	if pos == nil {
		return nil
	}
	cp := *pos
	return &cp
}

func cloneCashBalance(bal *entity.CashBalance) *entity.CashBalance {
	if bal == nil {
		return nil
	}
	cp := *bal
	return &cp
}

func dec(v string) decimal.Decimal {
	d, err := decimal.NewFromString(v)
	if err != nil {
		panic(err)
	}
	return d
}
