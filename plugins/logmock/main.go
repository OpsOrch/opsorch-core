package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

type rpcRequest struct {
	Method  string          `json:"method"`
	Config  map[string]any  `json:"config"`
	Payload json.RawMessage `json:"payload"`
}

type rpcResponse struct {
	Result any    `json:"result,omitempty"`
	Error  string `json:"error,omitempty"`
}

type LogEntries struct {
	Entries []LogEntry `json:"entries"`
	URL     string     `json:"url,omitempty"`
}
type LogEntry struct {
	Timestamp time.Time      `json:"timestamp"`
	Message   string         `json:"message"`
	Severity  string         `json:"severity,omitempty"`
	Service   string         `json:"service,omitempty"`
	Labels    map[string]any `json:"labels,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type QueryScope struct {
	Service     string `json:"service,omitempty"`
	Team        string `json:"team,omitempty"`
	Environment string `json:"environment,omitempty"`
}

type LogQuery struct {
	Query     string         `json:"query"`
	Start     time.Time      `json:"start"`
	End       time.Time      `json:"end"`
	Scope     QueryScope     `json:"scope,omitempty"`
	Limit     int            `json:"limit,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	Providers []string       `json:"providers,omitempty"`
}

func main() {
	dec := json.NewDecoder(os.Stdin)
	for {
		var req rpcRequest
		if err := dec.Decode(&req); err != nil {
			if err.Error() == "EOF" {
				return
			}
			writeErr(err)
			return
		}

		switch req.Method {
		case "log.query":
			var q LogQuery
			if err := json.Unmarshal(req.Payload, &q); err != nil {
				writeErr(err)
				continue
			}
			res := LogEntries{
				Entries: []LogEntry{{
					Timestamp: time.Now(),
					Message:   fmt.Sprintf("plugin log: %s", q.Query),
					Severity:  "info",
					Service:   q.Scope.Service,
				}},
				URL: "https://logs.example.com/query?q=" + q.Query,
			}
			writeOK(res)
		default:
			writeErr(fmt.Errorf("unknown method %s", req.Method))
		}
	}
}

func writeOK(result any) {
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(rpcResponse{Result: result})
}

func writeErr(err error) {
	enc := json.NewEncoder(os.Stdout)
	_ = enc.Encode(rpcResponse{Error: err.Error()})
}
