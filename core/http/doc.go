// Package http provides the core HTTP mocking functionality for mockchaos.
//
// This package implements the shared logic for building HTTP mock handlers that can be used
// both in-process (via httptest) and as standalone servers (for Kubernetes deployments).
// It provides a unified way to define mock HTTP endpoints with configurable responses,
// status codes, latencies, and headers.
//
// # Core Concepts
//
// A Handler represents a single mock endpoint configuration:
//   - Path and Method: Define the route (e.g., GET /api/users)
//   - Response: The response body to return
//   - Statuses: One or more HTTP status codes to randomly select from
//   - Latencies: Optional delays to simulate network latency
//   - Headers: Custom HTTP headers to include in the response
//
// # Usage
//
// Handlers can be created programmatically:
//
//	handlers := []http.Handler{
//	    {
//	        Path:     "/api/users",
//	        Method:   "GET",
//	        Response:     []byte(`{"id": 1, "name": "Alice"}`),
//	        Statuses: []int{200, 500},
//	        Latencies: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
//	        Headers:  map[string]string{"Content-Type": "application/json"},
//	    },
//	}
//	handler := http.NewHTTPHandler(handlers...)
//
// Or loaded from JSON configuration files:
//
//	content, _ := os.ReadFile("mocks/handlers.json")
//	handlers, err := http.ParseHandlers(content)
//	if err != nil {
//	    log.Fatal(err)
//	}
//	handler := http.NewHTTPHandler(handlers...)
//
// # Behavior
//
// When a request matches a registered handler:
//   - A status code is randomly selected from the Statuses slice
//   - If Latencies is non-empty, a random latency is applied
//   - Headers are set on the response
//   - Response is written to the response
//
// When no handler matches:
//   - Returns 404 Not Found with body: {"error": "Not found"}
//
// # JSON Format
//
// Handlers can be defined in JSON with the following structure:
//
//	[
//	    {
//	        "path": "/api/v1/users",
//	        "method": "GET",
//	        "body": {"id": 1, "name": "Alice"},
//	        "statuses": [200, 500],
//	        "latencies": ["100ms", "200ms"],
//	        "headers": {"Content-Type": "application/json"}
//	    }
//	]
//
// Latencies must be valid Go duration strings (e.g., "100ms", "1s", "500ms").
// Methods must be valid HTTP methods: GET, POST, PUT, DELETE, PATCH, HEAD, OPTIONS, TRACE, CONNECT.
package http
