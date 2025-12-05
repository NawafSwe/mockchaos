package grpc

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/bufbuild/protocompile"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"
	"google.golang.org/protobuf/types/dynamicpb"
)

// LoadMessageTypesFromProtoFilePaths compiles proto files under a protoDir
// and returns a map of the full message name mapped to a MessageType.
func LoadMessageTypesFromProtoFilePaths(ctx context.Context, protoDir string) (map[string]protoreflect.MessageType, error) {
	var protoFiles []string
	// Collect all .proto files under protoDir.
	err := filepath.Walk(protoDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if filepath.Ext(path) == ".proto" {
			protoFiles = append(protoFiles, path)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("failed to walk proto dir: %w", err)
	}
	if len(protoFiles) == 0 {
		return nil, fmt.Errorf("no .proto files found in %s", protoDir)
	}

	// Setup compiler with protoDir as an import path.
	resolver := &protocompile.SourceResolver{}
	compiler := protocompile.Compiler{
		Resolver: resolver,
	}

	// Compile all proto files.
	fileDescriptors, err := compiler.Compile(ctx, protoFiles...)
	if err != nil {
		return nil, fmt.Errorf("failed to compile proto files: %w", err)
	}

	// Build a message type map from descriptors.
	msgTypes := make(map[string]protoreflect.MessageType)

	for _, fd := range fileDescriptors {
		// fd is already a protoreflect.FileDescriptor
		// Register file globally.
		_ = protoregistry.GlobalFiles.RegisterFile(fd) // ignore "already registered" errors
		collectMessages(fd.Messages(), msgTypes)
	}

	return msgTypes, nil
}

// collectMessages collects all message descriptors and their corresponding message types into the provided map.
// It recursively processes nested message descriptors.
// Params: msgs - a collection of message descriptors to process.
//
//	out - a map to store the full name of the message as the key and its type as the value.
func collectMessages(msgs protoreflect.MessageDescriptors, out map[string]protoreflect.MessageType) {
	for i := 0; i < msgs.Len(); i++ {
		md := msgs.Get(i)
		mt := dynamicpb.NewMessageType(md)
		out[string(md.FullName())] = mt

		// Recurse into nested messages
		if nested := md.Messages(); nested.Len() > 0 {
			collectMessages(nested, out)
		}
	}
}
