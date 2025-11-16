package service

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

type stubServiceProvider struct{}

func (stubServiceProvider) Query(ctx context.Context, q schema.ServiceQuery) ([]schema.Service, error) {
	return nil, nil
}

func TestServiceRegisterLookup(t *testing.T) {
	name := "test-service"
	ctor := func(cfg map[string]any) (Provider, error) { return stubServiceProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-service already registered" {
		t.Fatalf("register: %v", err)
	}
	_, ok := LookupProvider(name)
	if !ok {
		t.Fatalf("expected provider lookup success")
	}
	names := Providers()
	found := false
	for _, n := range names {
		if n == name {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected provider name in list: %v", names)
	}
}

func TestServiceDuplicateFails(t *testing.T) {
	name := "dup-service"
	ctor := func(cfg map[string]any) (Provider, error) { return stubServiceProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
