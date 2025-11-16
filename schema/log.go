package schema

import "time"

// LogQuery requests a slice of normalized log entries.
type LogQuery struct {
	Query     string         `json:"query"`
	Start     time.Time      `json:"start"`
	End       time.Time      `json:"end"`
	Scope     QueryScope     `json:"scope,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Providers []string       `json:"providers,omitempty"`
}

// LogEntry is a normalized log record.
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string            `json:"message"`
	Severity  string            `json:"severity,omitempty"`
	Service   string            `json:"service,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`   // filterable
	Fields    map[string]any    `json:"fields,omitempty"`   // structured JSON
	Metadata  map[string]any    `json:"metadata,omitempty"` // provider-specific
}
