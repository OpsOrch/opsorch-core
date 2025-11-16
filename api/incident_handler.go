package api

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/opsorch/opsorch-core/incident"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// IncidentHandler wraps provider wiring for incidents.
type IncidentHandler struct {
	provider incident.Provider
}

func newIncidentHandlerFromEnv(sec SecretProvider) (IncidentHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "incident", "OPSORCH_INCIDENT_PROVIDER", "OPSORCH_INCIDENT_CONFIG", "OPSORCH_INCIDENT_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return IncidentHandler{}, err
	}
	if pluginPath != "" {
		return IncidentHandler{provider: newIncidentPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := incident.LookupProvider(name)
	if !ok {
		return IncidentHandler{}, fmt.Errorf("incident provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return IncidentHandler{}, err
	}
	return IncidentHandler{provider: provider}, nil
}

func (s *Server) handleIncident(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/incidents") {
		return false
	}
	if s.incident.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "incident_provider_missing", Message: "incident provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
		var query schema.IncidentQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		incidents, err := s.incident.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, incidents)
		return true
	case len(segments) == 1 && r.Method == http.MethodGet:
		incidents, err := s.incident.provider.List(r.Context())
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, incidents)
		return true
	case len(segments) == 1 && r.Method == http.MethodPost:
		var input schema.CreateIncidentInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		inc, err := s.incident.provider.Create(r.Context(), input)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logMutatingAction(r, "create_incident", "incident", inc.ID)
		writeJSON(w, http.StatusCreated, inc)
		return true
	case len(segments) == 2 && r.Method == http.MethodGet:
		id := segments[1]
		inc, err := s.incident.provider.Get(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, inc)
		return true
	case len(segments) == 2 && r.Method == http.MethodPatch:
		id := segments[1]
		var input schema.UpdateIncidentInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		inc, err := s.incident.provider.Update(r.Context(), id, input)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logMutatingAction(r, "update_incident", "incident", id)
		writeJSON(w, http.StatusOK, inc)
		return true
	case len(segments) == 3 && segments[2] == "timeline" && r.Method == http.MethodGet:
		id := segments[1]
		timeline, err := s.incident.provider.GetTimeline(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, timeline)
		return true
	case len(segments) == 3 && segments[2] == "timeline" && r.Method == http.MethodPost:
		id := segments[1]
		var input schema.TimelineAppendInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		if input.At.IsZero() {
			input.At = time.Now()
		}
		if err := s.incident.provider.AppendTimeline(r.Context(), id, input); err != nil {
			writeProviderError(w, err)
			return true
		}
		logMutatingAction(r, "append_incident_timeline", "incident", id)
		writeJSON(w, http.StatusCreated, map[string]string{"status": "ok"})
		return true
	default:
		return false
	}
}
