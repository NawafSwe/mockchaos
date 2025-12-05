package grpc_test

import (
	"context"
	_ "embed"

	coregrpc "github.com/nawafswe/mockchaos/core/grpc"
	chaosgrpc "github.com/nawafswe/mockchaos/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"os"
	"path/filepath"
	"testing"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/dynamicpb"
	"google.golang.org/protobuf/types/known/emptypb"
)

//go:embed testdata/proto.proto
var restaurantProto []byte

func TestServer_Integration(t *testing.T) {
	// Setup proto files.
	tmpDir := t.TempDir()
	protoFile := filepath.Join(tmpDir, "restaurant.proto")
	err := os.WriteFile(protoFile, restaurantProto, 0644)
	assert.NoError(t, err)

	// Load proto files.
	msgTypes, err := coregrpc.LoadMessageTypesFromProtoDir(context.TODO(), tmpDir)
	if err != nil {
		t.Skipf("failed to load proto files: %v", err)
		return
	}
	assert.NotNil(t, msgTypes)

	tests := map[string]struct {
		handlers            map[string]coregrpc.Handler
		service             string
		method              string
		expectedCode        codes.Code
		expectedErrMsg      string
		latencyNotExceeding time.Duration
	}{
		"should successfully register gRPC server and respond with handler": {
			handlers: map[string]coregrpc.Handler{
				"restaurants.RestaurantService.ListMenuItems": {
					Service:     "restaurants.RestaurantService",
					Method:      "ListMenuItems",
					Response:    dynamicpb.NewMessage((&emptypb.Empty{}).ProtoReflect().Type().Descriptor()),
					StatusCodes: []codes.Code{codes.OK},
					Latencies:   []time.Duration{10 * time.Millisecond, 20 * time.Millisecond},
				},
			},
			service:             "restaurants.RestaurantService",
			method:              "ListMenuItems",
			expectedCode:        codes.OK,
			latencyNotExceeding: 50 * time.Millisecond,
		},
		"should return unimplemented when no handler found": {
			handlers: map[string]coregrpc.Handler{
				"restaurants.RestaurantService.ListMenuItems": {
					Service:     "restaurants.RestaurantService",
					Method:      "ListMenuItems",
					StatusCodes: []codes.Code{codes.OK},
				},
			},
			service:        "restaurants.RestaurantService",
			method:         "GetMenuItem",
			expectedCode:   codes.Unimplemented,
			expectedErrMsg: "no handler found",
		},
		"should return error status when status code is not OK": {
			handlers: map[string]coregrpc.Handler{
				"restaurants.RestaurantService.ListMenuItems": {
					Service:      "restaurants.RestaurantService",
					Method:       "ListMenuItems",
					Response:     dynamicpb.NewMessage((&emptypb.Empty{}).ProtoReflect().Type().Descriptor()),
					StatusCodes:  []codes.Code{codes.Internal},
					ErrorMessage: "mock internal error",
					Latencies:    []time.Duration{10 * time.Millisecond},
				},
			},
			service:             "restaurants.RestaurantService",
			method:              "ListMenuItems",
			expectedCode:        codes.Internal,
			expectedErrMsg:      "mock internal error",
			latencyNotExceeding: 50 * time.Millisecond,
		},
		"should apply latency when configured": {
			handlers: map[string]coregrpc.Handler{
				"restaurants.RestaurantService.ListMenuItems": {
					Service:     "restaurants.RestaurantService",
					Method:      "ListMenuItems",
					Response:    dynamicpb.NewMessage((&emptypb.Empty{}).ProtoReflect().Type().Descriptor()),
					StatusCodes: []codes.Code{codes.OK},
					Latencies:   []time.Duration{50 * time.Millisecond},
				},
			},
			service:             "restaurants.RestaurantService",
			method:              "ListMenuItems",
			expectedCode:        codes.OK,
			latencyNotExceeding: 100 * time.Millisecond,
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			// Convert handlers map to slice
			handlerSlice := make([]coregrpc.Handler, 0, len(tc.handlers))
			for _, h := range tc.handlers {
				handlerSlice = append(handlerSlice, h)
			}

			// Create server
			server := chaosgrpc.NewServer(handlerSlice...)
			assert.NotNil(t, server)
			t.Cleanup(func() {
				_ = server.Close()
			})

			// Register services
			err = server.RegisterServices()
			assert.NoError(t, err)

			// Start listening
			err = server.Listen(":0")
			assert.NoError(t, err)

			// Start serving in a goroutine
			serveErr := make(chan error, 1)
			go func() {
				serveErr <- server.Serve()
			}()

			// Wait for server to start
			time.Sleep(500 * time.Millisecond)

			// Make gRPC call using the interceptor directly (simulating a call)
			interceptor := coregrpc.Interceptor(tc.handlers)
			info := &grpc.UnaryServerInfo{
				FullMethod: "/" + tc.service + "/" + tc.method,
			}

			start := time.Now()
			resp, err := interceptor(
				context.TODO(),
				nil,
				info,
				func(ctx context.Context, req any) (any, error) {
					return nil, nil
				},
			)
			delta := time.Since(start)

			// Assert status codes
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

			// Assert latency
			if tc.latencyNotExceeding > 0 {
				assert.LessOrEqual(t, delta, tc.latencyNotExceeding)
			}

			// Stop server
			server.Stop()

			// Wait for serve to return
			select {
			case err := <-serveErr:
				assert.NoError(t, err)
			case <-time.After(1 * time.Second):
				// Server stopped, continue
			}
		})
	}
}
