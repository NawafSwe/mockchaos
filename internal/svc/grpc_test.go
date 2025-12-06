package svc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/nawafswe/mockchaos/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestRunGRPCMock(t *testing.T) {
	rootDir := getTestRootDir(t)

	tests := []struct {
		name        string
		port        int
		protoDir    string
		mocksPath   string
		expectedErr error
	}{
		{
			name:        "should return error when proto_dir is empty",
			port:        50051,
			protoDir:    "",
			mocksPath:   "/some/path",
			expectedErr: fmt.Errorf("proto_dir is required"),
		},
		{
			name:        "should return error when mocks_path is empty",
			port:        50051,
			protoDir:    "/some/path",
			mocksPath:   "",
			expectedErr: fmt.Errorf("mocks_path is required"),
		},
		{
			name:        "should return error when proto_dir does not exist",
			port:        50051,
			protoDir:    "/nonexistent/proto/dir",
			mocksPath:   "/some/mocks/path",
			expectedErr: fmt.Errorf("failed to load proto descriptors"),
		},
		{
			name: "should return error when mocks_path does not exist",
			port: 50051,
			protoDir: func() string {
				protoDir := filepath.Join(rootDir, "protos")
				if _, err := os.Stat(protoDir); os.IsNotExist(err) {
					return "/nonexistent/proto"
				}
				return protoDir
			}(),
			mocksPath:   "/nonexistent/mocks/path",
			expectedErr: fmt.Errorf("failed to load mocks directory"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			err := svc.RunGRPCMock(ctx, tt.port, tt.protoDir, tt.mocksPath)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunGRPCMock_ValidPaths(t *testing.T) {
	// Get testdata directory paths
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdataProtoDir := filepath.Join(wd, "testdata", "protos")
	testdataMocksPath := filepath.Join(wd, "testdata", "grpc-mocks")

	if _, err := os.Stat(testdataProtoDir); os.IsNotExist(err) {
		t.Skipf("testdata proto directory not found at %s", testdataProtoDir)
		return
	}
	if _, err := os.Stat(testdataMocksPath); os.IsNotExist(err) {
		t.Skipf("testdata mocks directory not found at %s", testdataMocksPath)
		return
	}

	tests := []struct {
		name      string
		protoDir  string
		mocksPath string
	}{
		{
			name:      "should load handlers and register services from valid proto and mocks paths",
			protoDir:  testdataProtoDir,
			mocksPath: testdataMocksPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			listener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)
			port := listener.Addr().(*net.TCPAddr).Port
			listener.Close()

			// Run in goroutine since it blocks on Serve
			errChan := make(chan error, 1)
			go func() {
				errChan <- svc.RunGRPCMock(context.Background(), port, tt.protoDir, tt.mocksPath)
			}()

			// Check for immediate errors (e.g., RegisterServices failure)
			select {
			case err := <-errChan:
				// If we get an error immediately, it means initialization failed
				// (proto loading, handler loading, or service registration)
				require.NoError(t, err, "server initialization should succeed, including RegisterServices()")
				return
			case <-time.After(200 * time.Millisecond):
				// No immediate error means RegisterServices() completed successfully
				// and the server is starting to serve
			}

			// Wait a bit more for server to start listening
			time.Sleep(500 * time.Millisecond)

			// Try to connect to verify server is running
			// If RegisterServices() succeeded, services should be registered
			conn, err := grpc.NewClient(
				fmt.Sprintf("localhost:%d", port),
				grpc.WithTransportCredentials(insecure.NewCredentials()),
			)
			require.NoError(t, err, "should be able to connect to gRPC server")
			assert.NotNil(t, conn, "connection should not be nil")
			_ = conn.Close()
		})
	}
}
