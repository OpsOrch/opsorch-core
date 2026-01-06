package schema

import "time"

// ---- Plans ----

// OrchestrationPlanQuery filters plans from the orchestration provider.
type OrchestrationPlanQuery struct {
	// Query is a free-form search string for plan title or description.
	Query string `json:"query,omitempty"`

	// Scope provides shared service/team/environment hints.
	Scope QueryScope `json:"scope,omitempty"`

	// Tags filters plans by key-value tags.
	Tags map[string]string `json:"tags,omitempty"`

	// Limit caps the maximum number of plans returned.
	Limit int `json:"limit,omitempty"`

	// Metadata carries provider-specific filter hints.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// OrchestrationPlan is a provider-owned template describing ordered steps.
type OrchestrationPlan struct {
	ID          string `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`

	// Steps is the ordered list of steps in this plan.
	Steps []OrchestrationStep `json:"steps"`

	// URL is an upstream link to view/edit the plan in the provider system.
	URL string `json:"url,omitempty"`

	// Version is provider-defined (git sha, revision, updated timestamp, etc.).
	Version string `json:"version,omitempty"`

	// Tags are key-value labels for filtering and organization.
	Tags map[string]string `json:"tags,omitempty"`

	// Fields carries provider-specific structured data.
	Fields map[string]any `json:"fields,omitempty"`

	// Metadata carries provider-specific unstructured data.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// OrchestrationStep is a single unit of work within a plan.
type OrchestrationStep struct {
	ID    string `json:"id"`
	Title string `json:"title"`

	// Type is a normalized hint: "manual", "observe", "invoke", "verify", "record".
	Type string `json:"type,omitempty"`

	// Description is operator-facing text. May include Markdown for manual steps.
	Description string `json:"description,omitempty"`

	// DependsOn lists step IDs that must complete before this step can start.
	DependsOn []string `json:"dependsOn,omitempty"`

	// Fields carries provider-specific structured data.
	Fields map[string]any `json:"fields,omitempty"`

	// Metadata carries provider-specific unstructured data.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// ---- Runs ----

// OrchestrationRunQuery filters runs from the orchestration provider.
type OrchestrationRunQuery struct {
	// Query is a free-form search string.
	Query string `json:"query,omitempty"`

	// Statuses filters runs by status: "created", "running", "blocked", "completed", "failed", "cancelled".
	Statuses []string `json:"statuses,omitempty"`

	// PlanIDs filters runs by plan ID.
	PlanIDs []string `json:"planIds,omitempty"`

	// Scope provides shared service/team/environment hints.
	Scope QueryScope `json:"scope,omitempty"`

	// Limit caps the maximum number of runs returned.
	Limit int `json:"limit,omitempty"`

	// Metadata carries provider-specific filter hints.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// OrchestrationRun is a live instance of a plan with runtime state.
type OrchestrationRun struct {
	ID     string `json:"id"`
	PlanID string `json:"planId"`

	// Plan is optional denormalization for UI convenience.
	Plan *OrchestrationPlan `json:"plan,omitempty"`

	// Status is the overall run state: "created", "running", "blocked", "completed", "failed", "cancelled".
	Status string `json:"status"`

	// Scope is the service/team/environment context for this run.
	Scope QueryScope `json:"scope,omitempty"`

	// Steps is the runtime state for each plan step.
	Steps []OrchestrationStepState `json:"steps"`

	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`

	// URL is an upstream link to the run in the provider system.
	URL string `json:"url,omitempty"`

	// Fields carries provider-specific structured data.
	Fields map[string]any `json:"fields,omitempty"`

	// Metadata carries provider-specific unstructured data.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// OrchestrationStepState is the runtime state of a single step.
type OrchestrationStepState struct {
	StepID string `json:"stepId"`

	// Status is the step state: "pending", "ready", "running", "blocked", "succeeded", "failed", "skipped", "cancelled".
	Status string `json:"status"`

	// Actor is the user or system that completed this step (free text).
	Actor string `json:"actor,omitempty"`

	// Note is an optional completion note.
	Note string `json:"note,omitempty"`

	StartedAt  *time.Time `json:"startedAt,omitempty"`
	FinishedAt *time.Time `json:"finishedAt,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`

	// Fields carries provider-specific structured data.
	Fields map[string]any `json:"fields,omitempty"`

	// Metadata carries provider-specific unstructured data.
	Metadata map[string]any `json:"metadata,omitempty"`
}
