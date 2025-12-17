package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
	"github.com/opsorch/opsorch-core/team"
)

// TeamHandler wraps provider wiring for teams.
type TeamHandler struct {
	provider team.Provider
}

func newTeamHandlerFromEnv(sec SecretProvider) (TeamHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "team", "OPSORCH_TEAM_PROVIDER", "OPSORCH_TEAM_CONFIG", "OPSORCH_TEAM_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return TeamHandler{}, err
	}
	if pluginPath != "" {
		return TeamHandler{provider: newTeamPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := team.LookupProvider(name)
	if !ok {
		return TeamHandler{}, fmt.Errorf("team provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return TeamHandler{}, err
	}
	return TeamHandler{provider: provider}, nil
}

func (s *Server) handleTeam(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/teams") {
		return false
	}
	if s.team.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "team_provider_missing", Message: "team provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
		var query schema.TeamQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		teams, err := s.team.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "team.query")
		writeJSON(w, http.StatusOK, teams)
		return true
	case len(segments) == 2 && r.Method == http.MethodGet:
		id := segments[1]
		team, err := s.team.provider.Get(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "team.get")
		writeJSON(w, http.StatusOK, team)
		return true
	case len(segments) == 3 && segments[2] == "members" && r.Method == http.MethodGet:
		teamID := segments[1]
		members, err := s.team.provider.Members(r.Context(), teamID)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logAudit(r, "team.members")
		writeJSON(w, http.StatusOK, members)
		return true
	default:
		return false
	}
}