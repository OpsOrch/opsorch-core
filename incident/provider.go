package incident

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface an incident adapter must satisfy.
type Provider interface {
	Query(ctx context.Context, query schema.IncidentQuery) ([]schema.Incident, error)
	Get(ctx context.Context, id string) (schema.Incident, error)
	Create(ctx context.Context, in schema.CreateIncidentInput) (schema.Incident, error)
	Update(ctx context.Context, id string, in schema.UpdateIncidentInput) (schema.Incident, error)
	GetTimeline(ctx context.Context, id string) ([]schema.TimelineEntry, error)
	AppendTimeline(ctx context.Context, id string, entry schema.TimelineAppendInput) error
}

// ProviderConstructor builds a Provider instance from decrypted config.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider adds an incident provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a named provider constructor if registered.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists all registered incident provider names.
func Providers() []string {
	return providers.Names()
}
