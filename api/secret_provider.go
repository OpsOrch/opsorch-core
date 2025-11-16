package api

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/opsorch/opsorch-core/secret"
)

// SecretProvider is a thin alias to the secret.Provider interface for API wiring.
type SecretProvider = secret.Provider

// newSecretProviderFromEnv loads the secret backend from environment variables.
// OPSORCH_SECRET_PROVIDER is the provider name; OPSORCH_SECRET_PLUGIN is a local plugin path; OPSORCH_SECRET_CONFIG contains JSON config for that provider.
func newSecretProviderFromEnv() (SecretProvider, error) {
	name := strings.TrimSpace(os.Getenv("OPSORCH_SECRET_PROVIDER"))
	pluginPath := strings.TrimSpace(os.Getenv("OPSORCH_SECRET_PLUGIN"))
	cfg, err := decodeSecretConfig()
	if err != nil {
		return nil, err
	}
	if pluginPath != "" {
		return newSecretPluginProvider(pluginPath, cfg), nil
	}
	if name == "" {
		return nil, nil
	}
	constructor, ok := secret.LookupProvider(name)
	if !ok {
		return nil, fmt.Errorf("secret provider %s not registered", name)
	}
	return constructor(cfg)
}

func decodeSecretConfig() (map[string]any, error) {
	raw := os.Getenv("OPSORCH_SECRET_CONFIG")
	if raw == "" {
		return map[string]any{}, nil
	}
	var cfg map[string]any
	if err := json.Unmarshal([]byte(raw), &cfg); err != nil {
		return nil, fmt.Errorf("invalid OPSORCH_SECRET_CONFIG: %w", err)
	}
	return cfg, nil
}
