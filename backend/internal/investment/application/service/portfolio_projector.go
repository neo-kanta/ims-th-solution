// Package service hosts cross-handler application logic that does not fit
// cleanly inside a single command — namely the projector (applies a posted
// transaction to position + cash projections) and the valuation runner.
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/shopspring/decimal"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/entity"
	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/policy"
	vo "github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain/valueobject"
)

// PortfolioProjector is the side-effect engine that updates the position +
// cash projections after a ledger row has been written.
//
// It is intentionally split from the command handler so the same logic
// applies whether the row is a fresh post, a reversal, or a replay during a
// drift-repair job. All operations run on the caller-provided pgx.Tx.
type PortfolioProjector struct {
	positions domain.PortfolioPositionRepository
	cash      domain.CashLedgerRepository
	txns      domain.PortfolioTransactionRepository
}

// NewPortfolioProjector wires the projector.
func NewPortfolioProjector(
	positions domain.PortfolioPositionRepository,
	cash domain.CashLedgerRepository,
) *PortfolioProjector {
	return &PortfolioProjector{positions: positions, cash: cash}
}

func NewPortfolioProjectorWithTransactions(
	positions domain.PortfolioPositionRepository,
	cash domain.CashLedgerRepository,
	txns domain.PortfolioTransactionRepository,
) *PortfolioProjector {
	return &PortfolioProjector{positions: positions, cash: cash, txns: txns}
}

// Apply is called immediately after a transaction is inserted into
// investment__portfolio_transactions, within the SAME pgx transaction.
//
// `direction` is +1 for normal posts and -1 for reversals — the average-cost
// math is identical in both cases except for sign of quantities and net cash.
//
// Returns *domain.ErrPortfolioVersionMismatch on optimistic-lock collisions.
func (p *PortfolioProjector) Apply(
	ctx context.Context,
	tx pgx.Tx,
	t *entity.PortfolioTransaction,
	direction int8,
) error {
	if p == nil {
		return fmt.Errorf("portfolio projector not initialised")
	}
	if t == nil {
		return fmt.Errorf("nil transaction passed to projector")
	}
	if direction != 1 && direction != -1 {
		return fmt.Errorf("invalid direction %d", direction)
	}

	if t.TransactionType.IsSecurityTrade() {
		if err := p.applyToPosition(ctx, tx, t, direction); err != nil {
			return err
		}
	}

	return p.applyToCash(ctx, tx, t, direction)
}

// ApplyReversal applies a reversal row using the original row's position
// semantics while preserving the reversal row's audit metadata.
func (p *PortfolioProjector) ApplyReversal(
	ctx context.Context,
	tx pgx.Tx,
	reversal *entity.PortfolioTransaction,
	original *entity.PortfolioTransaction,
) error {
	if p == nil {
		return fmt.Errorf("portfolio projector not initialised")
	}
	if reversal == nil || original == nil {
		return fmt.Errorf("nil transaction passed to reversal projector")
	}
	if reversal.TransactionType != vo.TransactionTypeReversal {
		return fmt.Errorf("reversal transaction type required")
	}

	if original.TransactionType.IsSecurityTrade() {
		if p.txns != nil {
			if err := p.replayPositionAfterReversal(ctx, tx, reversal, original); err != nil {
				return err
			}
			return p.applyToCash(ctx, tx, reversal, 1)
		}
		if err := p.applyToPositionWithEffect(ctx, tx, reversal, -1, original.TransactionType); err != nil {
			return err
		}
	}

	return p.applyToCash(ctx, tx, reversal, 1)
}

func (p *PortfolioProjector) applyToPosition(
	ctx context.Context,
	tx pgx.Tx,
	t *entity.PortfolioTransaction,
	direction int8,
) error {
	return p.applyToPositionWithEffect(ctx, tx, t, direction, t.TransactionType)
}

