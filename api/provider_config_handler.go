package api

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/opsorch/opsorch-core/incident"
	"github.com/opsorch/opsorch-core/log"
	"github.com/opsorch/opsorch-core/messaging"
	"github.com/opsorch/opsorch-core/metric"
	"github.com/opsorch/opsorch-core/orcherr"
	"github.com/opsorch/opsorch-core/service"
	"github.com/opsorch/opsorch-core/ticket"
)

// providerConfigRequest captures the payload to set a provider for a capability.
type providerConfigRequest struct {
	Provider string         `json:"provider"`
	Config   map[string]any `json:"config"`
	Plugin   string         `json:"plugin,omitempty"`
}

func (s *Server) handleIncidentProviderConfig(name, pluginPath string, cfg map[string]any) error {
	if pluginPath != "" {
		s.incident.provider = newIncidentPluginProvider(pluginPath, cfg)
		return nil
	}
	constructor, ok := incident.LookupProvider(name)
	if !ok {
		return fmt.Errorf("incident provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return err
	}
	s.incident.provider = provider
	return nil
}

func (s *Server) handleLogProviderConfig(name, pluginPath string, cfg map[string]any) error {
	if pluginPath != "" {
		s.log.provider = newLogPluginProvider(pluginPath, cfg)
		return nil
	}
	constructor, ok := log.LookupProvider(name)
	if !ok {
		return fmt.Errorf("log provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return err
	}
	s.log.provider = provider
	return nil
}

func (s *Server) handleMetricProviderConfig(name, pluginPath string, cfg map[string]any) error {
	if pluginPath != "" {
		s.metric.provider = newMetricPluginProvider(pluginPath, cfg)
		return nil
	}
	constructor, ok := metric.LookupProvider(name)
	if !ok {
		return fmt.Errorf("metric provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return err
	}
	s.metric.provider = provider
	return nil
}

func (s *Server) handleTicketProviderConfig(name, pluginPath string, cfg map[string]any) error {
	if pluginPath != "" {
		s.ticket.provider = newTicketPluginProvider(pluginPath, cfg)
		return nil
	}
	constructor, ok := ticket.LookupProvider(name)
	if !ok {
		return fmt.Errorf("ticket provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return err
	}
	s.ticket.provider = provider
	return nil
}

func (s *Server) handleMessagingProviderConfig(name, pluginPath string, cfg map[string]any) error {
	if pluginPath != "" {
		s.messaging.provider = newMessagingPluginProvider(pluginPath, cfg)
		return nil
	}
	constructor, ok := messaging.LookupProvider(name)
	if !ok {
		return fmt.Errorf("messaging provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return err
	}
	s.messaging.provider = provider
	return nil
}

func (s *Server) handleServiceProviderConfig(name, pluginPath string, cfg map[string]any) error {
	if pluginPath != "" {
		s.service.provider = newServicePluginProvider(pluginPath, cfg)
		return nil
	}
	constructor, ok := service.LookupProvider(name)
	if !ok {
		return fmt.Errorf("service provider %s not registered", name)
	}
	provider, err := constructor(cfg)
	if err != nil {
		return err
	}
	s.service.provider = provider
	return nil
}

func (s *Server) handleProviderConfig(w http.ResponseWriter, r *http.Request) bool {
	if !strings.HasPrefix(r.URL.Path, "/providers/") || r.Method != http.MethodPost {
		return false
	}
	if s.secret == nil {
		writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "secret_provider_missing", Message: "secret provider not configured"})
		return true
	}

	raw := strings.TrimPrefix(r.URL.Path, "/providers/")
	capability, ok := normalizeCapability(raw)
	if !ok {
		writeError(w, http.StatusNotFound, orcherr.OpsOrchError{Code: "not_found", Message: "unknown capability"})
		return true
	}
	var req providerConfigRequest
	if err := decodeJSON(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
		return true
	}
	if req.Provider == "" {
		writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: "provider required"})
		return true
	}

	var applyErr error
	switch capability {
	case "incident":
		applyErr = s.handleIncidentProviderConfig(req.Provider, req.Plugin, req.Config)
	case "log":
		applyErr = s.handleLogProviderConfig(req.Provider, req.Plugin, req.Config)
	case "metric":
		applyErr = s.handleMetricProviderConfig(req.Provider, req.Plugin, req.Config)
	case "ticket":
		applyErr = s.handleTicketProviderConfig(req.Provider, req.Plugin, req.Config)
	case "messaging":
		applyErr = s.handleMessagingProviderConfig(req.Provider, req.Plugin, req.Config)
	case "service":
		applyErr = s.handleServiceProviderConfig(req.Provider, req.Plugin, req.Config)
	default:
		writeError(w, http.StatusNotFound, orcherr.OpsOrchError{Code: "not_found", Message: "unknown capability"})
		return true
	}

	if applyErr != nil {
		writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: applyErr.Error()})
		return true
	}

	// Persist config via secret provider for reuse.
	if err := s.storeProviderConfig(capability, req); err != nil {
		writeError(w, http.StatusBadGateway, orcherr.OpsOrchError{Code: "secret_store_error", Message: err.Error()})
		return true
	}

	logMutatingAction(r, "configure_provider", "provider", capability)

	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	return true
}

func (s *Server) storeProviderConfig(capability string, req providerConfigRequest) error {
	bytes, err := json.Marshal(req)
	if err != nil {
		return err
	}
	key := providerConfigKey(capability)
	return s.secret.Put(rctx(), key, string(bytes))
}

func loadProviderConfig(sec SecretProvider, capability, envProvider, envConfig, envPlugin string) (string, map[string]any, string, error) {
	name := strings.TrimSpace(strings.ToLower(os.Getenv(envProvider)))
	pluginPath := strings.TrimSpace(os.Getenv(envPlugin))
	cfg, err := decodeConfig(envConfig)
	if err != nil {
		return "", nil, "", err
	}
	if name != "" || pluginPath != "" {
		return name, cfg, pluginPath, nil
	}
	if sec == nil {
		return "", nil, "", nil
	}
	raw, err := sec.Get(rctx(), providerConfigKey(capability))
	if err != nil {
		return "", nil, "", nil
	}
	var stored providerConfigRequest
	if err := json.Unmarshal([]byte(raw), &stored); err != nil {
		return "", nil, "", err
	}
	return stored.Provider, stored.Config, stored.Plugin, nil
}

func providerConfigKey(capability string) string {
	return fmt.Sprintf("providers/%s/default", strings.ToLower(capability))
}

// rctx returns a background context; used for secret provider calls.
func rctx() context.Context {
	return context.Background()
}
