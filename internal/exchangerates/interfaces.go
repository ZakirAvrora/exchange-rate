package exchangerates

import "context"

type RecodsService interface {
	Refresh(context.Context, string, string) (string, error)
	FetchByIdentifier(context.Context, string) (*Record, error)
	FetchLatest(context.Context, string, string) (*Record, error)
	Update(context.Context, string, float64) error
}

type RecordsRepo interface {
	Insert(context.Context, *Record) error
	FetchByIdentifier(context.Context, string) (*Record, error)
	FetchLatest(context.Context, string, string) (*Record, error)
	Update(context.Context, string, float64) error
}
