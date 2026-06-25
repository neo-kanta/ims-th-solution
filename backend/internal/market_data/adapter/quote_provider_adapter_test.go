package adapter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/require"

	mddomain "github.com/neo-kanta/ims-th-solution/backend/internal/market_data/domain"
	"github.com/neo-kanta/ims-th-solution/backend/pkg/contract"
)

func TestTranslateErrorMapsTypedProviderErrorsToContract(t *testing.T) {
	cases := []struct {
		name    string
		input   error
		wantErr error
	}{
		{"rate_limited", &mddomain.ProviderError{Provider: "alpha_vantage", Code: "rate_limited", Message: "throttled"}, contract.ErrRateLimited},
		{"unauthorized", &mddomain.ProviderError{Provider: "alpha_vantage", Code: "unauthorized", Message: "bad key"}, contract.ErrProviderUnauthorized},
		{"missing_api_key", &mddomain.ProviderError{Provider: "alpha_vantage", Code: "missing_api_key", Message: "no key"}, contract.ErrProviderUnauthorized},
		{"timeout", &mddomain.ProviderError{Provider: "yahoo", Code: "timeout", Message: "slow"}, contract.ErrProviderTimeout},
		{"symbol_not_found", &mddomain.ProviderError{Provider: "yahoo", Code: "symbol_not_found", Message: "no such symbol"}, contract.ErrSymbolNotMapped},
		{"malformed_response", &mddomain.ProviderError{Provider: "yahoo", Code: "malformed_response", Message: "bad json"}, contract.ErrMalformedResponse},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := translateError(c.input)
			require.True(t, errors.Is(got, c.wantErr), "translated error = %v; want wrap of %v", got, c.wantErr)
		})
	}
}

func TestTranslateErrorMapsContextDeadlineExceededToTimeout(t *testing.T) {
	got := translateError(context.DeadlineExceeded)
	require.True(t, errors.Is(got, contract.ErrProviderTimeout))
}

func TestTranslateErrorWrapsUnknownAsUnavailable(t *testing.T) {
	got := translateError(errors.New("kapow"))
	require.True(t, errors.Is(got, contract.ErrMarketDataUnavailable))
}
