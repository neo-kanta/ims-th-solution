package adapter

import (
	"context"
	"fmt"
	"net/http"

	"github.com/google/uuid"

	"github.com/neo-kanta/ims-th-solution/backend/internal/investment/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/errcode"
)

// ContractCatalogAdapter implements contract.ContractCatalog by delegating to
// the investment module's fund repository.
//
// This adapter is temporary. Contract reference data may later move to a
// ReferenceData/ContractMaster module. Workflow must depend only on the
// ContractCatalog interface, never on this adapter or investment internals.
type ContractCatalogAdapter struct {
	funds domain.FundRepository
}

// NewContractCatalogAdapter wires the adapter.
func NewContractCatalogAdapter(funds domain.FundRepository) *ContractCatalogAdapter {
	return &ContractCatalogAdapter{funds: funds}
}

// ResolveContractCode maps a business-readable contractCode to its internal UUID.
// Returns ErrContractCodeEmpty for blank input; ErrContractNotFound when no
// alive fund carries the requested contractCode.
func (a *ContractCatalogAdapter) ResolveContractCode(ctx context.Context, contractCode string) (uuid.UUID, error) {
	if contractCode == "" {
		return uuid.Nil, &ErrContractCodeEmpty{}
	}
	fund, err := a.funds.GetByContractCode(ctx, contractCode)
	if err != nil {
		return uuid.Nil, fmt.Errorf("contract catalog: resolving %q: %w", contractCode, err)
	}
	if fund == nil {
		return uuid.Nil, &ErrContractNotFound{ContractCode: contractCode}
	}
	return fund.ID, nil
}

var _ contract.ContractCatalog = (*ContractCatalogAdapter)(nil)

// ─────────────────────────────────────────────────────────────────────────────
// Typed errors (implement errcode.Coded + errcode.HTTPStatus)
// ─────────────────────────────────────────────────────────────────────────────

// ErrContractNotFound is returned when no alive fund carries the requested
// contract_code. Consumers may check errcode.CodeOf(err) for "CONTRACT_NOT_FOUND".
type ErrContractNotFound struct {
	ContractCode string
}

func (e *ErrContractNotFound) Error() string {
	return fmt.Sprintf("contract not found: %s", e.ContractCode)
}

func (*ErrContractNotFound) ErrorCode() string { return errcode.CodeContractNotFound }
func (*ErrContractNotFound) HTTPStatus() int   { return http.StatusNotFound }
func (e *ErrContractNotFound) ErrorDetails() map[string]any {
	return map[string]any{"contract_code": e.ContractCode}
}

// ErrContractCodeEmpty is returned when contractCode is blank.
type ErrContractCodeEmpty struct{}

func (*ErrContractCodeEmpty) Error() string     { return "contract code must not be empty" }
func (*ErrContractCodeEmpty) ErrorCode() string { return errcode.CodeInvalidRequest }
func (*ErrContractCodeEmpty) HTTPStatus() int   { return http.StatusBadRequest }
