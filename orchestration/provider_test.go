package orchestration

import (
	"context"
	"testing"

	"github.com/opsorch/opsorch-core/schema"
)

// mockOrchestrationProvider implements Provider for testing.
type mockOrchestrationProvider struct {
	queryPlansFunc   func(ctx context.Context, query schema.OrchestrationPlanQuery) ([]schema.OrchestrationPlan, error)
	getPlanFunc      func(ctx context.Context, planID string) (*schema.OrchestrationPlan, error)
	queryRunsFunc    func(ctx context.Context, query schema.OrchestrationRunQuery) ([]schema.OrchestrationRun, error)
	getRunFunc       func(ctx context.Context, runID string) (*schema.OrchestrationRun, error)
	startRunFunc     func(ctx context.Context, planID string) (*schema.OrchestrationRun, error)
	completeStepFunc func(ctx context.Context, runID string, stepID string, actor string, note string) error
}

func (m *mockOrchestrationProvider) QueryPlans(ctx context.Context, query schema.OrchestrationPlanQuery) ([]schema.OrchestrationPlan, error) {
	if m.queryPlansFunc != nil {
		return m.queryPlansFunc(ctx, query)
	}
	return []schema.OrchestrationPlan{}, nil
}

func (m *mockOrchestrationProvider) GetPlan(ctx context.Context, planID string) (*schema.OrchestrationPlan, error) {
	if m.getPlanFunc != nil {
		return m.getPlanFunc(ctx, planID)
	}
	return nil, nil
}

func (m *mockOrchestrationProvider) QueryRuns(ctx context.Context, query schema.OrchestrationRunQuery) ([]schema.OrchestrationRun, error) {
	if m.queryRunsFunc != nil {
		return m.queryRunsFunc(ctx, query)
	}
	return []schema.OrchestrationRun{}, nil
}

func (m *mockOrchestrationProvider) GetRun(ctx context.Context, runID string) (*schema.OrchestrationRun, error) {
	if m.getRunFunc != nil {
		return m.getRunFunc(ctx, runID)
	}
	return nil, nil
}

func (m *mockOrchestrationProvider) StartRun(ctx context.Context, planID string) (*schema.OrchestrationRun, error) {
	if m.startRunFunc != nil {
		return m.startRunFunc(ctx, planID)
	}
	return nil, nil
}

func (m *mockOrchestrationProvider) CompleteStep(ctx context.Context, runID string, stepID string, actor string, note string) error {
	if m.completeStepFunc != nil {
		return m.completeStepFunc(ctx, runID, stepID, actor, note)
	}
	return nil
}

func TestRegisterProvider(t *testing.T) {
	name := "test-orch-provider"
	constructor := func(config map[string]any) (Provider, error) {
		return &mockOrchestrationProvider{}, nil
	}

	err := RegisterProvider(name, constructor)
	if err != nil {
		t.Fatalf("expected no error registering provider, got: %v", err)
	}

	// Verify provider is registered
	got, ok := LookupProvider(name)
	if !ok {
		t.Fatalf("expected provider %s to be registered", name)
	}
	if got == nil {
		t.Fatalf("expected non-nil constructor")
	}
}

func TestRegisterProviderDuplicate(t *testing.T) {
	name := "test-orch-duplicate"
	constructor := func(config map[string]any) (Provider, error) {
		return &mockOrchestrationProvider{}, nil
	}

	// First registration should succeed
	err := RegisterProvider(name, constructor)
	if err != nil {
		t.Fatalf("first registration should succeed: %v", err)
	}

	// Second registration should fail
	err = RegisterProvider(name, constructor)
	if err == nil {
		t.Fatalf("expected error for duplicate registration")
	}
}

func TestLookupProviderNotFound(t *testing.T) {
	_, ok := LookupProvider("nonexistent-provider")
	if ok {
		t.Fatalf("expected provider not to be found")
	}
}

func TestProviders(t *testing.T) {
	// Register a unique provider for this test
	name := "test-orch-list"
	constructor := func(config map[string]any) (Provider, error) {
		return &mockOrchestrationProvider{}, nil
	}
	_ = RegisterProvider(name, constructor)

	names := Providers()
	found := false
	for _, n := range names {
		if n == name {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected provider %s in list, got %v", name, names)
	}
}

func TestProviderInterface(t *testing.T) {
	// Verify that mockOrchestrationProvider implements Provider interface
	var _ Provider = (*mockOrchestrationProvider)(nil)
}

func TestProviderConstructorReturnsProvider(t *testing.T) {
	name := "test-orch-constructor"
	expectedConfig := map[string]any{"key": "value"}
	var receivedConfig map[string]any

	constructor := func(config map[string]any) (Provider, error) {
		receivedConfig = config
		return &mockOrchestrationProvider{}, nil
	}

	_ = RegisterProvider(name, constructor)

	got, ok := LookupProvider(name)
	if !ok {
		t.Fatalf("expected provider to be registered")
	}

	provider, err := got(expectedConfig)
	if err != nil {
		t.Fatalf("expected no error from constructor: %v", err)
	}
	if provider == nil {
		t.Fatalf("expected non-nil provider")
	}
	if receivedConfig["key"] != expectedConfig["key"] {
		t.Fatalf("expected config to be passed through, got %v", receivedConfig)
	}
}
