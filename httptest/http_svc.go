// Package httptest provides a wrapper around httptest.Server
package httptest

import (
	"net/http/httptest"
	"testing"

	"github.com/nawafswe/mockchaos/core/http"
)

// Server represents an http server registered with handlers.
type Server struct {
	srv *httptest.Server
}

// NewServer creates a new http server with given handlers.
// The server is automatically started and will be cleaned up when the test completes.
func NewServer(t *testing.T, handlers ...http.Handler) *Server {
	t.Helper()
	svc := Server{}
	svc.srv = httptest.NewServer(http.NewHTTPHandler(handlers...))
	t.Cleanup(func() {
		svc.Close()
	})
	return &svc
}

// URL returns the base URL of the test server.
// This is useful for making requests to the mock server in tests.
func (s *Server) URL() string {
	return s.srv.URL
}

// Close shuts down the test server and closes all idle connections.
// It should be called when the test server is no longer needed, typically in defer or t cleanup function.
func (s *Server) Close() {
	s.srv.Client().CloseIdleConnections()
	s.srv.CloseClientConnections()
	s.srv.Close()
}
