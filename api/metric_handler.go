package api

import (
	"fmt"
	"net/http"

	"github.com/opsorch/opsorch-core/metric"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// MetricHandler wraps provider wiring for metrics.
type MetricHandler struct {
	provider metric.Provider
}

func newMetricHandlerFromEnv(sec SecretProvider) (MetricHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "metric", "OPSORCH_METRIC_PROVIDER", "OPSORCH_METRIC_CONFIG", "OPSORCH_METRIC_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return MetricHandler{}, err
	}
	if pluginPath != "" {
		return MetricHandler{provider: newMetricPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := metric.LookupProvider(name)
	if !ok {
		return MetricHandler{}, fmt.Errorf("metric provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return MetricHandler{}, err
	}
	return MetricHandler{provider: provider}, nil
}

func (s *Server) handleMetric(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/metrics/query" && r.URL.Path != "/metrics/describe" {
		return false
	}
	if s.metric.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "metric_provider_missing", Message: "metric provider not configured"})
		return true
	}

	switch {
	case r.URL.Path == "/metrics/query" && r.Method == http.MethodPost:
		var query schema.MetricQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		results, err := s.metric.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "metric.query")
		writeJSON(w, http.StatusOK, results)
		return true
	case r.URL.Path == "/metrics/describe" && (r.Method == http.MethodGet || r.Method == http.MethodPost):
		var scope schema.QueryScope
		if r.Method == http.MethodPost {
			if err := decodeJSON(r, &scope); err != nil {
				writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
				return true
			}
		} else {
			scope.Service = r.URL.Query().Get("service")
			scope.Environment = r.URL.Query().Get("environment")
			scope.Team = r.URL.Query().Get("team")
		}

		descriptors, err := s.metric.provider.Describe(r.Context(), scope)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "metric.describe")
		writeJSON(w, http.StatusOK, map[string]any{"metrics": descriptors})
		return true
	default:
		return false
	}
}
