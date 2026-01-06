package orchestration

import (
	"context"

	"github.com/opsorch/opsorch-core/registry"
	"github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface an orchestration adapter must satisfy.
type Provider interface {
	// QueryPlans returns plans matching the query.
	QueryPlans(ctx context.Context, query schema.OrchestrationPlanQuery) ([]schema.OrchestrationPlan, error)

	// GetPlan returns a single plan by ID.
	GetPlan(ctx context.Context, planID string) (*schema.OrchestrationPlan, error)

	// QueryRuns returns runs matching the query.
	QueryRuns(ctx context.Context, query schema.OrchestrationRunQuery) ([]schema.OrchestrationRun, error)

	// GetRun returns a single run by ID, including current step states.
	GetRun(ctx context.Context, runID string) (*schema.OrchestrationRun, error)

	// StartRun creates a new run from a plan.
	StartRun(ctx context.Context, planID string) (*schema.OrchestrationRun, error)

	// CompleteStep marks a manual/blocked step as complete.
	CompleteStep(ctx context.Context, runID string, stepID string, actor string, note string) error
}

// ProviderConstructor builds a Provider instance from decrypted config.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider adds an orchestration provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
	return providers.Register(name, constructor)
}

// LookupProvider returns a named provider constructor if registered.
func LookupProvider(name string) (ProviderConstructor, bool) {
	return providers.Get(name)
}

// Providers lists all registered orchestration provider names.
func Providers() []string {
	return providers.Names()
}
