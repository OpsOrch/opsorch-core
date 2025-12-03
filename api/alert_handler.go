package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/opsorch/opsorch-core/alert"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// AlertHandler wraps provider wiring for alerts.
type AlertHandler struct {
	provider alert.Provider
}

func newAlertHandlerFromEnv(sec SecretProvider) (AlertHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "alert", "OPSORCH_ALERT_PROVIDER", "OPSORCH_ALERT_CONFIG", "OPSORCH_ALERT_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return AlertHandler{}, err
	}
	if pluginPath != "" {
		// Plugins not yet implemented for alerts, but structure is here
		return AlertHandler{}, fmt.Errorf("alert plugins not yet supported")
	}
	constructor, ok := alert.LookupProvider(name)
	if !ok {
		return AlertHandler{}, fmt.Errorf("alert provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return AlertHandler{}, err
	}
	return AlertHandler{provider: provider}, nil
}

func (s *Server) handleAlert(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/alerts") {
		return false
	}
	if s.alert.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "alert_provider_missing", Message: "alert provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
		var query schema.AlertQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		alerts, err := s.alert.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "alert.query")
		writeJSON(w, http.StatusOK, alerts)
		return true
	case len(segments) == 2 && r.Method == http.MethodGet:
		id := segments[1]
		al, err := s.alert.provider.Get(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "alert.get")
		writeJSON(w, http.StatusOK, al)
		return true
	default:
		return false
	}
}