func (p *PortfolioProjector) applyToPositionWithEffect(
	ctx context.Context,
	tx pgx.Tx,
	t *entity.PortfolioTransaction,
	direction int8,
	effectType vo.TransactionType,
) error {
	if t.InstrumentID == nil || t.Quantity == nil || t.Price == nil {
		return fmt.Errorf("security trade missing instrument/quantity/price")
	}

	for attempt := 0; attempt < 2; attempt++ {
		pos, err := p.positions.GetForUpdate(ctx, tx, t.PortfolioID, *t.InstrumentID)
		if err != nil {
			return fmt.Errorf("loading position for update: %w", err)
		}

		expectedVersion := 0
		if pos == nil {
			pos = &entity.PortfolioPosition{
				PortfolioID:  t.PortfolioID,
				InstrumentID: *t.InstrumentID,
				Quantity:     decimal.Zero,
				AverageCost:  decimal.Zero,
				CostBasis:    decimal.Zero,
				Version:      0,
			}
		} else {
			expectedVersion = pos.Version
		}

		if decreasesPosition(effectType, direction) && t.Quantity.GreaterThan(pos.Quantity) {
			return &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationOversell),
				Detail:    "sell quantity exceeds locked position",
			}
		}
		applyPositionMath(pos, t, direction, effectType)
		if pos.Quantity.Sign() < 0 {
			return &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationOversell),
				Detail:    "projector arrived at negative quantity",
			}
		}

		pos.LastTransactionID = &t.ID
		bd := t.BusinessDate
		pos.LastBusinessDate = &bd
		pos.Version = expectedVersion + 1

		if err := p.positions.Upsert(ctx, tx, pos, expectedVersion); err != nil {
			if attempt == 0 && isVersionMismatch(err) {
				continue
			}
			return fmt.Errorf("upserting position: %w", err)
		}
		return nil
	}

	return &domain.ErrPortfolioVersionMismatch{PortfolioID: t.PortfolioID.String(), ExpectedVersion: -1, ActualVersion: -1}
}

func applyPositionMath(pos *entity.PortfolioPosition, t *entity.PortfolioTransaction, direction int8, effectType vo.TransactionType) {
	priceBase := *t.Price
	feesBase := t.Fees
	if t.FxRateToBase != nil {
		priceBase = priceBase.Mul(*t.FxRateToBase)
		feesBase = feesBase.Mul(*t.FxRateToBase)
	}
	qty := *t.Quantity

	switch {
	case (effectType.IsBuyLike() && direction > 0) ||
		(effectType.IsSellLike() && direction < 0):
		if effectType.IsSellLike() && direction < 0 {
			restoreSell(pos, qty, priceBase)
			return
		}
		out := policy.ApplyBuyAverageCost(policy.AverageCostInputs{
			OldQuantity:    pos.Quantity,
			OldAverageCost: pos.AverageCost,
			BuyQuantity:    qty,
			BuyPriceBase:   priceBase,
			BuyFeesBase:    feesBase,
		})
		pos.Quantity = out.NewQuantity
		pos.AverageCost = out.NewAverageCost
		pos.CostBasis = out.NewCostBasis

	case (effectType.IsSellLike() && direction > 0) ||
		(effectType.IsBuyLike() && direction < 0):
		out := policy.ApplySellAverageCost(policy.SellInputs{
			OldQuantity:    pos.Quantity,
			OldAverageCost: pos.AverageCost,
			SellQuantity:   qty,
			SellPriceBase:  priceBase,
			SellFeesBase:   feesBase,
		})
		pos.Quantity = out.NewQuantity
		pos.AverageCost = out.NewAverageCost
		pos.CostBasis = out.NewCostBasis
		if direction > 0 {
			t.RealisedPnLBase = out.RealisedPnLBase
		}
	}
}

func (p *PortfolioProjector) replayPositionAfterReversal(
	ctx context.Context,
	tx pgx.Tx,
	reversal *entity.PortfolioTransaction,
	original *entity.PortfolioTransaction,
) error {
	if original.InstrumentID == nil {
		return fmt.Errorf("security reversal missing instrument")
	}

	current, err := p.positions.GetForUpdate(ctx, tx, original.PortfolioID, *original.InstrumentID)
	if err != nil {
		return fmt.Errorf("locking position for replay: %w", err)
	}
	expectedVersion := 0
	if current != nil {
		expectedVersion = current.Version
	}

	rows, err := p.txns.ListForPositionReplay(ctx, tx, original.PortfolioID, *original.InstrumentID)
	if err != nil {
		return err
	}

	reversed := map[uuid.UUID]bool{original.ID: true}
	for _, row := range rows {
		if row.IsReversal() && row.ReversesTransactionID != nil {
			reversed[*row.ReversesTransactionID] = true
		}
	}

	replayed := &entity.PortfolioPosition{
		PortfolioID:  original.PortfolioID,
		InstrumentID: *original.InstrumentID,
		Quantity:     decimal.Zero,
		AverageCost:  decimal.Zero,
		CostBasis:    decimal.Zero,
		Version:      expectedVersion + 1,
	}
	if current != nil {
		replayed.ID = current.ID
	}

	for _, row := range rows {
		if row.IsReversal() || reversed[row.ID] {
			continue
		}
		if !row.TransactionType.IsSecurityTrade() {
			continue
		}
		if row.TransactionType.IsSellLike() && row.Quantity != nil && row.Quantity.GreaterThan(replayed.Quantity) {
			return &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationOversell),
				Detail:    "ledger replay arrived at oversell",
			}
		}
		applyPositionMath(replayed, row, 1, row.TransactionType)
		if replayed.Quantity.Sign() < 0 {
			return &domain.ErrPostPreconditionFailed{
				Violation: string(policy.PostViolationOversell),
				Detail:    "ledger replay arrived at negative quantity",
			}
		}
	}

	replayed.LastTransactionID = &reversal.ID
	bd := reversal.BusinessDate
	replayed.LastBusinessDate = &bd

	if err := p.positions.Upsert(ctx, tx, replayed, expectedVersion); err != nil {
		return fmt.Errorf("upserting replayed position: %w", err)
	}
	return nil
}

