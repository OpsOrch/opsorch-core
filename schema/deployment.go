package schema

import "time"

// DeploymentQuery filters normalized deployments from the active deployment provider.
// Providers choose how to map DeploymentQuery to their upstream search/filter API.
type DeploymentQuery struct {
	// Query is a free-form search string providers can map to ref, version,
	// commit message, description, or provider-specific syntax.
	Query string `json:"query,omitempty"`

	// Statuses filters deployments by one or more normalized status values.
	// Recommended normalized set: "queued", "running", "success", "failed", "cancelled".
	Statuses []string `json:"statuses,omitempty"`

	// Versions filters deployments by one or more version identifiers
	// (e.g. image tag, semantic version, build label, ref, short SHA).
	Versions []string `json:"versions,omitempty"`

	// Scope provides shared service/team/environment hints. Providers can ignore fields
	// they do not support.
	Scope QueryScope `json:"scope,omitempty"`

	// Limit caps the maximum number of deployments returned.
	Limit int `json:"limit,omitempty"`

	// Metadata carries provider-specific filter hints (e.g. project keys, pipeline IDs,
	// repo identifiers, branch selectors).
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Deployment captures the normalized deployment shape for the current schema revision.
// OpsOrch only queries deployments in the initial iteration; providers own the source of truth.
type Deployment struct {
	// ID is the stable, normalized OpsOrch identifier for this deployment.
	// Typically derived from the provider's unique field (run id, build id, deployment id).
	ID string `json:"id"`

	// Service is the canonical service name.
	// Providers map this to repo/project/application identifiers and tags.
	Service string `json:"service,omitempty"`

	// Environment is a simple "prod", "staging", "dev"-style value.
	// Providers map this to environment names, stages, or deployment targets.
	Environment string `json:"environment,omitempty"`

	// Version is a provider-agnostic identifier for the deployed version
	// (image tag, semantic version, build number, ref, commit SHA).
	Version string `json:"version,omitempty"`

	// Status is the normalized deployment state
	// (queued, running, success, failed, cancelled).
	Status string `json:"status"`

	// StartedAt is when the deployment began rolling out.
	StartedAt time.Time `json:"startedAt"`

	// FinishedAt is when the deployment completed (successfully or not).
	FinishedAt time.Time `json:"finishedAt"`

	// URL is the upstream link to the deployment (e.g. GitHub Actions run, Jenkins build).
	URL string `json:"url,omitempty"`

	// Actor represents who triggered the deployment.
	Actor map[string]any `json:"actor,omitempty"`

	// Fields is a flexible bag for commonly useful structured attributes
	// (commit SHA, branch, pipeline/plan key, actor, URLs).
	Fields map[string]any `json:"fields,omitempty"`

	// Metadata is for provider-specific payload fragments that don't map cleanly into the
	// normalized shape (raw JSON, upstream objects, etc.).
	Metadata map[string]any `json:"metadata,omitempty"`
}
