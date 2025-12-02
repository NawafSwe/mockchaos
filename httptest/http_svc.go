package httptest

import (
	"net/http/httptest"

	"github.com/NawafSwe/gofi/internal/core"
)

// Server represents an http server registered with handlers.
type Server struct {
	srv *httptest.Server
}

// NewServer creates a new http server with given handlers.
func NewServer(handlers ...core.Handler) *Server {
	svc := Server{}
	svc.srv = httptest.NewServer(core.NewHTTPHandler(handlers...))
	return &svc
}

func (s *Server) URL() string {
	return s.srv.URL
}

func (s *Server) Close() {
	s.srv.Client().CloseIdleConnections()
	s.srv.CloseClientConnections()
	s.srv.Close()
}
