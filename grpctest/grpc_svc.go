// Package grpctest provides utilities for testing gRPC servers.
package grpctest

import (
	"net"
	"testing"

	grpcretry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Server represents a gRPC test server.
type Server struct {
	srv      *grpc.Server
	Client   *grpc.ClientConn
	listener net.Listener
}

// NewServer creates a new gRPC test server with handlers.
// It automatically starts the server on a random port and will be cleaned up when the test completes.
func NewServer(t *testing.T, registerServices func(*grpc.Server)) *Server {
	t.Helper()

	listener, err := net.Listen("tcp", "localhost:0") // Bind to any available port
	if err != nil {
		t.Fatalf("failed to listen: %v", err)
	}
	srv := grpc.NewServer()
	registerServices(srv) // Register all required services

	// Start the server in a goroutine
	go func() {
		if err := srv.Serve(listener); err != nil {
			t.Errorf("gRPC server failed to serve: %v", err)
		}
	}()

	clientConn, err := grpc.NewClient(
		listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(grpcretry.UnaryClientInterceptor()),
	)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	server := &Server{
		srv:      srv,
		Client:   clientConn,
		listener: listener,
	}

	t.Cleanup(func() {
		server.Close()
	})

	return server
}

// Addr returns the server address.
func (s *Server) Addr() string {
	return s.listener.Addr().String()
}

// Close shuts down the test server and closes the connection.
func (s *Server) Close() {
	if s.Client != nil {
		_ = s.Client.Close()
	}
	if s.srv != nil {
		s.srv.Stop()
	}
	if s.listener != nil {
		_ = s.listener.Close()
	}
}
