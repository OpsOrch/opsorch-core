package schema

// QueryScope is the minimal shared filter used across all OpsOrch queries.
// Providers can choose to ignore fields they don't understand.
type QueryScope struct {
	// Service is the canonical service name.
	// Providers map this to service tags, project IDs, components, etc.
	Service string `json:"service,omitempty"`

	// Team is the team owning the workload or resource.
	// Providers map this to escalation policies, Jira components, Datadog tags, etc.
	Team string `json:"team,omitempty"`

	// Environment is a simple "prod", "staging", "dev"-style filter.
	// Providers map this to labels/tags in their own systems.
	Environment string `json:"environment,omitempty"`
}
