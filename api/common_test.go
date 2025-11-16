package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/opsorch/opsorch-core/orcherr"
)

func decodeBody(t *testing.T, rr *httptest.ResponseRecorder) map[string]string {
	t.Helper()
	var body map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode body: %v", err)
	}
	return body
}

func TestWriteProviderErrorValue(t *testing.T) {
	rr := httptest.NewRecorder()
	writeProviderError(rr, orcherr.New("bad_request", "bad input", nil))

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected status %d, got %d", http.StatusBadRequest, rr.Code)
	}
	body := decodeBody(t, rr)
	if body["code"] != "bad_request" || body["message"] != "bad input" {
		t.Fatalf("unexpected body: %+v", body)
	}
}

func TestWriteProviderErrorPointer(t *testing.T) {
	rr := httptest.NewRecorder()
	writeProviderError(rr, &orcherr.OpsOrchError{Code: "not_found", Message: "missing"})

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected status %d, got %d", http.StatusNotFound, rr.Code)
	}
	body := decodeBody(t, rr)
	if body["code"] != "not_found" || body["message"] != "missing" {
		t.Fatalf("unexpected body: %+v", body)
	}
}
