package core

import (
	"io"
	"math/rand/v2"
	nethttp "net/http"
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

// NewHTTPHandler builds a net/http.Handler from the given handler specs.
func NewHTTPHandler(handlers ...Handler) nethttp.Handler {
	handlersMap := make(map[string]Handler, len(handlers))
	for _, h := range handlers {
		handlersMap[handlerKey(h.Method, h.Path)] = h
	}

	return nethttp.HandlerFunc(func(w nethttp.ResponseWriter, r *nethttp.Request) {
		_, _ = io.ReadAll(r.Body)
		_ = r.Body.Close()

		hd, ok := handlersMap[handlerKey(r.Method, r.URL.Path)]
		if !ok {
			w.WriteHeader(defaultHTTPStatus)
			_, _ = w.Write(defaultResponse)
			return
		}

		randomStatus := hd.Statuses[rand.IntN(len(hd.Statuses))]
		if hd.Latencies != nil {
			time.Sleep(hd.Latencies[rand.IntN(len(hd.Latencies))])
		}

		w.WriteHeader(randomStatus)
		_, _ = w.Write(hd.Body)
	})
}

func handlerKey(method, path string) string {
	return method + " " + path
}
