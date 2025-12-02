package http

import (
	nethttp "net/http"

	"github.com/NawafSwe/gofi/internal/core"
)

// NewServer creates a new http server with given handlers and returns a function to start the server and a function to close the server.
func NewServer(handlers ...core.Handler) *nethttp.Server {
	handler := nethttp.HandlerFunc(core.NewHTTPHandler(handlers...).ServeHTTP)
	return &nethttp.Server{
		Handler: handler,
	}
}
