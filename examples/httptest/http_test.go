package httptest

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"io"
	nethttp "net/http"
	"testing"
	"time"

	"github.com/nawafswe/mockchaos/core/http"
	httptestpkg "github.com/nawafswe/mockchaos/httptest"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/users_handlers.json
var usersHandlersJSON []byte

func TestExampleHTTPTestUsage(t *testing.T) {
	tests := []struct {
		name                 string
		handlers             []http.Handler
		method               string
		path                 string
		requestBody          []byte
		expectedStatusCodes  []int
		expectedBodyContains string
		minLatency           time.Duration
		maxLatency           time.Duration
		validateJSON         bool
		validateHeader       map[string]string
	}{
		{
			name: "programmatic handlers with latency",
			handlers: []http.Handler{
				{
					Path:      "/api/v1/users",
					Method:    nethttp.MethodGet,
					Response:  []byte(`{"id": 1, "name": "Alice", "email": "alice@example.com"}`),
					Statuses:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
					Latencies: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
					Headers:   map[string]string{"Content-Type": "application/json"},
				},
			},
			method:               nethttp.MethodGet,
			path:                 "/api/v1/users",
			expectedStatusCodes:  []int{nethttp.StatusOK, nethttp.StatusInternalServerError},
			expectedBodyContains: "Alice",
			minLatency:           90 * time.Millisecond,
		},
		{
			name: "multiple latency options",
			handlers: []http.Handler{
				{
					Path:     "/api/v1/slow",
					Method:   nethttp.MethodGet,
					Response: []byte(`{"status": "ok"}`),
					Statuses: []int{nethttp.StatusOK},
					Latencies: []time.Duration{
						50 * time.Millisecond,
						100 * time.Millisecond,
						200 * time.Millisecond,
						300 * time.Millisecond,
					},
				},
			},
			method:               nethttp.MethodGet,
			path:                 "/api/v1/slow",
			expectedStatusCodes:  []int{nethttp.StatusOK},
			expectedBodyContains: "ok",
			minLatency:           40 * time.Millisecond,
			maxLatency:           400 * time.Millisecond,
		},
		{
			name: "404 for unregistered paths",
			handlers: []http.Handler{
				{
					Path:     "/api/v1/users",
					Method:   nethttp.MethodGet,
					Response: []byte(`{"users": []}`),
					Statuses: []int{nethttp.StatusOK},
				},
			},
			method:               nethttp.MethodGet,
			path:                 "/api/v1/unknown",
			expectedStatusCodes:  []int{nethttp.StatusNotFound},
			expectedBodyContains: "Not found",
		},
		{
			name: "POST requests with body",
			handlers: []http.Handler{
				{
					Path:      "/api/v1/users",
					Method:    nethttp.MethodPost,
					Response:  []byte(`{"id": 123, "name": "Bob", "created": true}`),
					Statuses:  []int{nethttp.StatusCreated, nethttp.StatusBadRequest},
					Latencies: []time.Duration{50 * time.Millisecond, 100 * time.Millisecond},
					Headers:   map[string]string{"Content-Type": "application/json"},
				},
			},
			method:               nethttp.MethodPost,
			path:                 "/api/v1/users",
			requestBody:          []byte(`{"name": "Bob", "email": "bob@example.com"}`),
			expectedStatusCodes:  []int{nethttp.StatusCreated, nethttp.StatusBadRequest},
			expectedBodyContains: "Bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			svc := httptestpkg.NewServer(t, tt.handlers...)
			defer svc.Close()

			client := nethttp.Client{Timeout: 5 * time.Second}
			var bodyReader io.Reader
			if tt.requestBody != nil {
				bodyReader = bytes.NewReader(tt.requestBody)
			}

			req, err := nethttp.NewRequest(tt.method, svc.URL()+tt.path, bodyReader)
			assert.NoError(t, err)
			if tt.requestBody != nil {
				req.Header.Set("Content-Type", "application/json")
			}

			start := time.Now()
			resp, err := client.Do(req)
			latency := time.Since(start)
			assert.NoError(t, err)
			defer func() {
				_ = resp.Body.Close()
			}()

			assert.Contains(t, tt.expectedStatusCodes, resp.StatusCode)

			if tt.minLatency > 0 {
				assert.GreaterOrEqual(t, latency, tt.minLatency)
			}
			if tt.maxLatency > 0 {
				assert.LessOrEqual(t, latency, tt.maxLatency)
			}

			body, err := io.ReadAll(resp.Body)
			assert.NoError(t, err)
			assert.Contains(t, string(body), tt.expectedBodyContains)

			if tt.validateJSON {
				var jsonResponse map[string]interface{}
				err = json.Unmarshal(body, &jsonResponse)
				assert.NoError(t, err)
			}

			for key, value := range tt.validateHeader {
				assert.Equal(t, value, resp.Header.Get(key))
			}
		})
	}
}

func TestExampleHTTPTestUsageWithJSON(t *testing.T) {
	handlers, err := http.ParseHandlers(usersHandlersJSON)
	assert.NoError(t, err)
	assert.NotEmpty(t, handlers)

	svc := httptestpkg.NewServer(t, handlers...)
	defer svc.Close()

	client := nethttp.Client{Timeout: 5 * time.Second}
	req, err := nethttp.NewRequest(nethttp.MethodGet, svc.URL()+"/api/v1/users", nil)
	assert.NoError(t, err)

	resp, err := client.Do(req)
	assert.NoError(t, err)
	defer func() {
		_ = resp.Body.Close()
	}()

	assert.Contains(t, []int{nethttp.StatusOK, nethttp.StatusInternalServerError}, resp.StatusCode)
	assert.Equal(t, "application/json", resp.Header.Get("Content-Type"))

	body, err := io.ReadAll(resp.Body)
	assert.NoError(t, err)

	var usersResponse map[string]interface{}
	err = json.Unmarshal(body, &usersResponse)
	assert.NoError(t, err)
	assert.NotNil(t, usersResponse["users"])
}
