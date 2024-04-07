package httpx_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/ZakirAvrora/exchange-rate/pkg/httpx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testDTO struct {
	Foo string `json:"foo"`
}

func TestGet(t *testing.T) {

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(testDTO{"bar"})
	}))

	defer srv.Close()

	c := httpx.NewClient("test", srv.URL)

	ctx := context.Background()

	var res testDTO

	err := c.Do(ctx, httpx.Call{
		Method:   http.MethodGet,
		Response: &res,
	})

	require.NoError(t, err)
	require.Equal(t, "bar", res.Foo)
}

func TestJSONPost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		headers := make(http.Header)
		headers.Set("Accept-Encoding", "gzip")
		headers.Set("Content-Length", "13")
		headers.Set("Content-Type", "application/json")
		headers.Set("User-Agent", "Go-http-client/1.1")

		assert.EqualValues(t, headers, r.Header)
		defer r.Body.Close()

		var req testDTO

		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(req)
	}))

	defer srv.Close()

	c := httpx.NewClient("test", srv.URL)

	ctx := context.Background()
	req := testDTO{Foo: "bar"}
	var res testDTO

	err := c.Do(ctx, httpx.Call{
		Method:   http.MethodPost,
		Request:  &req,
		Response: &res,
	})

	require.NoError(t, err)
	require.Equal(t, "bar", res.Foo)
}

func TestInvalidQueryParams(t *testing.T) {
	ctx := context.Background()
	c := httpx.NewClient("test", "localhost")

	err := c.Do(ctx, httpx.Call{
		Method: http.MethodPost,
		Path:   "path?bar=baz",
	})

	require.Error(t, err)
	require.ErrorContains(t, err, "path may not contain query parameters")
}

func TestURLQueryParams(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/path?bar=bar&foo=foo", r.URL.String())
		b, err := io.ReadAll(r.Body)

		require.NoError(t, err)
		require.NoError(t, r.Body.Close())
		require.Empty(t, b)
	}))

	defer srv.Close()

	ctx := context.Background()
	c := httpx.NewClient("test", srv.URL)
	err := c.Do(ctx, httpx.Call{
		Method: http.MethodPost,
		Path:   "path",
		Query: url.Values{
			"bar": []string{"bar"},
			"foo": []string{"foo"},
		},
	})

	require.NoError(t, err)
}

func TestStatusNOK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		var req testDTO
		w.WriteHeader(http.StatusBadRequest)
		json.NewDecoder(r.Body).Decode(&req)
		json.NewEncoder(w).Encode(req)
	}))

	defer srv.Close()

	c := httpx.NewClient("test", srv.URL)

	ctx := context.Background()
	ts := time.Now().String()
	req := testDTO{ts}
	var res, errRes testDTO

	err := c.Do(ctx, httpx.Call{
		Method:      http.MethodPost,
		Request:     &req,
		Response:    &res,
		ErrResponse: &errRes,
	})

	require.ErrorIs(t, err, httpx.ErrNokResponse)
	require.Equal(t, testDTO{}, res)
	require.Equal(t, ts, errRes.Foo)
}

func TestWithResponseHeaders(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json;charset=UTF-8")
		w.Header().Set("Bm-After", "789123132")
		w.Header().Set("Bm-Before", "123654")
		w.Write([]byte(`{"test":"success"}`))
	}))

	defer srv.Close()

	c := httpx.NewClient("test", srv.URL)

	ctx := context.Background()
	var res any
	resHeaders := make(url.Values)
	err := c.Do(ctx, httpx.Call{
		Method:          http.MethodGet,
		Response:        &res,
		ResponseHeaders: resHeaders,
	})

	require.NoError(t, err)
	require.Equal(t, []string{"application/json;charset=UTF-8"}, resHeaders["Content-Type"])
	require.Equal(t, []string{"789123132"}, resHeaders["Bm-After"])
	require.Equal(t, []string{"123654"}, resHeaders["Bm-Before"])
}

func TestOverrideBaseUrl(t *testing.T) {
	ctx := context.Background()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	defer srv.Close()

	c := httpx.NewClient("test", "invalid_base_url", httpx.WithHTTPClient(srv.Client()))

	err := c.Do(ctx, httpx.Call{
		Method:  http.MethodGet,
		BaseURL: srv.URL,
	})

	require.NoError(t, err)
}

func TestWithHTTPClient(t *testing.T) {
	httpClient := &http.Client{}

	c := httpx.NewClient("test", "localhost:8080", httpx.WithHTTPClient(httpClient))
	require.Equal(t, httpClient, c.GetClient())
}

func TestDefaultHTTPClientWithTimeout(t *testing.T) {
	tm := time.Minute * 2
	c := httpx.NewClient("test", "localhost:8080", httpx.WithDefaultHTTPClientWithTimeout(tm))
	cl := c.GetClient()

	require.NotNil(t, cl)
	if hcl, ok := cl.(*http.Client); ok {
		require.Equal(t, tm, hcl.Timeout)
	} else {
		require.Fail(t, "must return default http.Client")
	}
}
