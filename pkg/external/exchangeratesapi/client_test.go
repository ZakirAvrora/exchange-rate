package exchangeratesapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	testAPIKey = "00da5b7225d0ad29aa49021a9b6bc022"
)

func TestClient_GetLatestRate(t *testing.T) {
	tests := []struct {
		name         string
		base         string
		target       string
		body         string
		httpCode     int
		expectedRate Rate
		expectError  bool
		expectedErr  error
	}{
		{
			name:   "Golden result: 200 OK",
			base:   "EUR",
			target: "USD",
			body: `{
				"success": true,
				"timestamp": 1711698845,
				"base": "EUR",
				"date": "2024-03-29",
				"rates": {
					"USD": 1.077064
				}
			}`,

			httpCode: http.StatusOK,
			expectedRate: Rate{
				Timestamp: 1711698845,
				Value:     1.077064,
			},
			expectError: false,
		},
		{
			name:   "Unsupported base currency: 200 OK",
			base:   "STD",
			target: "USD",
			body: `{
				"success": true,
				"timestamp": 1711698845,
				"base": "EUR",
				"date": "2024-03-29",
				"rates": {
					"USD": 1.077064
				}
			}`,

			httpCode:    http.StatusOK,
			expectError: true,
			expectedErr: ErrNotSupportedBaseCurrency,
		},
		{
			name:   "Unsupported base currency: 400 Bad Request",
			base:   "EUR123",
			target: "TZS",
			body: `{
  				"error": {
        			"code": "invalid_base_currency",
        			"message": "An unexpected error ocurred. [Technical Support: support@apilayer.com]"
    			}
			}`,
			httpCode:    http.StatusBadRequest,
			expectError: true,
			expectedErr: ErrNotSupportedBaseCurrency,
		},
		{
			name:   "Unsupported target currency: 200 OK",
			base:   "EUR",
			target: "",
			body: `{
				"success": true,
				"timestamp": 1711698845,
				"base": "EUR",
				"date": "2024-03-29",
				"rates": {
					"USD": 1.077064
				}
			}`,

			httpCode:    http.StatusOK,
			expectError: true,
			expectedErr: ErrNotSupportedTargetCurrency,
		},
		{
			name:   "Unsupported target currency: 400 Bad Request",
			base:   "EUR",
			target: "TZS123",
			body: `{
  				"error": {
        			"code": "invalid_currency_codes",
        			"message": "You have provided one or more invalid Currency Codes. [Required format: currencies=EUR,USD,GBP,...]"
    			}
			}`,
			httpCode:    http.StatusBadRequest,
			expectError: true,
			expectedErr: ErrNotSupportedTargetCurrency,
		},
		{
			name:   "Invalid API key: 401 Unauthorized",
			base:   "EUR",
			target: "USD",
			body: `{
  				"error": {
        			"code": "invalid_access_key",
        			"message": "You have not supplied a valid API Access Key."
    			}
			}`,
			httpCode:    http.StatusUnauthorized,
			expectError: true,
			expectedErr: ErrInvalidAPIKey,
		},
		{
			name:   "Max API request reached: 402 Payment Required",
			base:   "EUR",
			target: "USD",
			body: `{
  				"error": {
        			"code": "max_requests_reached",
        			"message": "The maximum allowed API amount of monthly API requests has been reached"
    			}
			}`,
			httpCode:    http.StatusPaymentRequired,
			expectError: true,
			expectedErr: ErrMaxAllowedAPICalls,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpCode)
				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				w.Write([]byte(tt.body))
			}))

			defer srv.Close()

			ctx := context.Background()
			c, err := new(
				withAPIKey(testAPIKey),
				withBaseURL(srv.URL),
			)

			require.NoError(t, err)

			rate, err := c.GetLatestRate(ctx, tt.base, tt.target)
			if tt.expectError {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.expectedRate.Timestamp, rate.Timestamp)
				assert.Equal(t, tt.expectedRate.Value, rate.Value)
			}

		})
	}
}

func TestClient_GetLatestRate_Integration(t *testing.T) {
	t.Skip("This is an integration test for ExchangeRates API to get latest rate for pair")

	ctx := context.Background()

	c, err := new(
		withAPIKey(testAPIKey),
		withDebug(),
	)

	require.Nil(t, err)

	resp, err := c.GetLatestRate(ctx, "EUR", "USD")
	require.Nil(t, err)
	require.NotNil(t, resp)
}

func TestClient_GetSupportedCurrencies(t *testing.T) {
	tests := []struct {
		name               string
		base               string
		target             string
		body               string
		httpCode           int
		expectedCurrencies []string
		expectError        bool
		expectedErr        error
	}{
		{
			name:   "Golden result: 200 OK",
			base:   "EUR",
			target: "USD",
			body: `{
				"success": true,
				"symbols": {
					"AED": "United Arab Emirates Dirham",
					"BHD": "Bahraini Dinar",
					"BIF": "Burundian Franc",
					"ETB": "Ethiopian Birr",
					"EUR": "Euro",
					"UGX": "Ugandan Shilling",
					"USD": "United States Dollar"
				}
			}`,

			httpCode:           http.StatusOK,
			expectedCurrencies: []string{"AED", "BHD", "BIF", "ETB", "EUR", "UGX", "USD"},
			expectError:        false,
		},
		{
			name:   "Invalid API key: 401 Unauthorized",
			base:   "EUR",
			target: "USD",
			body: `{
				"error": {
					"code": "invalid_access_key",
					"message": "You have not supplied a valid API Access Key."
				}
			}`,
			httpCode:    http.StatusUnauthorized,
			expectError: true,
			expectedErr: ErrInvalidAPIKey,
		},
		{
			name:   "Max API request reached: 402 Payment Required",
			base:   "EUR",
			target: "USD",
			body: `{
				"error": {
					"code": "max_requests_reached",
					"message": "The maximum allowed API amount of monthly API requests has been reached"
				}
			}`,
			httpCode:    http.StatusPaymentRequired,
			expectError: true,
			expectedErr: ErrMaxAllowedAPICalls,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.httpCode)
				w.Header().Set("Content-Type", "application/json;charset=UTF-8")
				w.Write([]byte(tt.body))
			}))

			defer srv.Close()

			ctx := context.Background()
			c, err := new(
				withAPIKey(testAPIKey),
				withBaseURL(srv.URL),
				withDebug(),
			)

			require.NoError(t, err)

			currencies, err := c.GetSupportedCurrencies(ctx)
			if tt.expectError {
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				require.NoError(t, err)
				require.ElementsMatch(t, tt.expectedCurrencies, currencies)
			}

		})
	}
}

func TestClient_GetSupportedCurrencies_Integration(t *testing.T) {
	t.Skip("This is an integration test for ExchangeRates API to get all supported currencies")

	ctx := context.Background()

	c, err := new(
		withAPIKey(testAPIKey),
		withDebug(),
	)

	require.Nil(t, err)

	resp, err := c.GetSupportedCurrencies(ctx)
	require.NoError(t, err)
	require.NotNil(t, resp)
}
