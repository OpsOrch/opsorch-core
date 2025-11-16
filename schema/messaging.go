package schema

import "time"

type Message struct {
	Channel   string         `json:"channel"`
	Body      string         `json:"body"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	ThreadRef string         `json:"threadRef,omitempty"`
}

type MessageResult struct {
	ID       string         `json:"id"`
	Channel  string         `json:"channel"`
	SentAt   time.Time      `json:"sentAt"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
