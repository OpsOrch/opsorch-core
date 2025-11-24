package metric

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface for metric adapters.
type Provider interface {
	Query(ctx context.Context, query schema.MetricQuery) ([]schema.MetricSeries, error)
	Describe(ctx context.Context, scope schema.QueryScope) ([]schema.MetricDescriptor, error)
}

// ProviderConstructor builds a metric provider from decrypted configuration.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider registers a metric provider constructor.
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
