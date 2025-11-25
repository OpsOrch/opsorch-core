package secret

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
)

// JsonProvider implements a simple secret provider backed by a map.
// It can be initialized from a JSON file and supports nested configurations.
type JsonProvider struct {
	mu    sync.RWMutex
	store map[string]any
}

// NewJsonProvider creates a new JsonProvider.
// If the config contains a "path" key, it attempts to load secrets from that JSON file.
func NewJsonProvider(config map[string]any) (Provider, error) {
	store := make(map[string]any)

	// 1. Load from file if "path" is present
	if path, ok := config["path"].(string); ok && path != "" {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, fmt.Errorf("failed to read secret file %s: %w", path, err)
		}
		if err := json.Unmarshal(data, &store); err != nil {
			return nil, fmt.Errorf("failed to parse secret file %s: %w", path, err)
		}
	} else {
		return nil, fmt.Errorf("basic provider requires 'path' in config")
	}

	return &JsonProvider{
		store: store,
	}, nil
}

// Get returns the plaintext value for a logical key.
// If the value is a string, it's returned as-is.
// If the value is a complex object (map, array), it's marshaled to JSON.
func (j *JsonProvider) Get(ctx context.Context, key string) (string, error) {
	j.mu.RLock()
	defer j.mu.RUnlock()

	val, ok := j.store[key]
	if !ok {
		return "", fmt.Errorf("secret not found: %s", key)
	}

	// If it's already a string, return it
	if str, ok := val.(string); ok {
		return str, nil
	}

	// Otherwise, marshal to JSON
	jsonBytes, err := json.Marshal(val)
	if err != nil {
		return "", fmt.Errorf("failed to marshal secret %s: %w", key, err)
	}
	return string(jsonBytes), nil
}

// Put stores a plaintext value at the logical key.
// Note: This only updates the in-memory store and does not persist to disk.
func (j *JsonProvider) Put(ctx context.Context, key, value string) error {
	j.mu.Lock()
	defer j.mu.Unlock()

	j.store[key] = value
	return nil
}

func init() {
	RegisterProvider("json", NewJsonProvider)
}
