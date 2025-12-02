// Package http provides http server functionality.
package http

import (
	nethttp "net/http"

	"github.com/NawafSwe/gofi/internal/core/http"
)

// NewServer creates a new http server with given handlers and returns a function to start the server and a function to close the server.
func NewServer(handlers ...http.Handler) *nethttp.Server {
	return &nethttp.Server{
		Handler: http.NewHTTPHandler(handlers...),
	}
}
