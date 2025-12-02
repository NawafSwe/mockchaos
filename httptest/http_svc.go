package httptest

import (
	"io"
	"math/rand/v2"
	nethttp "net/http"
	"net/http/httptest"
	"time"
)

const defaultHTTPStatus = nethttp.StatusNotFound

var defaultResponse = []byte(`{"error": "Not found"}`)

// Handler represents a http handler.
type Handler struct {
	Path      string
	Method    string
	Body      []byte
	Statuses  []int
	Latencies []time.Duration
}

// Server represents an http server registered with handlers.
type Server struct {
	srv *httptest.Server
}

// NewServer creates a new http server with given handlers.
func NewServer(handlers ...Handler) *Server {
	svc := Server{}

	handlersMap := make(map[string]Handler, len(handlers))
	for _, handler := range handlers {
		constructedKey := handlerKey(handler.Method, handler.Path)
		handlersMap[constructedKey] = handler
	}

	svc.srv = httptest.NewServer(nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()
		hd, ok := handlersMap[handlerKey(r.Method, r.URL.Path)]
		if !ok {
			w.WriteHeader(defaultHTTPStatus)
			_, _ = w.Write(defaultResponse)
			return
		}
		randomStatus := hd.Statuses[rand.IntN(len(hd.Statuses))]
		// simulation of latency.
		if hd.Latencies != nil {
			time.Sleep(hd.Latencies[rand.IntN(len(hd.Latencies))])
		}
		w.WriteHeader(randomStatus)
		_, _ = w.Write(hd.Body)
		return
	}))

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

func handlerKey(method, path string) string {
	return method + " " + path
}
