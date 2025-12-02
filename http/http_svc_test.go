package http_test

import (
	"bytes"
	"fmt"
	"io"
	"net"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/NawafSwe/gofi/http"
	http2 "github.com/NawafSwe/gofi/internal/core/http"
	"github.com/stretchr/testify/assert"
)

func Test_HTTP_Server(t *testing.T) {
	tests := map[string]struct {
		handlers            []http2.Handler
		method              string
		body                []byte
		path                string
		expectedBody        []byte
		expectedStatusCode  []int
		latencyNotExceeding time.Duration
	}{
		"should successfully register http test server and randomly respond with status code and latency": {
			handlers: []http2.Handler{
				{
					Path:      "/gofi",
					Body:      []byte(`GoFi!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			method:              nethttp.MethodGet,
			path:                "/gofi",
			expectedBody:        []byte(`GoFi!`),
			expectedStatusCode:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
			latencyNotExceeding: time.Millisecond * 50,
		},
		"should return default response when no handler found for the given path and method": {
			handlers: []http2.Handler{
				{
					Path:      "/gofi",
					Body:      []byte(`GoFi!`),
					Method:    nethttp.MethodGet,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusBadGateway},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
				{
					Path:      "/login",
					Body:      []byte(`GoFi!`),
					Method:    nethttp.MethodPost,
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusBadGateway},
					Latencies: []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			method:              nethttp.MethodPost,
			path:                "/gofi",
			expectedBody:        []byte(`{"error": "Not found"}`),
			expectedStatusCode:  []int{nethttp.StatusNotFound},
			latencyNotExceeding: time.Millisecond * 10,
		},
	}
	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			server := http.NewServer(tc.handlers...)
			assert.NotNil(t, server)
			t.Cleanup(func() {
				_ = server.Close()
			})
			// start a server in a goroutine.
			l, err := net.Listen("tcp", ":0")
			assert.NoError(t, err)
			go func() {
				_ = server.Serve(l)
			}()
			// wait for the server to start.
			time.Sleep(time.Second * 1)
			req, err := nethttp.NewRequest(tc.method, fmt.Sprintf("http://%s", l.Addr().String())+tc.path, bytes.NewBuffer(tc.body))
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
