package schema

import "time"

type Message struct {
	Channel   string         `json:"channel"`
	Body      string         `json:"body"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	ThreadRef string         `json:"threadRef,omitempty"`
	Blocks    []Block        `json:"blocks,omitempty"`
}

// BlockType defines the type of a UI block.
type BlockType string

const (
	BlockTypeHeader  BlockType = "header"
	BlockTypeSection BlockType = "section"
	BlockTypeDivider BlockType = "divider"
)

// Block represents a UI element in a message.
// Text fields support CommonMark Markdown. Adapters are responsible for converting
// this Markdown to the target platform's format (e.g., Slack mrkdwn).
type Block struct {
	Type   BlockType         `json:"type"`
	Text   string            `json:"text,omitempty"`
	Fields map[string]string `json:"fields,omitempty"`
}

type MessageResult struct {
	ID       string         `json:"id"`
	Channel  string         `json:"channel"`
	SentAt   time.Time      `json:"sentAt"`
	Metadata map[string]any `json:"metadata,omitempty"`
}
