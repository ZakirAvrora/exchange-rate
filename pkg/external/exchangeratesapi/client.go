package exchangeratesapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"net/url"

	"github.com/ZakirAvrora/exchange-rate/pkg/httpx"
)

const (
	BASE_URL = "http://api.exchangeratesapi.io"
	APIKey   = "00da5b7225d0ad29aa49021a9b6bc022"
)

type client struct {
	xcl    httpx.Client
	apiKey string
}

type clientOption struct {
	debug   bool
	baseURL string
	apiKey  string
}

type Option func(option *clientOption)

func withBaseURL(URL string) Option {
	return func(co *clientOption) {
		co.baseURL = URL
	}
}

func withAPIKey(apiKey string) Option {
	return func(co *clientOption) {
		co.apiKey = apiKey
	}
}

func withDebug() Option {
	return func(co *clientOption) {
		co.debug = true
	}
}

func NewProvider() (Client, error) {
	return new(withAPIKey(APIKey))
}

func new(opts ...Option) (Client, error) {
	clOpts := clientOption{
		baseURL: BASE_URL,
	}

	for _, o := range opts {
		o(&clOpts)
	}

	httpxOptions := []httpx.ClientOption{
		httpx.WithDefaultHTTPClientWithTimeout(time.Minute),
		httpx.WithParseErrResponse(parseError),
	}

	if clOpts.debug {
		httpxOptions = append(httpxOptions, httpx.WithDebug())
	}

	return &client{
		xcl: *httpx.NewClient(
			"ExchangeRatesAPI",
			clOpts.baseURL,
			httpxOptions...,
		),
		apiKey: clOpts.apiKey,
	}, nil
}

func (c *client) GetSupportedCurrencies(ctx context.Context) ([]string, error) {

	resp := struct {
		Symbols map[string]string `json:"symbols"`
	}{}

	call := httpx.Call{
		Method:   http.MethodGet,
		Path:     "/symbols",
		Query:    url.Values{"access_key": []string{c.apiKey}},
		Response: &resp,
		RequestHeaders: map[string]string{
			"Accept": "application/json",
		},
	}

	err := c.xcl.Do(ctx, call)
	if err != nil {
		return nil, err
	}

	var symbols []string

	for k := range resp.Symbols {
		symbols = append(symbols, k)
	}

	return symbols, nil
}

func (c *client) GetLatestRate(ctx context.Context, base string, target string) (*Rate, error) {
	base = strings.TrimSpace(strings.ToUpper(base))
	target = strings.TrimSpace(strings.ToUpper(target))

	resp := struct {
		Base      string             `json:"base"`
		Timestamp int64              `json:"timestamp"`
		Rates     map[string]float64 `json:"rates"`
	}{}

	call := httpx.Call{
		Method:   http.MethodGet,
		Path:     "/latest",
		Response: &resp,
		Query: url.Values{
			"access_key": []string{c.apiKey},
			"base":       []string{base},
			"symbols":    []string{target},
		},
		RequestHeaders: map[string]string{
			"Accept": "application/json",
		},
	}

	err := c.xcl.Do(ctx, call)
	if err != nil {
		return nil, err
	}

	// if base not supported, API return by default EUR
	if resp.Base != base {
		return nil, ErrNotSupportedBaseCurrency
	}

	if val, ok := resp.Rates[target]; ok {
		return &Rate{
			Value:     val,
			Timestamp: resp.Timestamp,
		}, nil
	}

	return nil, ErrNotSupportedTargetCurrency
}

type errorMsg struct {
	Err Error `json:"error"`
}

type Error struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func parseError(code int, body []byte, header http.Header, v any) error {
	if code >= 400 && code < 500 {
		var resp errorMsg

		if err := json.Unmarshal(body, &resp); err != nil {
			// NoReturn of unmarshalling error just return error msg
			return fmt.Errorf("%w: status code: %d, body: %s ", httpx.ErrNokResponse, code, string(body))
		}

		if err, ok := mapCodeToErrors[ErrorCode(resp.Err.Code)]; ok {
			return fmt.Errorf("%w: status code: %d, msg: %s", err, code, resp.Err.Message)
		} else {
			// By default return httpx.ErrNokResponse
			return fmt.Errorf("%w: status code: %d, msg: %s ", httpx.ErrNokResponse, code, resp.Err.Message)
		}

	}

	// By default server errors returned with default httpx.ErrNokResponse
	return fmt.Errorf("%w: status code: %d, body: %s ", httpx.ErrNokResponse, code, string(body))
}
