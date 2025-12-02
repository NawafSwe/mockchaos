package http

import (
	"io"
	"math/rand/v2"
	nethttp "net/http"
	"time"
)

// defaultHTTPStatus is the HTTP status code returned when no handler matches a request.
const defaultHTTPStatus = nethttp.StatusNotFound

// defaultResponse is the response body returned when no handler matches a request.
var defaultResponse = []byte(`{"error": "Not found"}`)

// Handler represents a mock HTTP endpoint configuration.
//
// Each Handler defines how to respond to requests matching a specific path and HTTP method.
// The handler supports randomized responses to simulate real-world variability in API behavior.
type Handler struct {
	// Path is the URL path pattern to match (e.g., "/api/users").
	// Exact match is performed - no wildcards or regex support.
	Path string

	// Method is the HTTP method to match (e.g., "GET", "POST", "PUT", "DELETE").
	// Must be a valid HTTP method as defined by RFC 7231.
	Method string

	// Body is the response body to return.
	// Can be any byte slice - JSON, XML, plain text, etc.
	Body []byte

	// Statuses is a slice of HTTP status codes to randomly select from.
	// At least one status code must be provided for the handler to be effective.
	// If multiple statuses are provided, one is randomly selected per request.
	Statuses []int

	// Headers are optional HTTP headers to include in the response.
	// Can be nil or empty if no custom headers are needed.
	Headers map[string]string

	// Latencies is an optional slice of durations to simulate network latency.
	// If provided, one duration is randomly selected and applied before sending the response.
	// Can be nil or empty to disable latency simulation.
	Latencies []time.Duration
}

// NewHTTPHandler builds a net/http.Handler from the given handler specifications.
//
// It creates a single http.Handler that routes requests to the appropriate mock handler
// based on the request's method and path. If no handler matches, it returns a 404 response.
func NewHTTPHandler(handlers ...Handler) nethttp.Handler {
	handlersMap := make(map[string]Handler, len(handlers))
	for _, h := range handlers {
		handlersMap[handlerKey(h.Method, h.Path)] = h
	}

	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()

		hd, ok := handlersMap[handlerKey(r.Method, r.URL.Path)]
		if !ok {
			w.WriteHeader(defaultHTTPStatus)
			_, _ = w.Write(defaultResponse)
			return
		}

		randomStatus := status(hd.Statuses)
		if hd.Latencies != nil {
			time.Sleep(hd.Latencies[rand.IntN(len(hd.Latencies))])
		}

		for k, v := range hd.Headers {
			w.Header().Set(k, v)
		}
		w.WriteHeader(randomStatus)
		_, _ = w.Write(hd.Body)
	})
}

// status returns a random status code from the given slice of integers.
func status(statuses []int) int {
	if len(statuses) == 0 {
		return nethttp.StatusOK
	}
	return statuses[rand.IntN(len(statuses))]
}

// handlerKey constructs a unique key for a method+path combination.
// Format: "METHOD PATH" (e.g., "GET /api/users").
func handlerKey(method, path string) string {
	return method + " " + path
}
