package api

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
	"github.com/opsorch/opsorch-core/ticket"
)

// TicketHandler wraps provider wiring for tickets.
type TicketHandler struct {
	provider ticket.Provider
}

func newTicketHandlerFromEnv(sec SecretProvider) (TicketHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "ticket", "OPSORCH_TICKET_PROVIDER", "OPSORCH_TICKET_CONFIG", "OPSORCH_TICKET_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return TicketHandler{}, err
	}
	if pluginPath != "" {
		return TicketHandler{provider: newTicketPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := ticket.LookupProvider(name)
	if !ok {
		return TicketHandler{}, fmt.Errorf("ticket provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return TicketHandler{}, err
	}
	return TicketHandler{provider: provider}, nil
}

func (s *Server) handleTicket(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/tickets") {
		return false
	}
	if s.ticket.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "ticket_provider_missing", Message: "ticket provider not configured"})
		return true
	}

	path := strings.TrimSuffix(r.URL.Path, "/")
	segments := strings.Split(strings.Trim(path, "/"), "/")

	switch {
	case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
		var query schema.TicketQuery
		if err := decodeJSON(r, &query); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		tickets, err := s.ticket.provider.Query(r.Context(), query)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, tickets)
		return true
	case len(segments) == 1 && r.Method == http.MethodPost:
		var input schema.CreateTicketInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		t, err := s.ticket.provider.Create(r.Context(), input)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logMutatingAction(r, "create_ticket", "ticket", t.ID)
		writeJSON(w, http.StatusCreated, t)
		return true
	case len(segments) == 2 && r.Method == http.MethodGet:
		id := segments[1]
		t, err := s.ticket.provider.Get(r.Context(), id)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		writeJSON(w, http.StatusOK, t)
		return true
	case len(segments) == 2 && r.Method == http.MethodPatch:
		id := segments[1]
		var input schema.UpdateTicketInput
		if err := decodeJSON(r, &input); err != nil {
			writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
			return true
		}
		t, err := s.ticket.provider.Update(r.Context(), id, input)
		if err != nil {
			writeProviderError(w, err)
			return true
		}
		logMutatingAction(r, "update_ticket", "ticket", id)
		writeJSON(w, http.StatusOK, t)
		return true
	default:
		return false
	}
}
