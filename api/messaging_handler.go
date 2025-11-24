package api

import (
	"fmt"
	"net/http"

	"github.com/opsorch/opsorch-core/messaging"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/schema"
)

// MessagingHandler wraps provider wiring for messaging.
type MessagingHandler struct {
	provider messaging.Provider
}

func newMessagingHandlerFromEnv(sec SecretProvider) (MessagingHandler, error) {
	name, cfg, pluginPath, err := loadProviderConfig(sec, "messaging", "OPSORCH_MESSAGING_PROVIDER", "OPSORCH_MESSAGING_CONFIG", "OPSORCH_MESSAGING_PLUGIN")
	if err != nil || (name == "" && pluginPath == "") {
		return MessagingHandler{}, err
	}
	if pluginPath != "" {
		return MessagingHandler{provider: newMessagingPluginProvider(pluginPath, cfg)}, nil
	}
	constructor, ok := messaging.LookupProvider(name)
	if !ok {
		return MessagingHandler{}, fmt.Errorf("messaging provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return MessagingHandler{}, err
	}
	return MessagingHandler{provider: provider}, nil
}

func (s *Server) handleMessaging(w http.ResponseWriter, r *http.Request) bool {
	if r.URL.Path != "/messages/send" {
		return false
	}
	if s.messaging.provider == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "messaging_provider_missing", Message: "messaging provider not configured"})
		return true
	}
	if r.Method != http.MethodPost {
		return false
	}
	var msg schema.Message
	if err := decodeJSON(r, &msg); err != nil {
		writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
		return true
	}
	res, err := s.messaging.provider.Send(r.Context(), msg)
	if err != nil {
		writeProviderError(w, err)
		return true
	}
	logAudit(r, "message.sent")
	writeJSON(w, http.StatusOK, res)
	return true
}
