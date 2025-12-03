package grpc_test

import (
	"context"
	"testing"
	"time"

	coregrpc "github.com/nawafswe/mockchaos/core/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func TestInterceptor(t *testing.T) {
	tests := map[string]struct {
		handlers       map[string]coregrpc.Handler
		service        string
		method         string
		expectedCode   codes.Code
		expectedErrMsg string
		checkLatency   bool
		maxLatency     time.Duration
	}{
		"should return handler response when handler exists": {
			handlers: map[string]coregrpc.Handler{
				"orders.OrderService.GetOrder": {
					Service:     "orders.OrderService",
					Method:      "GetOrder",
					Response:    dynamicpb.NewMessage((&emptypb.Empty{}).ProtoReflect().Type().Descriptor()),
					StatusCodes: []codes.Code{codes.OK},
				},
			},
			service:      "orders.OrderService",
			method:       "GetOrder",
			expectedCode: codes.OK,
		},
		"should return unimplemented when handler not found": {
			handlers: map[string]coregrpc.Handler{
				"orders.OrderService.GetOrder": {
					Service:     "orders.OrderService",
					Method:      "GetOrder",
					StatusCodes: []codes.Code{codes.OK},
				},
			},
			service:        "orders.OrderService",
			method:         "GetOrderNotFound",
			expectedCode:   codes.Unimplemented,
			expectedErrMsg: "no handler found",
		},
		"should return error status when status code is not OK": {
			handlers: map[string]coregrpc.Handler{
				"orders.OrderService.GetOrder": {
					Service:      "orders.OrderService",
					Method:       "GetOrder",
					StatusCodes:  []codes.Code{codes.Internal},
					ErrorMessage: "mock error",
				},
			},
			service:        "orders.OrderService",
			method:         "GetOrder",
			expectedCode:   codes.Internal,
			expectedErrMsg: "mock error",
		},
		"should return error status when status code is not OK and default error msg when no error configured": {
			handlers: map[string]coregrpc.Handler{
				"orders.OrderService.GetOrder": {
					Service:     "orders.OrderService",
					Method:      "GetOrder",
					StatusCodes: []codes.Code{codes.Internal},
				},
			},
			service:        "orders.OrderService",
			method:         "GetOrder",
			expectedCode:   codes.Internal,
			expectedErrMsg: "mock error",
		},
		"should return not found no response is configured": {
			handlers:       map[string]coregrpc.Handler{"orders.OrderService.GetOrder": {}},
			service:        "orders.OrderService",
			method:         "GetOrder",
			expectedCode:   codes.Internal,
			expectedErrMsg: "no response configured",
		},
		"should apply latency when configured": {
			handlers: map[string]coregrpc.Handler{
				"orders.OrderService.GetOrder": {
					Service:     "orders.OrderService",
					Method:      "GetOrder",
					Response:    dynamicpb.NewMessage((&emptypb.Empty{}).ProtoReflect().Type().Descriptor()),
					StatusCodes: []codes.Code{codes.OK},
					Latencies:   []time.Duration{50 * time.Millisecond},
				},
			},
			service:      "orders.OrderService",
			method:       "GetOrder",
			expectedCode: codes.OK,
			checkLatency: true,
			maxLatency:   100 * time.Millisecond,
		},
		"should apply latency when configured and return ok when no status code is configured": {
			handlers: map[string]coregrpc.Handler{
				"orders.OrderService.GetOrder": {
					Service:   "orders.OrderService",
					Method:    "GetOrder",
					Response:  dynamicpb.NewMessage((&emptypb.Empty{}).ProtoReflect().Type().Descriptor()),
					Latencies: []time.Duration{50 * time.Millisecond},
				},
			},
			service:      "orders.OrderService",
			method:       "GetOrder",
			expectedCode: codes.OK,
			checkLatency: true,
			maxLatency:   100 * time.Millisecond,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			interceptor := coregrpc.Interceptor(tc.handlers)

			info := &grpc.UnaryServerInfo{
				FullMethod: "/" + tc.service + "/" + tc.method,
			}

			start := time.Now()
			resp, err := interceptor(
				context.Background(),
				nil,
				info,
				func(ctx context.Context, req interface{}) (interface{}, error) {
					return nil, nil
				},
			)
			elapsed := time.Since(start)

			if tc.expectedCode == codes.OK {
				assert.NoError(t, err)
				assert.NotNil(t, resp)
			} else {
				assert.Error(t, err)
				st, ok := status.FromError(err)
				assert.True(t, ok)
				assert.Equal(t, tc.expectedCode, st.Code())
				if tc.expectedErrMsg != "" {
					assert.Contains(t, st.Message(), tc.expectedErrMsg)
				}
			}

			if tc.checkLatency {
				assert.GreaterOrEqual(t, elapsed, 50*time.Millisecond)
				assert.LessOrEqual(t, elapsed, tc.maxLatency)
			}
		})
	}
}
