package deployment

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

type stubDeploymentProvider struct{}

func (stubDeploymentProvider) Query(ctx context.Context, q schema.DeploymentQuery) ([]schema.Deployment, error) {
	return nil, nil
}

func (stubDeploymentProvider) Get(ctx context.Context, id string) (schema.Deployment, error) {
	return schema.Deployment{}, nil
}

func TestDeploymentRegisterLookup(t *testing.T) {
	name := "test-deployment"
	ctor := func(cfg map[string]any) (Provider, error) { return stubDeploymentProvider{}, nil }
	if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-deployment already registered" {
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

func TestDeploymentDuplicateFails(t *testing.T) {
	name := "dup-deployment"
	ctor := func(cfg map[string]any) (Provider, error) { return stubDeploymentProvider{}, nil }
	_ = RegisterProvider(name, ctor)
	if err := RegisterProvider(name, ctor); err == nil {
		t.Fatalf("expected duplicate registration to fail")
	}
}
