package httpx

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Doer interface {
	Do(r *http.Request) (*http.Response, error)
}

type Call struct {
	Method          string
	BaseURL         string
	Path            string
	Query           url.Values
	Request         any
	RequestHeaders  map[string]string
	Response        any
	ResponseHeaders url.Values
	ErrResponse     any
	Label           string
	Debug           bool
}

type CallOption func(*http.Request)

type Client struct {
	cl               Doer
	apiName          string
	baseURL          string
	callOpts         []CallOption
	debug            bool
	isResponseOK     func(code int, body []byte) bool
	parseResponse    func(body []byte, header http.Header, v any) error
	parseErrResponse func(code int, body []byte, header http.Header, v any) error
}

type ClientOption func(*Client)

func WithCallOptions(opts ...CallOption) ClientOption {
	return func(c *Client) {
		c.callOpts = append(c.callOpts, opts...)
	}
}

func WithHTTPClient(cl Doer) ClientOption {
	return func(c *Client) {
		c.cl = cl
	}
}

func WithDefaultHTTPClientWithTimeout(tm time.Duration) ClientOption {
	return func(c *Client) {
		c.cl = &http.Client{
			Timeout: tm,
		}
	}
}

func WithResponseCheck(isResponseOK func(code int, body []byte) bool) ClientOption {
	return func(c *Client) {
		c.isResponseOK = isResponseOK
	}
}

func WithBasicAuth(username, password string) ClientOption {
	return func(c *Client) {
		c.callOpts = append(c.callOpts, func(req *http.Request) {
			req.SetBasicAuth(username, password)
		})
	}
}

func WithAuthHeader(typ, cred string) ClientOption {
	return func(c *Client) {
		c.callOpts = append(c.callOpts, func(req *http.Request) {
			req.Header.Set("Authorization", fmt.Sprintf("%s %s", typ, cred))
		})
	}
}

func WithDebug() ClientOption {
	return func(c *Client) {
		c.debug = true
	}
}

func WithParseResponse(fn func(body []byte, header http.Header, val any) error) ClientOption {
	return func(c *Client) {
		c.parseResponse = fn
	}
}

func WithParseErrResponse(fn func(code int, body []byte, header http.Header, v any) error) ClientOption {
	return func(c *Client) {
		c.parseErrResponse = fn
	}
}

func (d *Client) Do(ctx context.Context, c Call, opts ...CallOption) error {
	// TODO: implement retry
	return d.doOnce(ctx, c, opts...)
}

func (d *Client) doOnce(ctx context.Context, c Call, opts ...CallOption) error {
	u := d.baseURL
	if c.BaseURL != "" {
		u = c.BaseURL
	}

	if c.Path != "" {
		u += "/" + strings.TrimLeft(c.Path, "/")
	}

	req, err := http.NewRequestWithContext(ctx, c.Method, u, nil)
	if err != nil {
		return err
	}

	if req.URL.RawQuery != "" {
		return fmt.Errorf("base url plus call path may not contain query parameters")
	}

	if len(c.Query) > 0 {
		req.URL.RawQuery = c.Query.Encode()
	}

	var reqBody []byte
	if c.Request != nil {
		req.Header.Set("Content-Type", "application/json")
		b, err := json.Marshal(c.Request)
		if err != nil {
			return err
		}
		reqBody = b
		req.Body = io.NopCloser(bytes.NewReader(reqBody))
		req.ContentLength = int64(len(reqBody))
	}

	for k, v := range c.RequestHeaders {
		req.Header.Set(k, v)
	}

	for _, opt := range d.callOpts {
		opt(req)
	}

	for _, opt := range opts {
		opt(req)
	}

	if d.debug || c.Debug {
		log.Printf("httpx request\nurl: %s\nbody: %s\nheaders: %s",
			req.URL, string(reqBody), fmt.Sprint(c.RequestHeaders))
	}

	res, err := d.cl.Do(req)
	if err != nil {
		return err
	}

	defer func() { _ = res.Body.Close() }()

	var resBody []byte
	if c.Response != nil || c.ErrResponse != nil {
		resBody, err = io.ReadAll(res.Body)
		if err != nil {
			return err
		}
	}

	if d.debug || c.Debug {
		log.Printf("httpx response\ncode: %d\nbody: %s",
			res.StatusCode, strings.TrimSpace(string(resBody)))
	}

	if c.ResponseHeaders != nil {
		for k, v := range res.Header {
			c.ResponseHeaders[k] = v
		}
	}

	if d.isResponseOK(res.StatusCode, resBody) {
		err = d.parseResponse(resBody, res.Header, c.Response)
		if err != nil {
			return fmt.Errorf("parsing successfull response error: %w", err)
		}

		return nil
	}

	err = d.parseErrResponse(res.StatusCode, resBody, res.Header, c.ErrResponse)
	if err != nil {
		return err
	}

	return nil
}

func (c *Client) GetClient() Doer {
	return c.cl
}

func (c *Client) GetBaseURL() string {
	return c.baseURL
}

func NewClient(apiName string, baseURL string, opts ...ClientOption) *Client {
	c := &Client{
		cl:               &http.Client{},
		apiName:          apiName,
		baseURL:          strings.TrimRight(baseURL, "/"),
		isResponseOK:     isResponseOK,
		parseResponse:    parseJSONresponse,
		parseErrResponse: parseErrResponse,
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

func isResponseOK(code int, _ []byte) bool {
	return code >= 200 && code < 300
}

func parseJSONresponse(body []byte, _ http.Header, v any) error {
	if v == nil {
		return nil
	}

	return json.Unmarshal(body, v)
}

func parseErrResponse(code int, body []byte, _ http.Header, v any) error {
	if v != nil {
		err := json.Unmarshal(body, v)
		if err != nil {
			return err
		}
	}

	return fmt.Errorf("%w: status code: %d, body: %s ", ErrNokResponse, code, string(body))
}
