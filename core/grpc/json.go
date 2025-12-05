package grpc

import (
	"encoding/json"
	"fmt"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/dynamicpb"
)

// handler represents a single handler specification in JSON format.
type handler struct {
	Service           string         `json:"service"`
	Method            string         `json:"method"`
	Response          map[string]any `json:"response"`
	ResponseProtoType string         `json:"response_proto_type"`
	StatusCodes       []int          `json:"status_codes"`
	Latencies         []string       `json:"latencies"`
	ErrorMessage      string         `json:"error_message"`
}

// ParseHandlers parses a JSON byte slice containing gRPC handler configurations.
//
// The JSON format must be an array of handler objects, where each handler has:
//   - service: string (required) - Fully qualified service name (e.g., "orders.OrderService")
//   - method: string (required) - RPC method name (e.g., "GetOrder")
//   - response: object (optional) - Response message as JSON
//   - status_codes: array of integers (required) - gRPC status codes (0=OK, 13=INTERNAL, etc.)
//   - latencies: array of strings (optional) - Duration strings (e.g., "100ms", "1s")
//   - error_message: string (optional) - Error message if status code is not OK
func ParseHandlers(content []byte, messageTypes map[string]protoreflect.MessageType) ([]Handler, error) {
	var handlers []handler
	if err := json.Unmarshal(content, &handlers); err != nil {
		return nil, fmt.Errorf("failed to unmarshal handlers: %w", err)
	}
	return toHandlers(handlers, messageTypes)
}

// toHandlers converts JSON handlers to Handler structs.
func toHandlers(handlers []handler, messageTypes map[string]protoreflect.MessageType) ([]Handler, error) {
	grpcHandlers := make([]Handler, 0, len(handlers))
	for _, hd := range handlers {
		// Parse latencies
		latencies, err := parseLatencies(hd.Latencies)
		if err != nil {
			return nil, err
		}

		// Convert status codes
		statusCodes := parseStatusCodes(hd.StatusCodes)

		var response protoreflect.Message
		if hd.Response != nil {
			mt, ok := messageTypes[hd.ResponseProtoType]
			if !ok {
				return nil, fmt.Errorf("response message type %s not found for %s.%s", hd.ResponseProtoType, hd.Service, hd.Method)
			}

			response = dynamicpb.NewMessage(mt.Descriptor())
			if err := jsonToProtobuf(hd.Response, response); err != nil {
				return nil, fmt.Errorf("failed to convert response JSON to protobuf: %w", err)
			}
		}
		grpcHandlers = append(grpcHandlers, Handler{
			Service:      hd.Service,
			Method:       hd.Method,
			Response:     response,
			ResponseJSON: hd.Response,
			StatusCodes:  statusCodes,
			Latencies:    latencies,
			ErrorMessage: hd.ErrorMessage,
		})
	}
	return grpcHandlers, nil
}

// parseLatencies converts duration strings to time.Duration values.
func parseLatencies(latencies []string) ([]time.Duration, error) {
	if len(latencies) == 0 {
		return nil, nil
	}
	durations := make([]time.Duration, len(latencies))
	for i, l := range latencies {
		d, err := time.ParseDuration(l)
		if err != nil {
			return nil, fmt.Errorf("failed to parse latency %q: %w", l, err)
		}
		durations[i] = d
	}
	return durations, nil
}

// parseStatusCodes converts integer codes to gRPC codes.
func parseStatusCodes(statuses []int) []codes.Code {
	if len(statuses) == 0 {
		return []codes.Code{codes.OK}
	}
	grpcCodes := make([]codes.Code, 0, len(statuses))
	for _, c := range statuses {
		grpcCodes = append(grpcCodes, codes.Code(c))
	}
	return grpcCodes
}

// jsonToProtobuf converts JSON to a protobuf message.
func jsonToProtobuf(jsonData map[string]any, msg protoreflect.Message) error {
	raw, err := json.Marshal(jsonData)
	if err != nil {
		return fmt.Errorf("failed to marshal response JSON: %w", err)
	}

	u := protojson.UnmarshalOptions{
		DiscardUnknown: true, // ignore extra fields in JSON instead of failing
	}
	if err := u.Unmarshal(raw, msg.Interface()); err != nil {
		return fmt.Errorf("failed to unmarshal JSON into protobuf: %w", err)
	}
	return nil
}
