package grpc_test

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/nawafswe/mockchaos/core/grpc"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"

	"os"
	"path/filepath"
	"testing"
)

//go:embed testdata/restaurant.proto
var restaurantProto []byte

//go:embed testdata/responses/get_menu_items.json
var getMenuItemsResponseWithLatency []byte

//go:embed testdata/responses/get_menu_items_with_no_latencies.json
var getMenuItemsResponseWithNoLatency []byte

//go:embed testdata/responses/get_menu_items_with_invalid_latencies.json
var getMenuItemsResponseWithInvalidLatency []byte

//go:embed testdata/responses/get_menu_list_v2.json
var getMenuItemsV2Response []byte

func TestParseHandlers(t *testing.T) {
	tmpDir := t.TempDir()
	protoFile := filepath.Join(tmpDir, "restaurant.proto")
	err := os.WriteFile(protoFile, restaurantProto, 0o644)
	assert.NoError(t, err)
	msgTypes, err := grpc.LoadMessageTypesFromProtoFilePaths(context.TODO(), tmpDir)
	assert.NoError(t, err)

	tests := map[string]struct {
		response    []byte
		handlers    []grpc.Handler
		expectedErr error
	}{
		"should parse handlers from JSON with latencies": {
			response: getMenuItemsResponseWithLatency,
			handlers: []grpc.Handler{
				{
					Service:      "restaurants.RestaurantService",
					Method:       "ListMenuItems",
					StatusCodes:  []codes.Code{codes.OK, codes.Internal},
					Latencies:    []time.Duration{50 * time.Millisecond, 150 * time.Millisecond},
					ErrorMessage: "mock error",
				},
			},
		},
		"should parse handlers from JSON with no latencies": {
			response: getMenuItemsResponseWithNoLatency,
			handlers: []grpc.Handler{
				{
					Service:      "restaurants.RestaurantService",
					Method:       "ListMenuItems",
					StatusCodes:  []codes.Code{codes.OK},
					ErrorMessage: "mock error",
				},
			},
		},
		"should fail to parse handlers from JSON when invalid json provided": {
			response:    []byte(`{}`),
			handlers:    []grpc.Handler{},
			expectedErr: fmt.Errorf("failed to unmarshal handlers: json: cannot unmarshal object into Go value of type []grpc.handler"),
		},
		"should fail to parse handlers from JSON when invalid latencies format provided": {
			response:    getMenuItemsResponseWithInvalidLatency,
			handlers:    []grpc.Handler{},
			expectedErr: fmt.Errorf("failed to parse latency \"150mfffs\": time: unknown unit \"mfffs\" in duration \"150mfffs\""),
		},
		"should fail to parse handlers from JSON when method not found": {
			response:    getMenuItemsV2Response,
			handlers:    []grpc.Handler{},
			expectedErr: fmt.Errorf("response message type restaurants.MenuItemListV2 not found for restaurants.RestaurantService.ListMenuItems"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			handlers, err := grpc.ParseHandlers(tc.response, msgTypes)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.Nil(t, handlers)
				assert.ErrorContains(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assertHandlers(t, tc.handlers, handlers)
			}
		})
	}
}

func assertHandlers(t *testing.T, expectedHandlers, handlers []grpc.Handler) {
	assert.Len(t, handlers, len(expectedHandlers))
	for i, h := range handlers {
		assert.Equal(t, expectedHandlers[i].Service, h.Service)
		assert.Equal(t, expectedHandlers[i].Method, h.Method)
		assert.Equal(t, expectedHandlers[i].StatusCodes, h.StatusCodes)
		assert.Equal(t, expectedHandlers[i].Latencies, h.Latencies)
		assert.Equal(t, expectedHandlers[i].ErrorMessage, h.ErrorMessage)

	}
}
