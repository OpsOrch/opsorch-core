package team

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface a team adapter must satisfy.
type Provider interface {
	// Query discovers teams by name/tags/scope.
	Query(ctx context.Context, query schema.TeamQuery) ([]schema.Team, error)

	// Get returns a single team by its canonical ID.
	Get(ctx context.Context, id string) (schema.Team, error)

	// Members returns a best-effort snapshot of members for a given team.
	// Callers pass the canonical team ID (schema.Team.ID).
	Members(ctx context.Context, teamID string) ([]schema.TeamMember, error)
}

// ProviderConstructor builds a team provider from decrypted configuration.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider adds a team provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a named provider constructor if registered.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists all registered team provider names.
func Providers() []string {
	return providers.Names()
}
