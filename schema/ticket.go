package schema

import "time"

// TicketQuery defines filters for querying tickets.
// Query is a free-form search string providers can map to JQL, name, description, etc.
type TicketQuery struct {
	Query     string         `json:"query,omitempty"`
	Statuses  []string       `json:"statuses,omitempty"`
	Assignees []string       `json:"assignees,omitempty"`
	Reporter  string         `json:"reporter,omitempty"`
	Scope     QueryScope     `json:"scope,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

// Ticket is a normalized representation of a work item.
type Ticket struct {
	ID          string         `json:"id"`
	Key         string         `json:"key,omitempty"`
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Status      string         `json:"status"`
	Assignees   []string       `json:"assignees,omitempty"`
	Reporter    string         `json:"reporter,omitempty"`
	CreatedAt   time.Time      `json:"createdAt"`
	UpdatedAt   time.Time      `json:"updatedAt"`
	Fields      map[string]any `json:"fields,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// CreateTicketInput defines the payload for ticket creation.
type CreateTicketInput struct {
	Title       string         `json:"title"`
	Description string         `json:"description,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}

// UpdateTicketInput defines mutable fields on a ticket.
type UpdateTicketInput struct {
	Title       *string        `json:"title,omitempty"`
	Description *string        `json:"description,omitempty"`
	Status      *string        `json:"status,omitempty"`
	Assignees   *[]string      `json:"assignees,omitempty"`
	Fields      map[string]any `json:"fields,omitempty"`
	Metadata    map[string]any `json:"metadata,omitempty"`
}
