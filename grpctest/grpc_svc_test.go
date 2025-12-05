package grpctest

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"
)

func TestGRPCServer(t *testing.T) {
	t.Run("should successfully creates a grpc server for test", func(t *testing.T) {
		t.Parallel()
		srv := NewServer(t, func(server *grpc.Server) {
		})
		assert.NotNil(t, srv)
		assert.NotEmpty(t, srv.Addr())
	})
}
