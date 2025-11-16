package incident

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

type stubIncidentProvider struct{}

func (stubIncidentProvider) Query(ctx context.Context, query schema.IncidentQuery) ([]schema.Incident, error) {
	return nil, nil
}
func (stubIncidentProvider) List(ctx context.Context) ([]schema.Incident, error) { return nil, nil }
func (stubIncidentProvider) Get(ctx context.Context, id string) (schema.Incident, error) {
	return schema.Incident{}, nil
}
func (stubIncidentProvider) Create(ctx context.Context, in schema.CreateIncidentInput) (schema.Incident, error) {
	return schema.Incident{}, nil
}
func (stubIncidentProvider) Update(ctx context.Context, id string, in schema.UpdateIncidentInput) (schema.Incident, error) {
	return schema.Incident{}, nil
}
func (stubIncidentProvider) GetTimeline(ctx context.Context, id string) ([]schema.TimelineEntry, error) {
	return nil, nil
}
func (stubIncidentProvider) AppendTimeline(ctx context.Context, id string, entry schema.TimelineAppendInput) error {
	return nil
}

func TestIncidentRegisterLookup(t *testing.T) {
	name := "test-incident"
	ctor := func(cfg map[string]any) (Provider, error) { return stubIncidentProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-incident already registered" {
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

func TestIncidentDuplicateFails(t *testing.T) {
	name := "dup-incident"
	ctor := func(cfg map[string]any) (Provider, error) { return stubIncidentProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
