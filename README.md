# Mockchaos

[![Build Status](https://github.com/nawafswe/mockchaos/actions/workflows/test.yaml/badge.svg)](https://github.com/nawafswe/mockchaos/actions)
[![Coverage](https://codecov.io/gh/nawafswe/mockchaos/branch/main/graph/badge.svg)](https://codecov.io/gh/nawafswe/mockchaos)
[![Go Report Card](https://goreportcard.com/badge/github.com/nawafswe/mockchaos)](https://goreportcard.com/report/github.com/nawafswe/mockchaos)
[![Go Version](https://img.shields.io/badge/go-1.25.3-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/license-Apache%202.0-blue.svg)](LICENSE)
[![Release](https://img.shields.io/github/v/release/nawafswe/mockchaos)](https://github.com/nawafswe/mockchaos/releases)
[![GitHub Stars](https://img.shields.io/github/stars/nawafswe/mockchaos?style=social)](https://github.com/nawafswe/mockchaos/stargazers)
[![GitHub Forks](https://img.shields.io/github/forks/nawafswe/mockchaos?style=social)](https://github.com/nawafswe/mockchaos/network/members)

**Mockchaos** is a Go library for creating mock HTTP and gRPC servers with configurable latencies, status codes, and responses. It's designed for chaos engineering, integration testing, and simulating real-world API behavior in development environments.

## Features

- 🚀 **HTTP & gRPC Mocking**: Create mock servers for both HTTP REST APIs and gRPC services
- ⏱️ **Latency Simulation**: Configure random latencies to simulate network delays
- 🎲 **Randomized Responses**: Randomly select from multiple status codes and latencies
- 📝 **JSON Configuration**: Define mocks using simple JSON files
- 🧪 **Test Utilities**: Built-in test helpers for integration testing
- 🐳 **Standalone Servers**: Run as standalone servers for Kubernetes deployments
- 🔄 **Dynamic gRPC**: Automatically register services from `.proto` files

## Installation

```bash
go get github.com/nawafswe/mockchaos
```

## When to Use Mockchaos

Mockchaos is ideal for:

- **Integration Testing**: Test your services against mock dependencies with realistic latencies
- **Chaos Engineering**: Simulate network delays, errors, and failures
- **Local Development**: Replace external services with mocks during development
- **CI/CD Pipelines**: Run integration tests without depending on external services
- **Kubernetes Deployments**: Deploy mock servers as sidecars or standalone services

## Quick Start

### HTTP Mocking

#### In Integration Tests

```go
package mytest

import (
    "net/http"
    "testing"
    "time"
    
    "github.com/nawafswe/mockchaos/core/http"
    "github.com/nawafswe/mockchaos/httptest"
)

func TestMyService(t *testing.T) {
    handlers := []http.Handler{
        {
            Path:      "/api/users",
            Method:    http.MethodGet,
            Response:  []byte(`{"id": 1, "name": "Alice"}`),
            Statuses:  []int{200, 500},  // Randomly returns 200 or 500
            Latencies: []time.Duration{100 * time.Millisecond, 200 * time.Millisecond},
            Headers:   map[string]string{"Content-Type": "application/json"},
        },
    }
    
    svc := httptest.NewServer(t, handlers...)
    defer svc.Close()
    
    // Use svc.URL() to make requests to the mock server
    resp, err := http.Get(svc.URL() + "/api/users")
    // ...
}
```

#### From JSON Configuration

```go
package main

import (
    "log"
    "os"
    "github.com/nawafswe/mockchaos/core/http"
)

func main() {
    content, err := os.ReadFile("mocks/handlers.json")
    if err != nil {
        log.Fatal(err)
    }
    
    handlers, err := http.ParseHandlers(content)
    if err != nil {
        log.Fatal(err)
    }
    
    handler := http.NewHTTPHandler(handlers...)
    // Use with your HTTP server
}
```

#### JSON Format for HTTP

```json
[
  {
    "path": "/api/v1/users",
    "method": "GET",
    "response": {
      "id": 1,
      "name": "Alice"
    },
    "statuses": [200, 500],
    "latencies": ["100ms", "200ms", "500ms"],
    "headers": {
      "Content-Type": "application/json"
    }
  }
]
```

### gRPC Mocking

#### In Integration Tests

See `examples/grpctest/grpc_test.go` for a complete example:

```go
package grpctest

import (
    "context"
    "testing"
    "time"
    
    "github.com/nawafswe/mockchaos/grpctest"
    "google.golang.org/grpc"
    "your-service/proto"
)

func TestMyGRPCService(t *testing.T) {
    svc := grpctest.NewServer(t, func(server *grpc.Server) {
        proto.RegisterMyServiceServer(server, &myMockServer{})
    })
    defer svc.Close()
    
    client := proto.NewMyServiceClient(svc.Client)
    resp, err := client.GetData(context.Background(), &proto.GetDataRequest{})
    // ...
}
```

#### From JSON Configuration

```json
[
  {
    "service": "orders.OrderService",
    "method": "GetOrder",
    "response_proto_type": "orders.GetOrderResponse",
    "response": {
      "order": {
        "orderId": "123",
        "status": "PENDING",
        "grandTotal": 29.99
      }
    },
    "status_codes": [0, 13],
    "latencies": ["100ms", "300ms"],
    "error_message": "mock internal error"
  }
]
```

**gRPC Status Codes**: Use numeric codes (0=OK, 13=INTERNAL, 14=UNAVAILABLE, etc.)

## Latency Simulation

Mockchaos supports random latency simulation to mimic real-world network conditions:

### How It Works

1. **Multiple Latencies**: Provide an array of durations
2. **Random Selection**: Each request randomly selects one latency from the array
3. **Applied Before Response**: The latency is applied before sending the response

### Examples

```go
// Single latency (always the same)
Latencies: []time.Duration{100 * time.Millisecond}

// Multiple latencies (randomly selected)
Latencies: []time.Duration{
    50 * time.Millisecond,
    100 * time.Millisecond,
    200 * time.Millisecond,
    500 * time.Millisecond,
}

// In JSON format
"latencies": ["50ms", "100ms", "200ms", "500ms"]
```

### Use Cases

- **Chaos Engineering**: Test how your service handles variable latencies
- **Timeout Testing**: Verify timeout handling with high latencies
- **Performance Testing**: Simulate slow dependencies
- **Realistic Testing**: Mimic real-world network conditions

## Standalone Server

Mockchaos includes a CLI tool for running standalone mock servers:

### Installation

```bash
go install github.com/nawafswe/mockchaos/cmd/chaosmock@latest
```

### HTTP Server

```bash
chaosmock \
  -mock_server=http-svc \
  -http_port=8080 \
  -mocks_path=./http-mocks
```

### gRPC Server

```bash
chaosmock \
  -mock_server=grpc-svc \
  -grpc_port=50051 \
  -grpc_proto_dir=./protos \
  -mocks_path=./grpc-mocks/orders
```

### Flags

- `-mock_server`: Service type (`http-svc` or `grpc-svc`)
- `-http_port`: HTTP server port (default: 8080)
- `-grpc_port`: gRPC server port (default: 50051)
- `-grpc_proto_dir`: Directory containing `.proto` files (required for gRPC)
- `-mocks_path`: Directory containing JSON mock files (required)

The `mocks_path` directory is recursively scanned for all `.json` files, which are automatically loaded and combined.

## Kubernetes Deployment

See `examples/k8s/app.yaml` for a complete Kubernetes deployment example.

### Example Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: http-mockserver
spec:
  template:
    spec:
      containers:
        - name: http
          image: your-image
          command: ["/bin/sh", "-c"]
          args:
            - |
              ./bin/chaosmock \
                -mock_server=http-svc \
                -http_port=8080 \
                -mocks_path=./http-mocks
```

## Package Structure

- **`core/http`**: Core HTTP mocking functionality
- **`core/grpc`**: Core gRPC mocking functionality
- **`httptest`**: Test utilities for HTTP mocking
- **`grpctest`**: Test utilities for gRPC mocking
- **`http`**: Standalone HTTP server implementation
- **`grpc`**: Standalone gRPC server implementation
- **`cmd/chaosmock`**: CLI tool for running standalone servers

## Examples

See the `examples/` directory for complete working examples:

- **`examples/httptest/`**: HTTP integration test example with latency simulation and JSON configuration
- **`examples/grpctest/`**: gRPC integration test example with latency simulation
- **`examples/k8s/`**: Kubernetes deployment configuration

### Running the Examples

**HTTP Example:**
```bash
cd examples/httptest
go test -v
```

**gRPC Example:**
```bash
cd examples/grpctest
go generate  # Generate protobuf files
go test -v
```

## Behavior

### HTTP Mocking

- **Matching**: Exact match on HTTP method and path
- **Status Codes**: Randomly selected from the `statuses` array
- **Latencies**: Randomly selected from the `latencies` array (if provided)
- **No Match**: Returns `404 Not Found` with `{"error": "Not found"}`

### gRPC Mocking

- **Service Registration**: Automatically registers services from `.proto` files
- **Status Codes**: Randomly selected from the `status_codes` array
- **Latencies**: Randomly selected from the `latencies` array (if provided)
- **No Match**: Returns `Unimplemented` status error

## Contributing
Contributions are welcome! Please see our [Contributing Guide](CONTRIBUTING.md) for details on how to contribute.

## License

See [LICENSE](LICENSE) file for details.
