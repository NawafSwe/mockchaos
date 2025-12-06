package svc

import (
	"context"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"

	"github.com/nawafswe/mockchaos/core/http"
	httpsvc "github.com/nawafswe/mockchaos/http"
)

// RunHTTPMock starts a new HTTP server with the given handlers and port.
func RunHTTPMock(_ context.Context, port int, mocksPath string) error {
	if mocksPath == "" {
		return fmt.Errorf("mocks_path is required")
	}
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	handlers, err := loadHandlersFromDirectory(mocksPath)
	if err != nil {
		return fmt.Errorf("failed to load mocks directory: %w", err)
	}
	svc := httpsvc.NewServer(handlers...)
	if svc == nil {
		return fmt.Errorf("failed to create http server")
	}
	log.Println("handlers registered")
	log.Printf("server started on port %d\n", port)
	if err := svc.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}
	return nil
}

// loadHandlersFromDirectory recursively loads all JSON handler files from a directory
// and combines them into a single slice of handlers.
func loadHandlersFromDirectory(dir string) ([]http.Handler, error) {
	var allHandlers []http.Handler

	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip directories, non-JSON files, and hidden files and system files
		if info.IsDir() || strings.HasPrefix(info.Name(), ".") || !strings.HasSuffix(strings.ToLower(path), ".json") {
			return nil
		}

		// Read and parse the JSON file
		content, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("failed to read file %s: %w", path, err)
		}

		// Parse handlers from this file
		handlers, err := http.ParseHandlers(content)
		if err != nil {
			return fmt.Errorf("failed to parse handlers from %s: %w", path, err)
		}
		log.Printf("loaded %d handlers from %s\n", len(handlers), path)
		allHandlers = append(allHandlers, handlers...)

		return nil
	})
	return allHandlers, err
}
