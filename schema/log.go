package schema

import "time"

// LogQuery requests a slice of normalized log entries.
//
// Example:
//
//	{
//	  "expression": {
//	    "search": "error connection",
//	    "filters": [
//	      {"field": "service", "operator": "=", "value": "api-gateway"},
//	      {"field": "status", "operator": "!=", "value": "200"}
//	    ],
//	    "severityIn": ["error", "critical"]
//	  },
//	  "start": "2023-10-01T00:00:00Z",
//	  "end": "2023-10-01T01:00:00Z"
//	}
type LogQuery struct {
	Expression *LogExpression `json:"expression,omitempty"`
	Start      time.Time      `json:"start"`
	End        time.Time      `json:"end"`
	Scope      QueryScope     `json:"scope,omitempty"`
	Limit      int            `json:"limit,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// LogExpression defines structured log search criteria.
type LogExpression struct {
	Search     string      `json:"search,omitempty"`     // Full-text search term
	Filters    []LogFilter `json:"filters,omitempty"`    // Structured filters
	SeverityIn []string    `json:"severityIn,omitempty"` // Filter by severity levels (normalized)
}

// LogFilter defines a field-level filter for logs.
type LogFilter struct {
	Field    string `json:"field"`    // Field name (e.g., "service", "message", "@http.status_code")
	Operator string `json:"operator"` // "=", "!=", "contains", "regex" (provider may not support all)
	Value    string `json:"value"`    // Filter value (adapter decides casting)
}

// LogEntry is a normalized log record.
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string            `json:"message"`
	Severity  string            `json:"severity,omitempty"` // map from provider severity/status
	Service   string            `json:"service,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`   // filterable tags (key:value)
	Fields    map[string]any    `json:"fields,omitempty"`   // structured JSON (attributes)
	Metadata  map[string]any    `json:"metadata,omitempty"` // provider-specific (raw event, ids, etc.)
}

// LogEntries represents a collection of log entries with optional URL to view in source system.
type LogEntries struct {
	Entries []LogEntry `json:"entries"`
	URL     string     `json:"url,omitempty"` // Link to view these results in the log system (e.g., Datadog)
}
