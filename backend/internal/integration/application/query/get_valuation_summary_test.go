package query

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"

	"github.com/neo-kanta/ims-th-solution/backend/internal/integration/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

func TestGetValuationSummary_DeniedPermissionReturnsUnavailableWithoutCallingProvider(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: false}
	prov := &fakeValuationSummaryProvider{}
	h := NewGetValuationSummaryHandler(iam, prov)

	res, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.False(t, res.DataAvailable)
	require.False(t, prov.called, "provider must not be consulted when the caller lacks dashboard permission")
}

func TestGetValuationSummary_InvalidScopeReturnsError(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: true}
	prov := &fakeValuationSummaryProvider{}
	h := NewGetValuationSummaryHandler(iam, prov)

	_, err := h.Execute(context.Background(), uuid.New().String(), "alice", "bogus-scope")
	require.ErrorIs(t, err, ErrInvalidValuationScope)
	require.False(t, prov.called)
}

func TestGetValuationSummary_DefaultScopeIsCompany(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{DataAvailable: true, Currency: "THB", AUM: decimal.NewFromInt(1)}}
	h := NewGetValuationSummaryHandler(iam, prov)

	res, err := h.Execute(context.Background(), uuid.New().String(), "alice", "")
	require.NoError(t, err)
	require.Equal(t, domain.ValuationScopeCompany, res.Scope)
	require.Equal(t, contract.ValuationSummaryScopeCompany, prov.gotReq.Scope)
}

// This is the impersonation-prevention test the task calls for: "mine"
// identity always comes from the server-resolved username argument (which
// transport populates from JWT claims), never from anything the caller
// could smuggle in through the scope/query parameters. There is no
// "username" request parameter at all — Execute's signature only accepts
// the scope string from the query string.
func TestGetValuationSummary_MineScopeEchoesServerResolvedUsernameOnly(t *testing.T) {
	t.Parallel()
	uid := uuid.New()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{DataAvailable: true, Currency: "THB", AUM: decimal.NewFromInt(1)}}
	h := NewGetValuationSummaryHandler(iam, prov)

	res, err := h.Execute(context.Background(), uid.String(), "real-session-user", "mine")
	require.NoError(t, err)
	require.Equal(t, "real-session-user", res.Username)
	require.Equal(t, uid, prov.gotReq.UserID, "the provider request's identity must be the authenticated subject, not a client-supplied value")
}

func TestGetValuationSummary_CompanyScopeNeverEchoesUsername(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{DataAvailable: true, Currency: "THB", AUM: decimal.NewFromInt(1)}}
	h := NewGetValuationSummaryHandler(iam, prov)

	res, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.Empty(t, res.Username)
}

func TestGetValuationSummary_WildcardDataScopePassesNilFundFilter(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{DataAvailable: true}}
	h := NewGetValuationSummaryHandler(iam, prov)

	_, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.Nil(t, prov.gotReq.AccessibleFundIDs, `"*" data scope must translate to a nil (no-filter) fund list`)
}

func TestGetValuationSummary_RestrictedDataScopeParsesFundUUIDs(t *testing.T) {
	t.Parallel()
	fundA := uuid.New()
	fundB := uuid.New()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{fundA.String(), fundB.String()}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{DataAvailable: true}}
	h := NewGetValuationSummaryHandler(iam, prov)

	_, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.ElementsMatch(t, []uuid.UUID{fundA, fundB}, prov.gotReq.AccessibleFundIDs)
}

func TestGetValuationSummary_NoProviderDegradesToUnavailable(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	h := NewGetValuationSummaryHandler(iam, nil)

	res, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.False(t, res.DataAvailable)
}

func TestGetValuationSummary_ProviderNoDataMapsToUnavailable(t *testing.T) {
	t.Parallel()
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{DataAvailable: false}}
	h := NewGetValuationSummaryHandler(iam, prov)

	res, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.False(t, res.DataAvailable)
}

func TestGetValuationSummary_MapsSuccessfulResult(t *testing.T) {
	t.Parallel()
	bd := time.Date(2026, 7, 14, 0, 0, 0, 0, time.UTC)
	asOf := bd.Add(6 * time.Hour)
	pct := decimal.RequireFromString("1.01")
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{result: &contract.ValuationSummaryResult{
		DataAvailable: true, Currency: "THB", BusinessDate: bd, AsOf: asOf,
		AUM: decimal.NewFromInt(1234567), TodayPnL: decimal.NewFromInt(12345), TodayPnLPercent: &pct,
	}}
	h := NewGetValuationSummaryHandler(iam, prov)

	res, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.NoError(t, err)
	require.True(t, res.DataAvailable)
	require.Equal(t, "THB", res.Currency)
	require.True(t, bd.Equal(res.BusinessDate))
	require.True(t, decimal.NewFromInt(1234567).Equal(res.AUM))
	require.True(t, decimal.NewFromInt(12345).Equal(res.TodayPnL))
	require.NotNil(t, res.TodayPnLPercent)
}

func TestGetValuationSummary_PropagatesProviderError(t *testing.T) {
	t.Parallel()
	wantErr := errors.New("db down")
	iam := &fakeIAMPort{hasPerm: true, contracts: []string{"*"}}
	prov := &fakeValuationSummaryProvider{err: wantErr}
	h := NewGetValuationSummaryHandler(iam, prov)

	_, err := h.Execute(context.Background(), uuid.New().String(), "alice", "company")
	require.ErrorIs(t, err, wantErr)
}

// ─── fakes ──────────────────────────────────────────────────────────────────

type fakeIAMPort struct {
	hasPerm   bool
	permErr   error
	contracts []string
	scopeErr  error
}

func (p *fakeIAMPort) HasFunctionPermission(context.Context, string, string) (bool, error) {
	return p.hasPerm, p.permErr
}
func (p *fakeIAMPort) GetAccessibleContracts(context.Context, string) ([]string, error) {
	return p.contracts, p.scopeErr
}

var _ IAMPort = (*fakeIAMPort)(nil)

type fakeValuationSummaryProvider struct {
	result *contract.ValuationSummaryResult
	err    error
	called bool
	gotReq contract.ValuationSummaryRequest
}

func (p *fakeValuationSummaryProvider) GetValuationSummary(_ context.Context, req contract.ValuationSummaryRequest) (*contract.ValuationSummaryResult, error) {
	p.called = true
	p.gotReq = req
	return p.result, p.err
}

var _ contract.ValuationSummaryProvider = (*fakeValuationSummaryProvider)(nil)
