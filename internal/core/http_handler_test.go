package core_test

import (
	"io"
	nethttp "net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/NawafSwe/gofi/internal/core"
	"github.com/stretchr/testify/assert"
)

func Test_NewHTTPHandler(t *testing.T) {
	tests := map[string]struct {
		handlers           []core.Handler
		expectedStatusCode []int
		path               string
		expectedBody       []byte
	}{
		"should successfully register http test server and randomly respond with status code and latency": {
			handlers: []core.Handler{
				{
					Path:      "/gofi",
					Body:      []byte(`GoFi!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			expectedStatusCode: []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
			path:               "/gofi",
			expectedBody:       []byte(`GoFi!`),
		},
		"should return default response when no handler found for the given path and method": {
			handlers: []core.Handler{
				{
					Path:      "/gofi",
					Body:      []byte(`GoFi!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
				{
					Path:      "/login",
					Body:      []byte(`GoFi!`),
					Method:    nethttp.MethodPost,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			expectedStatusCode: []int{nethttp.StatusNotFound},
			path:               "/unknown/path",
			expectedBody:       []byte(`{"error": "Not found"}`),
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			handlers := core.NewHTTPHandler(tc.handlers...)
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
