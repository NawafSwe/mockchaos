package grpc

import (
	"context"
	"fmt"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

const (
	defaultErrMsg = "mock error"
)

// Interceptor is a gRPC unary server interceptor that routes requests to mocked handlers provided in the handlers map.
// It extracts the service and method name from the request, applies mock latency and status codes, and returns mock responses.
// If no handler is found for the given method, it returns an Unimplemented status error.
// The function supports latency simulation, error simulation, and protobuf-based or JSON-based mock responses.
func Interceptor(handlers map[string]Handler) func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		service, method := parseServiceMethod(info.FullMethod)
		key := service + "." + method

		// Look up handler
		h, ok := handlers[key]
		if !ok {
			return nil, status.Error(codes.Unimplemented,
				fmt.Sprintf("no handler found for %s.%s", service, method))
		}

		if latency := randomLatency(h.Latencies); latency > 0 {
			time.Sleep(latency)
		}
		code := randomStatusCode(h.StatusCodes)
		if code != codes.OK {
			errMsg := h.ErrorMessage
			if errMsg == "" {
				errMsg = defaultErrMsg
			}
			return nil, status.Error(code, errMsg)
		}

		if h.Response != nil {
			return h.Response.Interface(), nil
		}

		return nil, status.Error(codes.Internal, "no response configured")
	}
}

// parseServiceMethod extracts the service and method name from the full method name.
func parseServiceMethod(fullMethod string) (service, method string) {
	if len(fullMethod) > 0 && fullMethod[0] == '/' {
		fullMethod = fullMethod[1:]
	}
	lastSlash := strings.LastIndex(fullMethod, "/")
	if lastSlash == -1 {
		return "", fullMethod
	}
	service = fullMethod[:lastSlash]
	method = fullMethod[lastSlash+1:]
	return service, method
}
