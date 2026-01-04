package schema

import "time"

// MetricQuery specifies a time-series request.
//
// Example:
//
//	{
//	  "expression": {
//	    "metricName": "http_requests_total",
//	    "aggregation": "sum",
//	    "filters": [
//	      {"label": "method", "operator": "=", "value": "POST"}
//	    ],
//	    "groupBy": ["service", "status"]
//	  },
//	  "start": "2023-10-01T00:00:00Z",
//	  "end": "2023-10-01T01:00:00Z",
//	  "step": 60
//	}
type MetricQuery struct {
	Expression *MetricExpression `json:"expression,omitempty"`
	Start      time.Time         `json:"start"`
	End        time.Time         `json:"end"`
	Step       int               `json:"step"` // in seconds
	Scope      QueryScope        `json:"scope,omitempty"`
	Metadata   map[string]any    `json:"metadata,omitempty"`
}

// MetricExpression defines structured metric selection criteria.
type MetricExpression struct {
	MetricName  string         `json:"metricName"`            // Required: metric name (e.g., "http_requests_total")
	Aggregation string         `json:"aggregation,omitempty"` // "avg", "sum", "max", "min", "count"
	Filters     []MetricFilter `json:"filters,omitempty"`     // Label filters
	GroupBy     []string       `json:"groupBy,omitempty"`     // Label keys to group by
}

// MetricFilter defines a label-based filter for metrics.
type MetricFilter struct {
	Label    string `json:"label"`    // Label key
	Operator string `json:"operator"` // "=", "!=", "=~", "!~"
	Value    string `json:"value"`    // Filter value
}

// MetricDescriptor describes an available metric for discovery.
type MetricDescriptor struct {
	Name        string         `json:"name"`               // Metric name (e.g., "http_requests_total")
	Type        string         `json:"type"`               // "counter", "gauge", "histogram"
	Description string         `json:"description"`        // Human-readable description
	Labels      []string       `json:"labels"`             // Available label keys
	Unit        string         `json:"unit,omitempty"`     // "bytes", "seconds", "requests"
	URL         string         `json:"url,omitempty"`      // Upstream link to the metric (e.g. Prometheus expression browser)
	Metadata    map[string]any `json:"metadata,omitempty"` // Provider-specific metadata
}

// MetricSeries is a normalized collection of metric points.
type MetricSeries struct {
	Name     string         `json:"name"`
	Service  string         `json:"service,omitempty"`
	Labels   map[string]any `json:"labels,omitempty"`
	Points   []MetricPoint  `json:"points"`
	URL      string         `json:"url,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// MetricPoint represents a single value in a series.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
