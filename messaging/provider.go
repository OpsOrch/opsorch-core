package messaging

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface for messaging adapters.
type Provider interface {
	Send(ctx context.Context, message schema.Message) (schema.MessageResult, error)
}

// ProviderConstructor builds a messaging provider from decrypted configuration.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider registers a messaging provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a registered provider constructor by name.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists registered messaging provider names.
func Providers() []string {
	return providers.Names()
}
