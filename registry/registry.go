package registry

import (
	"fmt"
	"sort"
	"strings"
	"sync"
)

// Registry stores a mapping of provider names to constructors for a capability.
type Registry[C any] struct {
	mu        sync.RWMutex
	providers map[string]C
}

// New allocates a registry for provider constructors.
func New[C any]() *Registry[C] {
	return &Registry[C]{providers: make(map[string]C)}
}

// Register adds a provider constructor by name. Names are case-insensitive.
func (r *Registry[C]) Register(name string, constructor C) error {
	if name == "" {
		return fmt.Errorf("registry: provider name required")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	key := normalize(name)
	if _, exists := r.providers[key]; exists {
		return fmt.Errorf("registry: provider %s already registered", name)
	}

	r.providers[key] = constructor
	return nil
}

// Get fetches a provider constructor by name.
func (r *Registry[C]) Get(name string) (C, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	constructor, ok := r.providers[normalize(name)]
	return constructor, ok
}

// Names returns the sorted provider keys registered for this capability.
func (r *Registry[C]) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

func normalize(name string) string {
	return strings.ToLower(name)
}
