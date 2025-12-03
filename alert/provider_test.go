package alert

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

type stubAlertProvider struct{}

func (stubAlertProvider) Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error) {
	return nil, nil
}

func (stubAlertProvider) Get(ctx context.Context, id string) (schema.Alert, error) {
	return schema.Alert{}, nil
}

func TestAlertRegisterLookup(t *testing.T) {
	name := "test-alert"
	ctor := func(cfg map[string]any) (Provider, error) { return stubAlertProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-alert already registered" {
		t.Fatalf("register: %v", err)
	}
	got, ok := LookupProvider(name)
	if !ok || got == nil {
		t.Fatalf("expected provider lookup success")
	}
	names := Providers()
	found := false
	for _, n := range names {
		if n == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected provider name in list: %v", names)
	}
}

func TestAlertDuplicateFails(t *testing.T) {
	name := "dup-alert"
	ctor := func(cfg map[string]any) (Provider, error) { return stubAlertProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
