package secret

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
)

// Provider is the abstraction for secret backends (Vault, AWS KMS, GCP KMS, local, etc.).
// Implementations must be stateless and receive all config through the constructor.
type Provider interface {
	// Get returns the plaintext value for a logical key.
	Get(ctx context.Context, key string) (string, error)
	// Put stores a plaintext value at the logical key, creating or updating as needed.
	Put(ctx context.Context, key, value string) error
}

// ProviderConstructor builds a Provider from decrypted config.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider registers a secret backend provider by name (case-insensitive).
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider finds a provider constructor by name.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists registered secret provider names.
func Providers() []string {
	return providers.Names()
}
