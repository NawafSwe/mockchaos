package httptest_test

import (
	"bytes"
	"io"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/nawafswe/mockchaos/core/http"
	"github.com/nawafswe/mockchaos/httptest"
	"github.com/stretchr/testify/assert"
)

func Test_HTTP_Server(t *testing.T) {
	tests := map[string]struct {
		handlers            []http.Handler
		method              string
		body                []byte
		path                string
		expectedBody        []byte
		expectedStatusCode  []int
		latencyNotExceeding time.Duration
	}{
		"should successfully register http test server and randomly respond with status code and latency": {
			handlers: []http.Handler{
				{
					Path:      "/chaos",
					Response:  []byte(`CHAOS!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			method:              nethttp.MethodGet,
			path:                "/chaos",
			expectedBody:        []byte(`CHAOS!`),
			expectedStatusCode:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
			latencyNotExceeding: time.Millisecond * 50,
		},
		"should return default response when no handler found for the given path and method": {
			handlers: []http.Handler{
				{
					Path:      "/chaos",
					Response:  []byte(`CHAOS!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
				{
					Path:      "/login",
					Response:  []byte(`CHAOS!`),
					Method:    nethttp.MethodPost,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			method:              nethttp.MethodPost,
			path:                "/chaos",
			expectedBody:        []byte(`{"error": "Not found"}`),
			expectedStatusCode:  []int{nethttp.StatusNotFound},
			latencyNotExceeding: time.Millisecond * 10,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			svc := httptest.NewServer(t, tc.handlers...)

			assert.NotNil(t, svc)
			req, err := nethttp.NewRequest(tc.method, svc.URL()+tc.path, bytes.NewBuffer(tc.body))
			assert.NoError(t, err)
			assert.NotNil(t, req)
			start := time.Now()
			c := nethttp.Client{}
			res, err := c.Do(req)
			delta := time.Since(start)
			assert.NoError(t, err)
			assert.NotNil(t, res)

			// assert statues codes.
			assert.Contains(t, tc.expectedStatusCode, res.StatusCode)

			// assert latency.
			assert.LessOrEqual(t, delta, tc.latencyNotExceeding)

			// assert body.
			body, err := io.ReadAll(res.Body)
			assert.NoError(t, res.Body.Close())
			assert.NoError(t, err)
			assert.Equal(t, tc.expectedBody, body)
		})
	}
}
