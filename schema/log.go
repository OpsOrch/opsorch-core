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
	Providers  []string       `json:"providers,omitempty"`
}

// LogExpression defines structured log search criteria.
type LogExpression struct {
	Search     string      `json:"search,omitempty"`     // Full-text search term
	Filters    []LogFilter `json:"filters,omitempty"`    // Structured filters
	SeverityIn []string    `json:"severityIn,omitempty"` // Filter by severity levels
}

// LogFilter defines a field-level filter for logs.
type LogFilter struct {
	Field    string `json:"field"`    // Field name (e.g., "service", "message")
	Operator string `json:"operator"` // "=", "!=", "contains", "regex"
	Value    string `json:"value"`    // Filter value
}

// LogEntry is a normalized log record.
type LogEntry struct {
	Timestamp time.Time         `json:"timestamp"`
	Message   string            `json:"message"`
	Severity  string            `json:"severity,omitempty"`
	Service   string            `json:"service,omitempty"`
	URL       string            `json:"url,omitempty"`
	Labels    map[string]string `json:"labels,omitempty"`   // filterable
	Fields    map[string]any    `json:"fields,omitempty"`   // structured JSON
	Metadata  map[string]any    `json:"metadata,omitempty"` // provider-specific
}
