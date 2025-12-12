package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/opsorch/opsorch-core/deployment"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// DeploymentHandler wraps provider wiring for deployments.
type DeploymentHandler struct {
	provider deployment.Provider
}

func newDeploymentHandlerFromEnv(sec SecretProvider) (DeploymentHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "deployment", "OPSORCH_DEPLOYMENT_PROVIDER", "OPSORCH_DEPLOYMENT_CONFIG", "OPSORCH_DEPLOYMENT_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return DeploymentHandler{}, err
	}
	if pluginPath != "" {
		return DeploymentHandler{provider: newDeploymentPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := deployment.LookupProvider(name)
	if !ok {
		return DeploymentHandler{}, fmt.Errorf("deployment provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return DeploymentHandler{}, err
	}
	return DeploymentHandler{provider: provider}, nil
}

// handleDeployment handles deployment HTTP requests from the server
func (s *Server) handleDeployment(w http.ResponseWriter, r *http.Request) bool {
	return s.deployment.handleDeploymentRequest(w, r)
}

// handleDeploymentRequest handles deployment HTTP requests
func (h *DeploymentHandler) handleDeploymentRequest(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/deployments") {
		return false
	}
	if h.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "deployment_provider_missing", Message: "deployment provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
		var query schema.DeploymentQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		deployments, err := h.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "deployment.query")
		writeJSON(w, http.StatusOK, deployments)
		return true
	case len(segments) == 2 && r.Method == http.MethodGet:
		id := segments[1]
		deployment, err := h.provider.Get(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "deployment.get")
		writeJSON(w, http.StatusOK, deployment)
		return true
	default:
		return false
	}
}
