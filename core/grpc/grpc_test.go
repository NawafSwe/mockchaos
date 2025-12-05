package grpc

import "testing"

func TestHandlerKey(t *testing.T) {
	t.Run("should return correct key", func(t *testing.T) {
		key := HandlerKey("users", "get")
		expectedKey := "users.get"
		if key != expectedKey {
			t.Errorf("expected key %s, got %s", expectedKey, key)
		}
	})
}
