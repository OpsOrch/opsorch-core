package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"
)

// AuditLogEntry captures structured details for audit actions.
// Action is a dot-separated string, e.g. "incident.created", "incident.query".
type AuditLogEntry struct {
	RequestID string    `json:"request_id"`
	ActorType string    `json:"actor_type"`
	ActorID   string    `json:"actor_id"`
	Timestamp time.Time `json:"timestamp"`
	Action    string    `json:"action"`
}

func logAudit(r *http.Request, action string) {
	entry := AuditLogEntry{
		RequestID: requestIDFromRequest(r),
		ActorType: actorTypeFromRequest(r),
		ActorID:   actorIDFromRequest(r),
		Timestamp: time.Now().UTC(),
		Action:    action,
	}

	encoded, err := json.Marshal(entry)
	if err != nil {
		log.Printf("audit_log_error action=%s request_id=%s err=%v", action, entry.RequestID, err)
		return
	}

	log.Printf("audit_log %s", string(encoded))
}

func actorTypeFromRequest(r *http.Request) string {
	typ := strings.ToLower(strings.TrimSpace(r.Header.Get("X-Actor-Type")))
	switch typ {
	case "copilot":
		return "copilot"
	case "user":
		return "user"
	default:
		return "user"
	}
}

func actorIDFromRequest(r *http.Request) string {
	for _, header := range []string{"X-User-Id", "X-User-ID", "X-Actor-ID", "X-OpsOrch-Actor-ID"} {
		if id := strings.TrimSpace(r.Header.Get(header)); id != "" {
			return id
		}
	}
	return "unknown"
}

func requestIDFromRequest(r *http.Request) string {
	for _, header := range []string{"X-Request-ID", "X-Amzn-Trace-Id", "X-Correlation-ID", "X-Trace-ID"} {
		if id := strings.TrimSpace(r.Header.Get(header)); id != "" {
			return id
		}
	}
	return fmt.Sprintf("generated-%d", time.Now().UnixNano())
}
