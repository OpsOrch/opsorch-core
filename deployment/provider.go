package deployment

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface a deployment adapter must satisfy.
// Query-only in the initial iteration.
type Provider interface {
	// Query returns deployments matching the given filters. Providers decide how to map
	// DeploymentQuery to upstream APIs (GitHub, GitLab, Bamboo, Argo, Jenkins, etc.).
	Query(ctx context.Context, query schema.DeploymentQuery) ([]schema.Deployment, error)

	// Get returns a single deployment by its normalized ID.
	Get(ctx context.Context, id string) (schema.Deployment, error)
}

// ProviderConstructor builds a Provider instance from decrypted config.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider adds a deployment provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a named provider constructor if registered.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists all registered deployment provider names.
func Providers() []string {
	return providers.Names()
}
