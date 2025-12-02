package http_test

import (
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NawafSwe/mockchaos/internal/core/http"
	"github.com/stretchr/testify/assert"
)

func Test_NewHTTPHandler(t *testing.T) {
	tests := map[string]struct {
		handlers           []http.Handler
		expectedStatusCode []int
		path               string
		expectedBody       []byte
	}{
		"should successfully register http test server and randomly respond with status code and latency": {
			handlers: []http.Handler{
				{
					Path:      "/chaos",
					Body:      []byte(`CHAOS!`),
					Headers:   map[string]string{"Content-Type": "application/json"},
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			expectedStatusCode: []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
			path:               "/chaos",
			expectedBody:       []byte(`CHAOS!`),
		},
		"should return default response when no handler found for the given path and method": {
			handlers: []http.Handler{
				{
					Path:      "/chaos",
					Body:      []byte(`CHAOS!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
				{
					Path:      "/login",
					Body:      []byte(`CHAOS!`),
					Method:    nethttp.MethodPost,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			expectedStatusCode: []int{nethttp.StatusNotFound},
			path:               "/unknown/path",
			expectedBody:       []byte(`{"error": "Not found"}`),
		},
		"should return ok response when no statuses and latencies are provided": {
			handlers: []http.Handler{
				{
					Path:   "/chaos",
					Body:   []byte(`CHAOS!`),
					Method: nethttp.MethodGet,
				},
			},
			path:               "/chaos",
			expectedBody:       []byte(`CHAOS!`),
			expectedStatusCode: []int{nethttp.StatusOK},
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			handlers := http.NewHTTPHandler(tc.handlers...)
			assert.NotNil(t, handlers)
			// assert handlers can be called.
			svc := httptest.NewServer(handlers)
			assert.NotNil(t, svc)

			req, err := nethttp.NewRequest(nethttp.MethodGet, svc.URL+tc.path, nil)
			assert.NoError(t, err)
			assert.NotNil(t, req)

			c := nethttp.Client{}
			res, err := c.Do(req)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			// assert statues codes.
			assert.Contains(t, tc.expectedStatusCode, res.StatusCode)

			// assert body
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBody, body)
		})
	}
}
