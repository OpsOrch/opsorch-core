package api

import (
	"fmt"
	"net/http"

	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
	"github.com/opsorch/opsorch-core/service"
)

// ServiceHandler wraps provider wiring for services.
type ServiceHandler struct {
	provider service.Provider
}

func newServiceHandlerFromEnv(sec SecretProvider) (ServiceHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "service", "OPSORCH_SERVICE_PROVIDER", "OPSORCH_SERVICE_CONFIG", "OPSORCH_SERVICE_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return ServiceHandler{}, err
	}
	if pluginPath != "" {
		return ServiceHandler{provider: newServicePluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := service.LookupProvider(name)
	if !ok {
		return ServiceHandler{}, fmt.Errorf("service provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return ServiceHandler{}, err
	}
	return ServiceHandler{provider: provider}, nil
}

func (s *Server) handleService(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/services" && r.URL.Path != "/services/query" {
		return false
	}
	if s.service.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "service_provider_missing", Message: "service provider not configured"})
		return true
	}

	switch {
	case r.URL.Path == "/services" && r.Method == http.MethodGet:
		services, err := s.service.provider.Query(r.Context(), schema.ServiceQuery{})
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, services)
		return true
	case r.URL.Path == "/services/query" && r.Method == http.MethodPost:
		var query schema.ServiceQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		services, err := s.service.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, services)
		return true
	default:
		return false
	}
}
