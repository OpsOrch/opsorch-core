package schema

import "time"

// MetricQuery specifies a time-series request.
type MetricQuery struct {
	Expression string         `json:"expression"`
	Start      time.Time      `json:"start"`
	End        time.Time      `json:"end"`
	Step       int            `json:"step"` // in seconds
	Scope      QueryScope     `json:"scope,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

// MetricSeries is a normalized collection of metric points.
type MetricSeries struct {
	Name     string         `json:"name"`
	Service  string         `json:"service,omitempty"`
	Labels   map[string]any `json:"labels,omitempty"`
	Points   []MetricPoint  `json:"points"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

// MetricPoint represents a single value in a series.
type MetricPoint struct {
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
}
