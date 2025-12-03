package schema

import "time"

// IncidentQuery filters normalized incidents from the active incident provider.
// Providers choose how to map these hints to their upstream search/filter API.
type IncidentQuery struct {
	// Query is a free-form search string providers can map to title, summary, or
	// provider-specific search syntax.
	Query string `json:"query,omitempty"`

	// Statuses filters incidents by one or more normalized status values.
	Statuses []string `json:"statuses,omitempty"`

	// Severities filters incidents by normalized severity values.
	Severities []string `json:"severities,omitempty"`

	// Scope provides shared service/team/environment hints.
	Scope QueryScope `json:"scope,omitempty"`

	// Limit caps the maximum number of incidents returned.
	Limit int `json:"limit,omitempty"`

	// Metadata carries provider-specific filter hints.
	Metadata map[string]any `json:"metadata,omitempty"`
}

// Incident captures the normalized incident shape for the current schema revision.
type Incident struct {
	ID          string         `json:"id"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Status      string         `json:"status"`
	Severity    string         `json:"severity"`
	Service     string         `json:"service,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Fields      map[string]any `json:"fields,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreateIncidentInput is the provider-agnostic payload for creating an incident.
type CreateIncidentInput struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Status      string         `json:"status"`
	Severity    string         `json:"severity"`
	Service     string         `json:"service,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateIncidentInput defines mutable fields on an incident.
type UpdateIncidentInput struct {
	Title       *string        `json:"title,omitempty"`
	Description *string        `json:"description,omitempty"`
	Status      *string        `json:"status,omitempty"`
	Severity    *string        `json:"severity,omitempty"`
	Service     *string        `json:"service,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// TimelineEntry represents a single entry on an incident timeline.
type TimelineEntry struct {
	ID         string         `json:"id"`
	IncidentID string         `json:"incidentId"`
	At         time.Time      `json:"at"`
	Kind       string         `json:"kind"`
	Body       string         `json:"body"`
	Actor      map[string]any `json:"actor,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// TimelineAppendInput is used to append to an incident timeline.
type TimelineAppendInput struct {
	At       time.Time      `json:"at"`
	Kind     string         `json:"kind"`
	Body     string         `json:"body"`
	Actor    map[string]any `json:"actor,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
