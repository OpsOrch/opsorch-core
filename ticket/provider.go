package ticket

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface for ticketing adapters.
type Provider interface {
	Query(ctx context.Context, query schema.TicketQuery) ([]schema.Ticket, error)
	Get(ctx context.Context, id string) (schema.Ticket, error)
	Create(ctx context.Context, in schema.CreateTicketInput) (schema.Ticket, error)
	Update(ctx context.Context, id string, in schema.UpdateTicketInput) (schema.Ticket, error)
}

// ProviderConstructor builds a ticket provider from decrypted configuration.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider registers a ticket provider constructor.
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
