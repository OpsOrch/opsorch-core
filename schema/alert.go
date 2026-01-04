package schema

import "time"

// AlertQuery filters normalized alerts from the active alert provider.
// Providers choose how to map these hints to their upstream search/filter API.
type AlertQuery struct {
	// Query is a free-form search string providers can map to title/message or
	// provider-specific search syntax.
	Query string `json:"query,omitempty"`

	// Statuses filters alerts by one or more normalized status values
	// (e.g. "open", "closed", "firing", "resolved").
	Statuses []string `json:"statuses,omitempty"`

	// Severities filters alerts by normalized severity values
	// (e.g. "info", "warning", "error", "critical", "P1").
	Severities []string `json:"severities,omitempty"`

	// Scope provides shared service/team/environment hints.
	Scope QueryScope `json:"scope,omitempty"`

	// Limit caps the maximum number of alerts returned.
	Limit int `json:"limit,omitempty"`

	// Metadata carries provider-specific filter hints (e.g. label selectors,
	// monitor types, project IDs).
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Alert captures the normalized alert shape for the current schema revision.
// It is intentionally similar to Incident but represents a lower-level signal.
type Alert struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	Status      string    `json:"status"`
	Severity    string    `json:"severity"`
	Service     string    `json:"service,omitempty"`
	URL         string    `json:"url,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`

	// Fields is a flexible bag for commonly useful, structured attributes
	// (e.g. metric name, monitor/issue ID, hostname, region, tags).
	Fields map[string]any `json:"fields,omitempty"`

	// Metadata is for provider-specific payload fragments that don't map cleanly
	// into the normalized shape (raw JSON, rule definitions, etc.).
	Metadata map[string]any `json:"metadata,omitempty"`
}
