package grpc_test

import (
	"context"
	_ "embed"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/nawafswe/mockchaos/core/grpc"
	"github.com/stretchr/testify/assert"
)

//go:embed testdata/one_level_proto.proto
var oneLevelProto []byte

//go:embed testdata/order.proto
var multiLevelProto []byte

//go:embed testdata/invalid_proto_version.proto
var invalidProtoVersion []byte

func TestLoadMessageTypesFromDir(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(dir string)
		loadedMessages []string
		expectedErr    error
	}{
		{
			name: "valid one level proto file",
			setup: func(dir string) {
				protoFile := filepath.Join(dir, "test.proto")
				_ = os.WriteFile(protoFile, oneLevelProto, 0o644)
			},
			loadedMessages: []string{"test.TestMessage"},
		},
		{
			name: "valid multi level proto file",
			setup: func(dir string) {
				protoFile := filepath.Join(dir, "test2.proto")
				_ = os.WriteFile(protoFile, multiLevelProto, 0o644)
			},
			loadedMessages: []string{"orders.GetOrderRequest", "orders.GetOrderResponse", "orders.Order", "orders.Order.OrderedItem"},
		},
		{
			name: "invalid proto file",
			setup: func(dir string) {
				protoFile := filepath.Join(dir, "test.proto")
				_ = os.WriteFile(protoFile, invalidProtoVersion, 0o644)
			},
			expectedErr: fmt.Errorf("syntax value must be \"proto2\" or \"proto3\""),
		},
		{
			name:        "no proto files",
			setup:       func(dir string) {},
			expectedErr: fmt.Errorf("no .proto files found"),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			tmpDir := t.TempDir()
			tc.setup(tmpDir)
			msgTypes, err := grpc.LoadMessageTypesFromProtoFilePaths(context.TODO(), tmpDir)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, msgTypes)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, msgTypes)
				keys := make([]string, 0, len(msgTypes))
				for k := range msgTypes {
					keys = append(keys, k)
				}
				// sort before comparing.
				slices.Sort(keys)
				assert.Equal(t, tc.loadedMessages, keys)
			}
		})
	}
}