func decreasesPosition(effectType vo.TransactionType, direction int8) bool {
	return (effectType.IsSellLike() && direction > 0) || (effectType.IsBuyLike() && direction < 0)
}

func restoreSell(pos *entity.PortfolioPosition, qty, fallbackPrice decimal.Decimal) {
	avg := pos.AverageCost
	if avg.Sign() == 0 {
		avg = fallbackPrice
	}
	pos.Quantity = pos.Quantity.Add(qty)
	pos.AverageCost = avg
	pos.CostBasis = pos.Quantity.Mul(avg)
}

func (p *PortfolioProjector) applyToCash(
	ctx context.Context,
	tx pgx.Tx,
	t *entity.PortfolioTransaction,
	direction int8,
) error {
	signed := t.NetAmount
	if direction < 0 {
		signed = signed.Neg()
	}

	movement := &entity.CashMovement{
		ID:            uuid.New(),
		PortfolioID:   t.PortfolioID,
		Currency:      t.Currency,
		Amount:        signed,
		BusinessDate:  t.BusinessDate,
		TransactionID: &t.ID,
		MovementType:  cashMovementTypeFor(t.TransactionType, direction),
		CreatedAt:     t.CreatedAt,
		CreatedBy:     t.CreatedBy,
	}
	if err := p.cash.InsertMovement(ctx, tx, movement); err != nil {
		return fmt.Errorf("inserting cash movement: %w", err)
	}

	for attempt := 0; attempt < 2; attempt++ {
		bal, err := p.cash.GetBalanceForUpdate(ctx, tx, t.PortfolioID, t.Currency)
		if err != nil {
			return fmt.Errorf("loading cash balance: %w", err)
		}

		expectedVersion := 0
		if bal == nil {
			bal = &entity.CashBalance{
				PortfolioID: t.PortfolioID,
				Currency:    t.Currency,
				Balance:     decimal.Zero,
				Version:     0,
			}
		} else {
			expectedVersion = bal.Version
		}

		bal.Balance = bal.Balance.Add(signed)
		bal.LastMovementID = &movement.ID
		bd := t.BusinessDate
		bal.LastBusinessDate = &bd
		bal.Version = expectedVersion + 1

		if err := p.cash.UpsertBalance(ctx, tx, bal, expectedVersion); err != nil {
			if attempt == 0 && isVersionMismatch(err) {
				continue
			}
			return fmt.Errorf("upserting cash balance: %w", err)
		}
		return nil
	}

	return &domain.ErrPortfolioVersionMismatch{PortfolioID: t.PortfolioID.String(), ExpectedVersion: -1, ActualVersion: -1}
}

func isVersionMismatch(err error) bool {
	var mismatch *domain.ErrPortfolioVersionMismatch
	return errors.As(err, &mismatch)
}

func cashMovementTypeFor(t vo.TransactionType, direction int8) string {
	if direction < 0 {
		return "REVERSAL"
	}
	switch t {
	case vo.TransactionTypeBuy, vo.TransactionTypeSell:
		return "TRADE"
	case vo.TransactionTypeSubscription:
		return "SUBSCRIPTION"
	case vo.TransactionTypeRedemption:
		return "REDEMPTION"
	case vo.TransactionTypeCashIn:
		return "CASH_IN"
	case vo.TransactionTypeCashOut:
		return "CASH_OUT"
	case vo.TransactionTypeFee:
		return "FEE"
	case vo.TransactionTypeDividend:
		return "DIVIDEND"
	case vo.TransactionTypeReversal:
		return "REVERSAL"
	default:
		return "TRADE"
	}
}
