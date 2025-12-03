package alert

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface an alert adapter must satisfy.
//
// Alerts in OpsOrch are read-only signals: adapters map upstream alerts into
// the normalized schema.Alert shape. Providers do NOT support create/update.
type Provider interface {
	// Query returns alerts that match the AlertQuery filter.
	Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error)

	// Get returns a single alert by its normalized ID. Providers that cannot
	// retrieve alerts by ID (e.g., ephemeral alerts) may return an error.
	Get(ctx context.Context, id string) (schema.Alert, error)
}

// ProviderConstructor builds a Provider instance from decrypted config.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider adds an alert provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a named alert provider constructor if registered.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists all registered alert provider names.
func Providers() []string {
	return providers.Names()
}
