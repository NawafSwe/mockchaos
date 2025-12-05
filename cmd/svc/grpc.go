package svc

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	coregrpc "github.com/nawafswe/mockchaos/core/grpc"
	chaosgrpc "github.com/nawafswe/mockchaos/grpc" // Add this - your grpc package
	"google.golang.org/protobuf/reflect/protoreflect"
)

// RunGRPCMock starts a new gRPC mock server.
// protoDir: directory containing .proto files (can have subdirs).
// mocksPath: directory containing JSON mocks.
func RunGRPCMock(ctx context.Context, port int, protoDir, mocksPath string) {
	if protoDir == "" {
		log.Fatal("grpc: proto_dir is required")
	}
	if mocksPath == "" {
		log.Fatal("grpc: mocks_path is required")
	}

	// Load message types from .proto files
	msgTypes, err := coregrpc.LoadMessageTypesFromProtoDir(ctx, protoDir)
	if err != nil {
		log.Fatalf("grpc: failed to load proto descriptors: %v", err)
	}

	// Load handlers from mocks-directory
	handlers, err := loadGRPCHandlersFromDirectory(mocksPath, msgTypes)
	if err != nil {
		log.Fatalf("grpc: failed to load mocks directory: %v", err)
	}
	if len(handlers) == 0 {
		log.Fatalf("grpc: no handlers found in %s", mocksPath)
	}

	// Create a gRPC server with handlers.
	srv := chaosgrpc.NewServer(handlers...)

	// Register services from proto files
	if err := srv.RegisterServices(); err != nil {
		log.Fatalf("grpc: failed to register services: %v", err)
	}

	// Listen and serve
	if err := srv.Listen(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("grpc: failed to listen: %v", err)
	}

	log.Printf("grpc: loaded %d handlers", len(handlers))
	log.Printf("grpc: server started on port %d\n", port)

	if err := srv.Serve(); err != nil {
		log.Fatalf("grpc: failed to serve: %v", err)
	}
}

// loadGRPCHandlersFromDirectory recursively loads all JSON handler files from a directory
// and combines them into a single slice of handlers.
func loadGRPCHandlersFromDirectory(mockPath string, msgTypes map[string]protoreflect.MessageType) ([]coregrpc.Handler, error) {
	var allHandlers []coregrpc.Handler

	err := filepath.Walk(mockPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		// Skip directories and non-JSON files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") || !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		handlers, err := coregrpc.ParseHandlers(content, msgTypes)
		if err != nil {
			return fmt.Errorf("failed to parse handlers from %s: %w", path, err)
		}

		allHandlers = append(allHandlers, handlers...)
		return nil
	})

	return allHandlers, err
}
