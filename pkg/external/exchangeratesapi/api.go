package exchangeratesapi

import "context"

type Client interface {
	/// GetSupportedCurrencies is
	GetSupportedCurrencies(ctx context.Context) ([]string, error)

	/// GetLatestRate is
	GetLatestRate(ctx context.Context, base string, target string) (*Rate, error)
}

type Rate struct {
	Value     float64
	Timestamp int64
}
