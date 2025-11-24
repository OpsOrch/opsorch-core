package api

import (
	"fmt"
	"net/http"

	"github.com/opsorch/opsorch-core/log"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// LogHandler wraps provider wiring for logs.
type LogHandler struct {
	provider log.Provider
}

func newLogHandlerFromEnv(sec SecretProvider) (LogHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "log", "OPSORCH_LOG_PROVIDER", "OPSORCH_LOG_CONFIG", "OPSORCH_LOG_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return LogHandler{}, err
	}
	if pluginPath != "" {
		return LogHandler{provider: newLogPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := log.LookupProvider(name)
	if !ok {
		return LogHandler{}, fmt.Errorf("log provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return LogHandler{}, err
	}
	return LogHandler{provider: provider}, nil
}

func (s *Server) handleLog(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/logs/query" {
		return false
	}
	if s.log.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "log_provider_missing", Message: "log provider not configured"})
		return true
	}
	if r.Method != http.MethodPost {
		return false
	}
	var query schema.LogQuery
	if err := decodeJSON(r, &query); err != nil {
		writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
		return true
	}
	results, err := s.log.provider.Query(r.Context(), query)
	if err != nil {
		writeProviderError(w, err)
		return true
	}
	logAudit(r, "log.query")
	writeJSON(w, http.StatusOK, results)
	return true
}
