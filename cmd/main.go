package main

import (
	"flag"
	"log"

	"github.com/nawafswe/mockchaos/cmd/svc"
)

func main() {
	mocksPath := flag.String("mocks_path", "", "path to mocks directory")
	httpPort := flag.Int("http_port", 8080, "port to run the http server on")
	mockServer := flag.String("mock_server", "http-svc", "service to run: http-svc or grpc-svc")
	flag.Parse()

	switch *mockServer {
	case "http-svc":
		svc.RunHTTPMock(*httpPort, *mocksPath)

	default:
		log.Fatalf("unknown service %q, must be one of: http-svc, grpc-svc", *mockServer)
	}
}
