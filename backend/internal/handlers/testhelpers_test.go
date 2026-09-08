package handlers

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

// mustDecode unmarshals a recorded JSON response body, failing the test on error.
func mustDecode(t *testing.T, w *httptest.ResponseRecorder, v any) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), v); err != nil {
		t.Fatalf("failed to decode response body %q: %v", w.Body.String(), err)
	}
}
