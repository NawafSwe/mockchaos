package grpc

import (
	"math/rand/v2"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/reflect/protoreflect"
)

// Handler represents a mock gRPC method configuration.
type Handler struct {
	// Service is the fully qualified service name (e.g., "orders.OrderService").
	Service string

	// Method is the RPC method name.
	Method string

	// Response is the response message as a protobuf message.
	Response protoreflect.Message

	// ResponseJSON is the response body as JSON (alternative to Response).
	// Will be converted to a protobuf message if Response is nil.
	ResponseJSON map[string]any

	// StatusCodes is a slice of gRPC status codes to randomly select from.
	// At least one status code must be provided.
	// If multiple codes are provided, one is randomly selected per request.
	// Use codes.OK for success, codes.INTERNAL for errors, etc.
	StatusCodes []codes.Code

	// Latencies is an optional slice of durations to simulate network latency.
	Latencies []time.Duration

	// ErrorMessage is an optional error message to return if statusCode is not OK.
	ErrorMessage string
}

// HandlerKey constructs a unique key for a service+method combination.
func HandlerKey(service, method string) string {
	return service + "." + method
}

// randomStatusCode returns a random status code from the given slice.
func randomStatusCode(c []codes.Code) codes.Code {
	if len(c) == 0 {
		return codes.OK
	}
	return c[rand.IntN(len(c))]
}

// randomLatency returns a random latency from the given slice, or 0 if empty.
func randomLatency(latencies []time.Duration) time.Duration {
	if len(latencies) == 0 {
		return 0
	}
	return latencies[rand.IntN(len(latencies))]
}
