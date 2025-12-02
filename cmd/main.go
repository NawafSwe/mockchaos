package main

import (
	"flag"
	"log"

	"github.com/NawafSwe/gofi/cmd/gofi"
)

func main() {
	mocksPath := flag.String("mocks_path", "", "path to mocks directory")
	httpPort := flag.Int("http_port", 8080, "port to run the http server on")
	mockServer := flag.String("mock_server", "http-svc", "service to run: http-svc or grpc-svc")
	flag.Parse()

	switch *mockServer {
	case "http-svc":
		gofi.RunHTTPMock(*httpPort, *mocksPath)

	default:
		log.Fatalf("Gofi: unknown service %q, must be one of: http-svc, grpc-svc", *mockServer)
	}
}
