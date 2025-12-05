package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"strings"

	coregrpc "github.com/nawafswe/mockchaos/core/grpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

const (
	reflectionV1PkgName    = "grpc.reflection.v1"
	reflectionAlphaPkgName = "grpc.reflection.v1alpha"
)

// Server represents a gRPC mock server.
type Server struct {
	srv      *grpc.Server
	lis      net.Listener
	handlers map[string]coregrpc.Handler
}

// NewServer creates a new gRPC server with handlers.
func NewServer(handlers ...coregrpc.Handler) *Server {
	// Build handler map
	handlerMap := make(map[string]coregrpc.Handler, len(handlers))
	for _, h := range handlers {
		key := coregrpc.HandlerKey(h.Service, h.Method)
		handlerMap[key] = h
	}
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(coregrpc.Interceptor(handlerMap)),
	)
	return &Server{
		srv:      srv,
		handlers: handlerMap,
	}
}

// RegisterServices registers all services found in the proto files.
func (s *Server) RegisterServices() error {
	// Enable gRPC reflection (this registers grpc.reflection services)
	reflection.Register(s.srv)

	// Register all services from proto files.
	protoregistry.GlobalFiles.RangeFiles(func(fd protoreflect.FileDescriptor) bool {
		// Skip reflection-related files
		pkg := string(fd.Package())
		if pkg == reflectionV1PkgName || pkg == reflectionAlphaPkgName {
			return true // Skip, already registered by reflection.Register
		}

		services := fd.Services()
		for i := 0; i < services.Len(); i++ {
			svc := services.Get(i)
			serviceName := string(svc.FullName())

			// Skip reflection services
			if strings.HasPrefix(serviceName, reflectionV1PkgName) {
				continue
			}
			if err := s.registerService(svc); err != nil {
				log.Printf("Warning: failed to register service %s: %v", serviceName, err)
			}
		}
		return true
	})

	return nil
}

// registerService registers a service using dynamic service info.
func (s *Server) registerService(svcDesc protoreflect.ServiceDescriptor) error {
	serviceName := string(svcDesc.FullName())
	log.Printf("Registering service: %s", serviceName)

	// Build method descriptors
	methods := make([]grpc.MethodDesc, 0, svcDesc.Methods().Len())

	for i := 0; i < svcDesc.Methods().Len(); i++ {
		method := svcDesc.Methods().Get(i)
		methodName := string(method.Name())
		fullMethod := "/" + serviceName + "/" + methodName

		// Get input and output types
		inputType := method.Input()
		outputType := method.Output()

		// Create method descriptor
		methodDesc := grpc.MethodDesc{
			MethodName: methodName,
			Handler: func(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
				// Decode request
				req := dynamicpb.NewMessage(inputType)
				if err := dec(req); err != nil {
					return nil, err
				}

				// Call interceptor with the actual handler
				return interceptor(ctx, req, &grpc.UnaryServerInfo{
					Server:     srv,
					FullMethod: fullMethod,
				}, func(ctx context.Context, req any) (any, error) {
					// This handler should never be called because the interceptor returns early
					return dynamicpb.NewMessage(outputType), nil
				})
			},
		}
		methods = append(methods, methodDesc)

		key := serviceName + "." + methodName
		if _, ok := s.handlers[key]; ok {
			log.Printf("  - Method: %s (has handler)", methodName)
		} else {
			log.Printf("  - Method: %s (no handler)", methodName)
		}
	}

	// Register service - RegisterService doesn't return a value
	serviceDesc := grpc.ServiceDesc{
		ServiceName: serviceName,
		Methods:     methods,
	}

	s.srv.RegisterService(&serviceDesc, nil)
	return nil
}

// Listen starts listening on the given address.
func (s *Server) Listen(addr string) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}
	s.lis = lis
	return nil
}

// Serve starts serving requests.
func (s *Server) Serve() error {
	if s.lis == nil {
		return fmt.Errorf("server not listening - call Listen first")
	}
	return s.srv.Serve(s.lis)
}

// Stop gracefully stops the server.
func (s *Server) Stop() {
	s.srv.GracefulStop()
}

// Close stops the server and closes the listener.
func (s *Server) Close() error {
	s.Stop()
	if s.lis != nil {
		return s.lis.Close()
	}
	return nil
}
