package http

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// validMethods is a map of valid HTTP methods.
var validMethods = map[string]bool{
	http.MethodGet:     true,
	http.MethodPost:    true,
	http.MethodPut:     true,
	http.MethodDelete:  true,
	http.MethodPatch:   true,
	http.MethodHead:    true,
	http.MethodOptions: true,
	http.MethodTrace:   true,
	http.MethodConnect: true,
}

// handler represents a single handler specification in JSON format.
// This is the internal representation used during JSON unmarshalling.
type handler struct {
	Path      string            `json:"path"`
	Method    string            `json:"method"`
	Response  map[string]any    `json:"response"`
	Statuses  []int             `json:"statuses"`
	Latencies []string          `json:"latencies"`
	Headers   map[string]string `json:"headers"`
}

// ParseHandlers parses a JSON byte slice containing handler configurations and returns
// a slice of Handler structs ready for use with NewHTTPHandler.
//
// The JSON format must be an array of handler objects, where each handler has:
//   - path: string (required) - The URL path pattern
//   - method: string (required) - HTTP method (GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT)
//   - body: object (optional) - Response body as a JSON object (will be marshaled to bytes)
//   - statuses: array of integers (required) - HTTP status codes to randomly select from
//   - latencies: array of strings (optional) - Duration strings in Go format (e.g., "100ms", "1s")
//   - headers: object (optional) - Key-value pairs of HTTP headers
func ParseHandlers(content []byte) ([]Handler, error) {
	var handlers []handler
	if err := json.Unmarshal(content, &handlers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handlers: %w", err)
	}
	return toHandlers(handlers)
}

// toHandlers converts a slice of internal handler structs (from JSON) to exported Handler structs.
// It validates HTTP methods and parses latency strings into time.Duration values.
func toHandlers(handlers []handler) ([]Handler, error) {
	httpHandlers := make([]Handler, len(handlers))
	for i, h := range handlers {
		if !validHTTPMethod(h.Method) {
			return nil, fmt.Errorf("invalid http method: %s for path %s", h.Method, h.Path)
		}
		l, err := latencies(h.Latencies)
		if err != nil {
			return nil, err
		}
		marshalledBody, err := json.Marshal(h.Response)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal body: %w", err)
		}
		httpHandlers[i] = Handler{
			Path:      h.Path,
			Method:    h.Method,
			Statuses:  h.Statuses,
			Response:  marshalledBody,
			Latencies: l,
			Headers:   h.Headers,
		}
	}
	return httpHandlers, nil
}

// validHTTPMethod checks if the given method is a valid HTTP method according to RFC 7231.
// Supported methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT.
func validHTTPMethod(method string) bool {
	return validMethods[method]
}

// latencies converts a slice of duration strings (e.g., "100ms", "1s") to time.Duration values.
// Each string must be a valid Go duration format as parsed by time.ParseDuration.
//
// Returns an error if any duration string cannot be parsed.
func latencies(latencies []string) ([]time.Duration, error) {
	var err error
	durations := make([]time.Duration, len(latencies))
	for i, l := range latencies {
		durations[i], err = time.ParseDuration(l)
		if err != nil {
			return nil, fmt.Errorf("failed to parse latency: %w", err)
		}
	}
	return durations, nil
}
