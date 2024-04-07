package exchangerates

import (
	"errors"
	"time"
)

type Record struct {
	Id         int
	Identifier string
	Base       string
	Secondary  string
	Status     Status
	Rate       float64
	Created_At time.Time
	Updated_At time.Time
}

type Status int

var (
	StatusUnknown  Status = 0
	StatusCreated  Status = 1
	StatusUpdated  Status = 2
	StatusFailed   Status = 3
	StatusSentinel Status = 4
)

var (
	ErrNoRecord                      = errors.New("no record was found")
	ErrNotSupportedBaseCurrency      = errors.New("invalid base currency code")
	ErrNotSupportedSecondaryCurrency = errors.New("invalid secondary currency code")
)

func validBaseCurrency(code string) bool {
	return supportedBaseCurrency[code]
}

func validSecondaryCurrency(code string) bool {
	return supportedSecondaryCurrency[code]
}

var supportedBaseCurrency = map[string]bool{
	"EUR": true,
}

var supportedSecondaryCurrency = map[string]bool{
	"BTC": true,
	"MXN": true,
	"USD": true,
	"BYR": true,
	"AED": true,
	"KZT": true,
	"RUB": true,
	"XAU": true,
	"XAG": true,
	"LYD": true,
}
