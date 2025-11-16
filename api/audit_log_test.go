package api

import (
	"net/http/httptest"
	"testing"
)

func TestActorIDFromRequestPrefersUserIDHeader(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("X-User-Id", " user-123 ")
	req.Header.Set("X-Actor-ID", "actor-legacy")

	if got := actorIDFromRequest(req); got != "user-123" {
		t.Fatalf("expected actor id from X-User-Id, got %q", got)
	}
}

func TestActorIDFromRequestUnknownWhenMissing(t *testing.T) {
	req := httptest.NewRequest("GET", "/", nil)

	if got := actorIDFromRequest(req); got != "unknown" {
		t.Fatalf("expected fallback to unknown, got %q", got)
	}
}
