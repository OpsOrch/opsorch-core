package log

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

type stubLogProvider struct{}

func (stubLogProvider) Query(ctx context.Context, q schema.LogQuery) (schema.LogEntries, error) {
	return schema.LogEntries{}, nil
}

func TestLogRegisterLookup(t *testing.T) {
	name := "test-log"
	ctor := func(cfg map[string]any) (Provider, error) { return stubLogProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-log already registered" {
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

func TestLogDuplicateFails(t *testing.T) {
	name := "dup-log"
	ctor := func(cfg map[string]any) (Provider, error) { return stubLogProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
