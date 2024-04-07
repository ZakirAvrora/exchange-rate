package mockery

import (
	"github.com/ZakirAvrora/exchange-rate/pkg/external/exchangeratesapi"
	"github.com/stretchr/testify/mock"
)

//go:generate mockery --dir=.. --outpkg=mockery --output=. --case=snale --name=Client --with-expecter

var _ exchangeratesapi.Client = (*Client)(nil)

// TSetup will assert the mock expectations once the test completes.
func (_m *Client) TSetup(t mock.TestingT, expectedCalls ...*mock.Call) *Client {
	_m.ExpectedCalls = append(_m.ExpectedCalls, expectedCalls...)

	if t, ok := t.(interface {
		mock.TestingT
		Cleanup(func())
	}); ok {
		t.Cleanup(func() {
			_m.AssertExpectations(t)
		})
	}

	return _m
}
