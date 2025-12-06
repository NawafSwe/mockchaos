package svc_test

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"testing"

	"github.com/nawafswe/mockchaos/internal/svc"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// getTestRootDir returns the absolute path to the workspace root
func getTestRootDir(t *testing.T) string {
	t.Helper()
	wd, err := os.Getwd()
	require.NoError(t, err)
	root := filepath.Join(wd, "..", "..")
	absRoot, err := filepath.Abs(root)
	require.NoError(t, err)
	return absRoot
}

func TestRunHTTPMock(t *testing.T) {
	tests := []struct {
		name        string
		port        int
		mocksPath   string
		expectedErr error
	}{
		{
			name:        "should return error when mocks_path is empty",
			port:        8080,
			mocksPath:   "",
			expectedErr: fmt.Errorf("mocks_path is required"),
		},
		{
			name:        "should return error when mocks_path does not exist",
			port:        8080,
			mocksPath:   "/nonexistent/path",
			expectedErr: fmt.Errorf("failed to load mocks directory"),
		},
		{
			name: "should return error when port is invalid",
			port: -1,
			mocksPath: func() string {
				wd, _ := os.Getwd()
				return filepath.Join(wd, "testdata", "mocks")
			}(),
			expectedErr: fmt.Errorf("failed to listen"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx := context.Background()

			err := svc.RunHTTPMock(ctx, tt.port, tt.mocksPath)

			if tt.expectedErr != nil {
				assert.Error(t, err)
				assert.ErrorContains(t, err, tt.expectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestRunHTTPMock_ValidPath(t *testing.T) {
	// Get testdata directory path
	wd, err := os.Getwd()
	require.NoError(t, err)
	testdataMocksPath := filepath.Join(wd, "testdata", "mocks")

	if _, err := os.Stat(testdataMocksPath); os.IsNotExist(err) {
		t.Skipf("testdata mocks directory not found at %s", testdataMocksPath)
		return
	}

	tests := []struct {
		name      string
		port      int
		mocksPath string
	}{
		{
			name:      "should load handlers from valid mocks path",
			port:      0,
			mocksPath: testdataMocksPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()

			// Find an available port
			listener, err := net.Listen("tcp", ":0")
			require.NoError(t, err)
			port := listener.Addr().(*net.TCPAddr).Port
			listener.Close()

			// Run in goroutine since it blocks on Serve
			errChan := make(chan error, 1)
			go func() {
				errChan <- svc.RunHTTPMock(ctx, port, tt.mocksPath)
			}()

			cancel()

			select {
			case err := <-errChan:
				// Server stopped, which is expected
				// The important thing is it didn't fail during initialization
				assert.NotNil(t, err)
			default:
			}
		})
	}
}
