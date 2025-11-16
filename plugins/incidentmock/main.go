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

type Incident struct {
	ID        string         `json:"id"`
	Title     string         `json:"title"`
	Status    string         `json:"status"`
	Severity  string         `json:"severity"`
	Service   string         `json:"service,omitempty"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	Fields    map[string]any `json:"fields,omitempty"`
	Metadata  map[string]any `json:"metadata,omitempty"`
}

type CreateIncidentInput struct {
	Title    string         `json:"title"`
	Status   string         `json:"status"`
	Severity string         `json:"severity"`
	Service  string         `json:"service,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type UpdateIncidentInput struct {
	Title    *string        `json:"title,omitempty"`
	Status   *string        `json:"status,omitempty"`
	Severity *string        `json:"severity,omitempty"`
	Service  *string        `json:"service,omitempty"`
	Fields   map[string]any `json:"fields,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type TimelineEntry struct {
	ID         string         `json:"id"`
	IncidentID string         `json:"incidentId"`
	At         time.Time      `json:"at"`
	Kind       string         `json:"kind"`
	Body       string         `json:"body"`
	Actor      map[string]any `json:"actor,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

type TimelineAppendInput struct {
	At       time.Time      `json:"at"`
	Kind     string         `json:"kind"`
	Body     string         `json:"body"`
	Actor    map[string]any `json:"actor,omitempty"`
	Metadata map[string]any `json:"metadata,omitempty"`
}

type IncidentQuery struct {
	Query      string         `json:"query,omitempty"`
	Statuses   []string       `json:"statuses,omitempty"`
	Severities []string       `json:"severities,omitempty"`
	Scope      map[string]any `json:"scope,omitempty"`
	Limit      int            `json:"limit,omitempty"`
	Metadata   map[string]any `json:"metadata,omitempty"`
}

var incidents = []Incident{{ID: "p1", Title: "plugin incident", Status: "open", Severity: "sev2", Service: "svc-plugin", CreatedAt: time.Now(), UpdatedAt: time.Now()}}

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
		case "incident.query":
			var query IncidentQuery
			if err := json.Unmarshal(req.Payload, &query); err != nil {
				writeErr(err)
				continue
			}
			res := incidents
			if query.Limit > 0 && query.Limit < len(res) {
				res = res[:query.Limit]
			}
			writeOK(res)
		case "incident.list":
			writeOK(incidents)
		case "incident.get":
			var payload struct {
				ID string `json:"id"`
			}
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				writeErr(err)
				continue
			}
			found := false
			for _, inc := range incidents {
				if inc.ID == payload.ID {
					writeOK(inc)
					found = true
					break
				}
			}
			if !found {
				writeErr(fmt.Errorf("not found"))
			}
		case "incident.create":
			var in CreateIncidentInput
			if err := json.Unmarshal(req.Payload, &in); err != nil {
				writeErr(err)
				continue
			}
			newInc := Incident{ID: fmt.Sprintf("p%d", len(incidents)+1), Title: in.Title, Status: in.Status, Severity: in.Severity, Service: in.Service, CreatedAt: time.Now(), UpdatedAt: time.Now()}
			incidents = append(incidents, newInc)
			writeOK(newInc)
		case "incident.update":
			var payload struct {
				ID    string              `json:"id"`
				Input UpdateIncidentInput `json:"input"`
			}
			if err := json.Unmarshal(req.Payload, &payload); err != nil {
				writeErr(err)
				continue
			}
			updated := false
			for i, inc := range incidents {
				if inc.ID == payload.ID {
					if payload.Input.Title != nil {
						inc.Title = *payload.Input.Title
					}
					if payload.Input.Status != nil {
						inc.Status = *payload.Input.Status
					}
					if payload.Input.Severity != nil {
						inc.Severity = *payload.Input.Severity
					}
					if payload.Input.Service != nil {
						inc.Service = *payload.Input.Service
					}
					inc.UpdatedAt = time.Now()
					incidents[i] = inc
					writeOK(inc)
					updated = true
					break
				}
			}
			if !updated {
				writeErr(fmt.Errorf("not found"))
			}
		case "incident.timeline.get":
			writeOK([]TimelineEntry{{ID: "t1", IncidentID: "p1", At: time.Now(), Kind: "note", Body: "from plugin"}})
		case "incident.timeline.append":
			writeOK(map[string]string{"status": "ok"})
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
