package http

import (
	_ "embed"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

//go:embed testdata/valid_handlers.json
var validHandlers []byte

//go:embed testdata/handlers_with_invalid_method.json
var invalidHandlers []byte

//go:embed testdata/invalid_handlers_file.json
var invalidJSONHandlers []byte

//go:embed testdata/handlers_with_invalid_duration.json
var invalidDurationHandlers []byte

func TestParseHandlers(t *testing.T) {
	tests := map[string]struct {
		handlers         []byte
		expectedHandlers []Handler
		expectedErr      error
	}{
		"should successfully parse valid handlers": {
			handlers: validHandlers,
			expectedHandlers: []Handler{
				{
					Path:      "/api/v1/hello",
					Method:    http.MethodGet,
					Response:  []byte(`{"hello":"world"}`),
					Statuses:  []int{http.StatusOK, http.StatusBadRequest, http.StatusInternalServerError},
					Headers:   map[string]string{"Content-Type": "application/json"},
					Latencies: []time.Duration{100 * time.Millisecond, time.Second, 300 * time.Millisecond},
				},
			},
			expectedErr: nil,
		},
		"should fail to parse invalid handlers when invalid http method is present": {
			handlers:    invalidHandlers,
			expectedErr: fmt.Errorf("invalid http method: ? for path /api/v1/hello"),
		},
		"should fail to parse invalid handlers when invalid latency duration is present": {
			handlers:    invalidDurationHandlers,
			expectedErr: fmt.Errorf("failed to parse latency: time: unknown unit \"MNMNMNMN\" in duration \"1MNMNMNMN\""),
		},
		"should fail to parse invalid handlers when invalid file format": {
			handlers:    invalidJSONHandlers,
			expectedErr: fmt.Errorf("failed to unmarshal handlers: json: cannot unmarshal object into Go value of type []http.handler"),
		},
	}

	for name, tc := range tests {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			handlers, err := ParseHandlers(tc.handlers)
			if tc.expectedErr != nil {
				assert.Error(t, err)
				assert.EqualError(t, err, tc.expectedErr.Error())
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedHandlers, handlers)
			}
		})
	}
}

func TestToHandlers(t *testing.T) {
	t.Run("should fail to convert handlers due to invalid http body", func(t *testing.T) {
		_, err := toHandlers([]handler{
			{
				Method:   http.MethodGet,
				Response: map[string]any{"h": func() {}},
			},
		})
		assert.EqualError(t, fmt.Errorf("failed to marshal body: json: unsupported type: func()"), err.Error())
	})
}
