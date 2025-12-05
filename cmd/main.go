package main

import (
	"context"
	"flag"
	"log"

	"github.com/nawafswe/mockchaos/cmd/svc"
)

func main() {
	mocksPath := flag.String("mocks_path", "", "path to mocks directory")
	httpPort := flag.Int("http_port", 8080, "port to run the http server on")
	grpcPort := flag.Int("grpc_port", 50051, "port to run the gRPC server on")
	grpcProtoDir := flag.String("grpc_proto_dir", "", "path to directory containing .proto files")
	mockServer := flag.String("mock_server", "http-svc", "service to run: http-svc or grpc-svc")
	flag.Parse()
	context, cancel := context.WithCancel(context.Background())
	defer cancel()
	switch *mockServer {
	case "http-svc":
		svc.RunHTTPMock(context, *httpPort, *mocksPath)
	case "grpc-svc":
		if *grpcProtoDir == "" {
			log.Fatal("grpc_proto_dir is required when using grpc-svc")
		}
		svc.RunGRPCMock(context, *grpcPort, *grpcProtoDir, *mocksPath)
	default:
		log.Fatalf("unknown service %q, must be one of: http-svc, grpc-svc", *mockServer)
	}
}
