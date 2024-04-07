package exchangeratesapi

import "errors"

var (
	ErrNotSupportedBaseCurrency   = errors.New("provided base currency not supported")
	ErrNotSupportedTargetCurrency = errors.New("provided target currency not supported")
	ErrMaxAllowedAPICalls         = errors.New("the maximum allowed API amount of monthly API requests has been reached")
	ErrInvalidAPIKey              = errors.New("no API Key was specified or an invalid API Key was specified")
	ErrBaseCurrencyRestricted     = errors.New("base currency restricted")
)

type ErrorCode string

var (
	ErrorCodeInvalidBase         ErrorCode = "invalid_base_currency"
	ErrorCodeInvalidKey          ErrorCode = "invalid_access_key"
	ErrorCodeInvalidCurrencyCode ErrorCode = "invalid_currency_codes"
	ErrorCodeMaxAPICallReached   ErrorCode = "max_requests_reached"
	ErrBaseCurrencyNotSupported  ErrorCode = "base_currency_access_restricted"
)

var mapCodeToErrors = map[ErrorCode]error{
	ErrorCodeInvalidKey:          ErrInvalidAPIKey,
	ErrorCodeInvalidBase:         ErrNotSupportedBaseCurrency,
	ErrorCodeInvalidCurrencyCode: ErrNotSupportedTargetCurrency,
	ErrorCodeMaxAPICallReached:   ErrMaxAllowedAPICalls,
	ErrBaseCurrencyNotSupported:  ErrBaseCurrencyRestricted,
}
