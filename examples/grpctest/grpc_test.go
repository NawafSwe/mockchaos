package grpctest

import (
	"context"
	_ "embed"
	"sync"
	"testing"
	"time"

	"github.com/nawafswe/mockchaos/examples/grpctest/github.com/nawafswe/orders-service/proto"
	"github.com/nawafswe/mockchaos/grpctest"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/encoding/protojson"
)

// MockGetOrderGRPCServer mocks orders grpc	service.
type MockGetOrderGRPCServer struct {
	proto.UnimplementedOrderServiceServer
	mu       sync.Mutex
	response *proto.GetOrderResponse
	err      error
	sleep    time.Duration
}

func (m *MockGetOrderGRPCServer) GetOrder(_ context.Context, _ *proto.GetOrderRequest) (*proto.GetOrderResponse, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.sleep > time.Duration(0) {
		time.Sleep(m.sleep)
	}
	return m.response, m.err
}

// SetResponse configures the response to be returned by the server.
func (m *MockGetOrderGRPCServer) SetResponse(resp *proto.GetOrderResponse, err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.response = resp
	m.err = err
}

// Sleep configures the sleep behavior.
func (m *MockGetOrderGRPCServer) Sleep(sleep time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sleep = sleep
}

//go:embed testdata/order_response.json
var orderResponse []byte

func TestExampleGRPCTestUsage(t *testing.T) {
	t.Run("Should call grpc service to fetch orders", func(t *testing.T) {
		t.Parallel()
		var res proto.GetOrderResponse
		assert.NoError(t, protojson.Unmarshal(orderResponse, &res))
		mc := MockGetOrderGRPCServer{sleep: 100 * time.Millisecond, response: &res}
		svc := grpctest.NewServer(t, func(server *grpc.Server) {
			proto.RegisterOrderServiceServer(server, &mc)
		})

		ordersClient := proto.NewOrderServiceClient(svc.Client)
		response, err := ordersClient.GetOrder(context.Background(), &proto.GetOrderRequest{OrderId: 1}, grpc.WaitForReady(true))
		assert.NoError(t, err)
		assert.Equal(t, response.GetOrder().GetOrderId(), res.GetOrder().GetOrderId())
	})
	t.Run("Should call grpc service to fetch orders with failed response", func(t *testing.T) {
		t.Parallel()
		mc := MockGetOrderGRPCServer{}
		mc.SetResponse(nil, status.Errorf(codes.Internal, "Internal error"))
		mc.Sleep(time.Second)
		svc := grpctest.NewServer(t, func(server *grpc.Server) {
			proto.RegisterOrderServiceServer(server, &mc)
		})

		ordersClient := proto.NewOrderServiceClient(svc.Client)
		response, err := ordersClient.GetOrder(context.Background(), &proto.GetOrderRequest{OrderId: 1}, grpc.WaitForReady(true))
		assert.Error(t, err)
		assert.Nil(t, response)
	})
}
