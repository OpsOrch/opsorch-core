package log

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface for log adapters.
type Provider interface {
	Query(ctx context.Context, query schema.LogQuery) ([]schema.LogEntry, error)
}

// ProviderConstructor builds a log provider from decrypted configuration.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider registers a log provider.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a registered provider constructor by name.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers returns registered adapter names.
func Providers() []string {
	return providers.Names()
}
