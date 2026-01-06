package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/orchestration"
	"github.com/opsorch/opsorch-core/schema"
)

// OrchestrationHandler wraps provider wiring for orchestration.
type OrchestrationHandler struct {
	provider orchestration.Provider
}

func newOrchestrationHandlerFromEnv(sec SecretProvider) (OrchestrationHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "orchestration", "OPSORCH_ORCHESTRATION_PROVIDER", "OPSORCH_ORCHESTRATION_CONFIG", "OPSORCH_ORCHESTRATION_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return OrchestrationHandler{}, err
	}
	if pluginPath != "" {
		return OrchestrationHandler{provider: newOrchestrationPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := orchestration.LookupProvider(name)
	if !ok {
		return OrchestrationHandler{}, fmt.Errorf("orchestration provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return OrchestrationHandler{}, err
	}
	return OrchestrationHandler{provider: provider}, nil
}

func (s *Server) handleOrchestration(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/orchestration") {
		return false
	}
	if s.orchestration.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "orchestration_provider_missing", Message: "orchestration provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	// POST /orchestration/plans/query
	case len(segments) == 3 && segments[1] == "plans" && segments[2] == "query" && r.Method == http.MethodPost:
		var query schema.OrchestrationPlanQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		plans, err := s.orchestration.provider.QueryPlans(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "orchestration.plans.query")
		writeJSON(w, http.StatusOK, plans)
		return true

	// GET /orchestration/plans/{planId}
	case len(segments) == 3 && segments[1] == "plans" && r.Method == http.MethodGet:
		planID := segments[2]
		plan, err := s.orchestration.provider.GetPlan(r.Context(), planID)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "orchestration.plans.get")
		writeJSON(w, http.StatusOK, plan)
		return true

	// POST /orchestration/runs/query
	case len(segments) == 3 && segments[1] == "runs" && segments[2] == "query" && r.Method == http.MethodPost:
		var query schema.OrchestrationRunQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		runs, err := s.orchestration.provider.QueryRuns(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "orchestration.runs.query")
		writeJSON(w, http.StatusOK, runs)
		return true

	// POST /orchestration/runs
	case len(segments) == 2 && segments[1] == "runs" && r.Method == http.MethodPost:
		var input struct {
			PlanID string `json:"planId"`
		}
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		if input.PlanID == "" {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: "planId is required"})
			return true
		}
		run, err := s.orchestration.provider.StartRun(r.Context(), input.PlanID)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "orchestration.runs.start")
		writeJSON(w, http.StatusCreated, run)
		return true

	// GET /orchestration/runs/{runId}
	case len(segments) == 3 && segments[1] == "runs" && r.Method == http.MethodGet:
		runID := segments[2]
		run, err := s.orchestration.provider.GetRun(r.Context(), runID)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "orchestration.runs.get")
		writeJSON(w, http.StatusOK, run)
		return true

	// POST /orchestration/runs/{runId}/steps/{stepId}/complete
	case len(segments) == 6 && segments[1] == "runs" && segments[3] == "steps" && segments[5] == "complete" && r.Method == http.MethodPost:
		runID := segments[2]
		stepID := segments[4]
		var input struct {
			Actor string `json:"actor"`
			Note  string `json:"note"`
		}
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		if err := s.orchestration.provider.CompleteStep(r.Context(), runID, stepID, input.Actor, input.Note); err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "orchestration.runs.steps.complete")
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
		return true

	default:
		return false
	}
}
