package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/opsorch/opsorch-core/alert"
	"github.com/opsorch/opsorch-core/deployment"
	"github.com/opsorch/opsorch-core/incident"
	"github.com/opsorch/opsorch-core/log"
	"github.com/opsorch/opsorch-core/messaging"
	"github.com/opsorch/opsorch-core/metric"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/service"
	"github.com/opsorch/opsorch-core/team"
	"github.com/opsorch/opsorch-core/ticket"
)

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/providers/") || r.Method != http.MethodGet {
		return false
	}
	raw := strings.TrimPrefix(r.URL.Path, "/providers/")
	capability, ok := normalizeCapability(raw)
	if !ok {
		writeError(w, http.StatusNotFound, orcherr.OpsOrchError{Code: "not_found", Message: "unknown capability"})
		return true
	}

	var providers []string
	switch capability {
	case "incident":
		providers = incident.Providers()
	case "alert":
		providers = alert.Providers()
	case "log":
		providers = log.Providers()
	case "metric":
		providers = metric.Providers()
	case "ticket":
		providers = ticket.Providers()
	case "messaging":
		providers = messaging.Providers()
	case "service":
		providers = service.Providers()
	case "deployment":
		providers = deployment.Providers()
	case "team":
		providers = team.Providers()
	}
	writeJSON(w, http.StatusOK, map[string]any{"providers": providers})
	return true
}

func decodeConfig(envVar string) (map[string]any, error) {
	raw := strings.TrimSpace(os.Getenv(envVar))
	if raw == "" {
		return map[string]any{}, nil
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("invalid config in %s: %w", envVar, err)
	}
	return cfg, nil
}
