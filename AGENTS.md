# OpsOrch Adapter Development Guide

This document explains how to build **external provider adapters** for OpsOrch Core.

OpsOrch Core is a **stateless orchestration layer**.  
It provides unified APIs, routing, secret management, and schema boundaries — but **does not define or enforce the exact shape of incidents, logs, metrics, tickets, or messages**.

Those schemas will evolve during implementation.  
Adapters must **conform to whatever schema is currently defined**, but this guide will not prescribe their structure.

---

## 1. Architecture Overview

OpsOrch Core delegates all provider-specific logic to external adapters.

```
opsorch-core/
  api/
  registry/
  schema/      <-- evolving, not finalized
  secret/
  runtime/

opsorch-adapter-<provider>/
  incident/
  alert/
  log/
  metric/
  ticket/
  messaging/
```

- **opsorch-core**: routing, secret management, registry, HTTP layer.
- **adapter repos**: implement the capability interfaces.

Adapters do not live inside opsorch-core. They are loaded at runtime either via registered constructors or local plugin binaries.

---

## 2. Capability Interfaces (Shape Agnostic)

OpsOrch defines **interfaces only**, not the schema contents.

Example (simplified):

```go
type IncidentProvider interface {
    Query(ctx context.Context, query schema.IncidentQuery) ([]schema.Incident, error)
    Get(ctx context.Context, id string) (schema.Incident, error)
    Create(ctx context.Context, in schema.CreateIncidentInput) (schema.Incident, error)
    Update(ctx context.Context, id string, in schema.UpdateIncidentInput) (schema.Incident, error)

    GetTimeline(ctx context.Context, id string) ([]schema.TimelineEntry, error)
    AppendTimeline(ctx context.Context, id string, entry schema.TimelineAppendInput) error
}
```

**Important:**  
The *contents* of `schema.Incident`, `schema.TimelineEntry`, etc. are intentionally not fixed here.  
They will be defined as we implement the system.

**Adapters must align with whatever schema version opsorch-core currently exposes.**

---

## 3. Adapter Structure

Each adapter:

- lives in its own repo  
- implements one or more capability interfaces  
- exposes a constructor, usually named `New(config map[string]any)`  
- receives decrypted config from opsorch-core  
- returns normalized objects matching current schemas  
- never stores state

Example skeleton:

```go
func New(config map[string]any) (incident.Provider, error) {
    token := config["apiKey"].(string)
    base := config["baseUrl"].(string)

    return &MyProvider{Token: token, BaseURL: base}, nil
}
```

The provider shape is up to you.

---

## 4. Config + Secrets

OpsOrch Core handles:

- storage
- encryption
- decryption
- validation

Adapters receive **only a decrypted config map**, never raw secrets or tokens.

You **must not log**, print, or expose config values.

---

## 5. Registration

Adapters register themselves in external repos.

OpsOrch Core uses a registry to match capability + provider:

```go
incident.RegisterProvider("pagerduty", pagerduty.New)
incident.RegisterProvider("incidentio", incidentio.New)
alert.RegisterProvider("prometheus", prometheus.New)
```

Providers can be:

- built-in for OSS
- dynamically loaded
- injected through config

### Plugin-Based Loading (no remote services)

OpsOrch Core can launch a local adapter binary as a child process (no network hops) when `OPSORCH_<CAP>_PLUGIN` is set or the provider config includes a `plugin` path. RPC is JSON over stdin/stdout:

- Request: `{ "method": "<capability>.<operation>", "config": {...}, "payload": {...} }`
- Response: `{ "result": <value>, "error": "<msg>" }`

**RPC Methods by Capability:**

| Capability | Method | Payload |
|------------|--------|---------|
| incident | `incident.query` | `IncidentQuery` |
| incident | `incident.list` | `null` |
| incident | `incident.get` | `{ "id": string }` |
| incident | `incident.create` | `CreateIncidentInput` |
| incident | `incident.update` | `{ "id": string, "input": UpdateIncidentInput }` |
| incident | `incident.timeline.get` | `{ "id": string }` |
| incident | `incident.timeline.append` | `{ "id": string, "entry": TimelineAppendInput }` |
| alert | `alert.query` | `AlertQuery` |
| alert | `alert.get` | `{ "id": string }` |
| log | `log.query` | `LogQuery` |
| metric | `metric.query` | `MetricQuery` |
| metric | `metric.describe` | `QueryScope` |
| ticket | `ticket.query` | `TicketQuery` |
| ticket | `ticket.get` | `{ "id": string }` |
| ticket | `ticket.create` | `CreateTicketInput` |
| ticket | `ticket.update` | `{ "id": string, "input": UpdateTicketInput }` |
| messaging | `messaging.send` | `Message` |
| service | `service.query` | `ServiceQuery` |
| deployment | `deployment.query` | `DeploymentQuery` |
| deployment | `deployment.get` | `{ "id": string }` |
| team | `team.query` | `TeamQuery` |
| team | `team.get` | `{ "id": string }` |
| team | `team.members` | `{ "teamID": string }` |
| orchestration | `orchestration.plans.query` | `OrchestrationPlanQuery` |
| orchestration | `orchestration.plans.get` | `{ "planId": string }` |
| orchestration | `orchestration.runs.query` | `OrchestrationRunQuery` |
| orchestration | `orchestration.runs.get` | `{ "runId": string }` |
| orchestration | `orchestration.runs.start` | `{ "planId": string }` |
| orchestration | `orchestration.runs.steps.complete` | `{ "runId": string, "stepId": string, "actor": string, "note": string }` |

| secret | `secret.get` | `{ "key": string }` |
| secret | `secret.put` | `{ "key": string, "value": string }` |

The plugin process stays alive and receives multiple RPC calls on the same stdio stream.

### Building Plugins

To build a plugin, simply build the Go binary:

```bash
go build -o my-plugin ./cmd/my-plugin
```

### Using Plugins

You can configure OpsOrch Core to use a plugin by setting the environment variable:

```bash
export OPSORCH_ALERT_PLUGIN=/path/to/my-plugin
```

Or by specifying the `plugin` path in the provider configuration.

---

## 6. Normalization Responsibilities

Adapters must normalize provider data into the **current schema version** of:

- Incident
- Alert
- TimelineEntry
- LogEntry
- MetricSeries
- Ticket
- MessageResult

### Shared Query Scope

Every query struct includes `schema.QueryScope` so providers can accept common filters without inventing per-provider shapes. Fields are optional, and providers may ignore ones they do not support.

- `Service`: canonical OpsOrch service ID; map to service tags, project IDs, components, etc.
- `Team`: owning team; map to escalation policies, components, tags, or labels.
- `Environment`: coarse environment such as `prod`, `staging`, `dev`; map to provider env labels.

Adapters should translate any recognized scope fields into their native query primitives (tags/labels/filters) and document which fields they honor.

**Schema shapes will change during development.**  
Adapters must track these changes.

Adapters must never return provider-specific fields except in:

```go
metadata map[string]any
```

This preserves backend agnosticism.

---

## 7. Error Handling

Adapters must:

- return typed errors (OpsOrchError)
- never panic
- wrap provider errors
- avoid leaking raw API responses
- avoid exposing secrets in error messages

---

## 8. How to Create a New Adapter

1. Create repo `opsorch-<provider>-adapter`
2. Add dependency:  
   ```
   go get github.com/opsorch/opsorch-core
   ```
3. Implement capability interfaces relevant to the provider  
4. Map provider responses → current OpsOrch schemas  
5. Add unit tests  
6. Add provider usage docs  
7. Ensure the adapter registers itself with the right provider name  

---

## 9. Testing Requirements

Each adapter must include:

- schema conformity tests  
- normalization tests  
- invalid config tests  
- provider error handling tests  
- integration tests using provider mocks  

---

## 10. Version Compatibility

Each adapter must define:

- its own version  
- minimum opsorch-core version supported  

Example:

```go
var AdapterVersion = "0.1.0"
var RequiresCore = ">=0.1.0"
```

---

## 11. Notes on Schema Evolution

OpsOrch Core may evolve:

- incident shape  
- timeline shape  
- log/metric/dash schema  
- custom fields  
- metadata rules  

Adapters must watch release notes and update accordingly.

No schema is locked at this stage.

---

# Summary

This guide explains:

- how adapters attach to OpsOrch Core  
- how they receive config  
- how they normalize responses  
- how they register themselves  
- how they evolve alongside schema changes  

It **intentionally avoids prescribing model shapes**, because those will be defined as the system grows.

---

## 12. Adding a New Capability to OpsOrch Core

This section explains how to add a new capability (like `alert`, `dashboard`, etc.) to OpsOrch Core itself.

### Prerequisites

Before adding a new capability, ensure:
- The capability has a clear, distinct purpose from existing capabilities
- You understand the provider pattern (read-only vs read-write)
- You have identified common fields across major providers in this domain

### Step-by-Step Guide

#### 1. Define the Schema

Create schema file: `schema/<capability>.go`

**Example**: For a read-only capability like alerts:

```go
package schema

import "time"

// AlertQuery filters normalized alerts from the active alert provider.
type AlertQuery struct {
    Query      string         `json:"query,omitempty"`
    Statuses   []string       `json:"statuses,omitempty"`
    Severities []string       `json:"severities,omitempty"`
    Scope      QueryScope     `json:"scope,omitempty"`
    Limit      int            `json:"limit,omitempty"`
    Metadata   map[string]any `json:"metadata,omitempty"`
}

// Alert captures the normalized alert shape.
type Alert struct {
    ID          string         `json:"id"`
    Title       string         `json:"title"`
    Description string         `json:"description,omitempty"`
    Status      string         `json:"status"`
    Severity    string         `json:"severity"`
    Service     string         `json:"service,omitempty"`
    CreatedAt   time.Time      `json:"createdAt"`
    UpdatedAt   time.Time      `json:"updatedAt"`
    Fields      map[string]any `json:"fields,omitempty"`
    Metadata    map[string]any `json:"metadata,omitempty"`
}
```

**For read-write capabilities**, add `Create<Capability>Input` and `Update<Capability>Input` structs.

#### 2. Define the Provider Interface

Create provider file: `<capability>/provider.go`

```go
package alert

import (
    "context"
    "github.com/opsorch/opsorch-core/registry"
    "github.com/opsorch/opsorch-core/schema"
)

// Provider defines the capability surface an alert adapter must satisfy.
type Provider interface {
    Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error)
    Get(ctx context.Context, id string) (schema.Alert, error)
    // For read-write: Add Create, Update, Delete as needed
}

// ProviderConstructor builds a Provider instance from decrypted config.
type ProviderConstructor func(config map[string]any) (Provider, error)

var providers = registry.New[ProviderConstructor]()

// RegisterProvider adds a provider constructor.
func RegisterProvider(name string, constructor ProviderConstructor) error {
    return providers.Register(name, constructor)
}

// LookupProvider returns a named provider constructor if registered.
func LookupProvider(name string) (ProviderConstructor, bool) {
    return providers.Get(name)
}

// Providers lists all registered provider names.
func Providers() []string {
    return providers.Names()
}
```

#### 3. Add Provider Tests

Create test file: `<capability>/provider_test.go`

```go
package alert

import (
    "context"
    "testing"
    "github.com/opsorch/opsorch-core/schema"
)

type stubAlertProvider struct{}

func (s stubAlertProvider) Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error) {
    return nil, nil
}

func (s stubAlertProvider) Get(ctx context.Context, id string) (schema.Alert, error) {
    return schema.Alert{}, nil
}

func TestAlertRegisterLookup(t *testing.T) {
    name := "test-alert"
    ctor := func(cfg map[string]any) (Provider, error) { return stubAlertProvider{}, nil }
    if err := RegisterProvider(name, ctor); err != nil && err.Error() != "registry: provider test-alert already registered" {
        t.Fatalf("register: %v", err)
    }
    got, ok := LookupProvider(name)
    if !ok || got == nil {
        t.Fatalf("expected provider lookup success")
    }
    names := Providers()
    found := false
    for _, n := range names {
        if n == name {
            found = true
            break
        }
    }
    if !found {
        t.Fatalf("expected provider name in list: %v", names)
    }
}

func TestAlertDuplicateFails(t *testing.T) {
    name := "dup-alert"
    ctor := func(cfg map[string]any) (Provider, error) { return stubAlertProvider{}, nil }
    _ = RegisterProvider(name, ctor)
    if err := RegisterProvider(name, ctor); err == nil {
        t.Fatalf("expected duplicate registration to fail")
    }
}
```

#### 4. Create API Handler

Create handler file: `api/<capability>_handler.go`

```go
package api

import (
    "fmt"
    "net/http"
    "strings"
    "github.com/opsorch/opsorch-core/alert"
    "github.com/opsorch/opsorch-core/orcherr"
    "github.com/opsorch/opsorch-core/schema"
)

// AlertHandler wraps provider wiring for alerts.
type AlertHandler struct {
    provider alert.Provider
}

func newAlertHandlerFromEnv(sec SecretProvider) (AlertHandler, error) {
    name, cfg, pluginPath, err := loadProviderConfig(sec, "alert", "OPSORCH_ALERT_PROVIDER", "OPSORCH_ALERT_CONFIG", "OPSORCH_ALERT_PLUGIN")
    if err != nil || (name == "" && pluginPath == "") {
        return AlertHandler{}, err
    }
    if pluginPath != "" {
        return AlertHandler{}, fmt.Errorf("alert plugins not yet supported")
    }
    constructor, ok := alert.LookupProvider(name)
    if !ok {
        return AlertHandler{}, fmt.Errorf("alert provider %s not registered", name)
    }
    provider, err := constructor(cfg)
    if err != nil {
        return AlertHandler{}, err
    }
    return AlertHandler{provider: provider}, nil
}

func (s *Server) handleAlert(w http.ResponseWriter, r *http.Request) bool {
    if !strings.HasPrefix(r.URL.Path, "/alerts") {
        return false
    }
    if s.alert.provider == nil {
        writeError(w, http.StatusNotImplemented, orcherr.OpsOrchError{Code: "alert_provider_missing", Message: "alert provider not configured"})
        return true
    }

    path := strings.TrimSuffix(r.URL.Path, "/")
    segments := strings.Split(strings.Trim(path, "/"), "/")

    switch {
    case len(segments) == 2 && segments[1] == "query" && r.Method == http.MethodPost:
        var query schema.AlertQuery
        if err := decodeJSON(r, &query); err != nil {
            writeError(w, http.StatusBadRequest, orcherr.OpsOrchError{Code: "bad_request", Message: err.Error()})
            return true
        }
        alerts, err := s.alert.provider.Query(r.Context(), query)
        if err != nil {
            writeProviderError(w, err)
            return true
        }
        logAudit(r, "alert.query")
        writeJSON(w, http.StatusOK, alerts)
        return true
    case len(segments) == 2 && r.Method == http.MethodGet:
        id := segments[1]
        al, err := s.alert.provider.Get(r.Context(), id)
        if err != nil {
            writeProviderError(w, err)
            return true
        }
        logAudit(r, "alert.get")
        writeJSON(w, http.StatusOK, al)
        return true
    default:
        return false
    }
}
```

#### 5. Wire Up the Server

**Modify** `api/server.go`:

1. Add handler field to `Server` struct:
```go
type Server struct {
    // ... existing fields
    alert AlertHandler
    // ... existing fields
}
```

2. Initialize handler in `NewServerFromEnv`:
```go
al, err := newAlertHandlerFromEnv(sec)
if err != nil {
    return nil, err
}
```

3. Set handler in `Server` struct initialization:
```go
return &Server{
    // ... existing fields
    alert: al,
    // ... existing fields
}, nil
```

4. Add dispatch in `ServeHTTP`:
```go
case s.handleAlert(w, r):
```

#### 6. Implement Plugin Provider

**Modify** `api/plugin_providers.go`:

1. Define the plugin provider struct:
```go
type alertPluginProvider struct {
    runner *pluginRunner
}

func newAlertPluginProvider(path string, cfg map[string]any) alertPluginProvider {
    return alertPluginProvider{runner: newPluginRunner(path, cfg)}
}
```

2. Implement the provider interface methods, delegating to `runner.call`:
```go
func (p alertPluginProvider) Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error) {
    var res []schema.Alert
    return res, p.runner.call(ctx, "alert.query", query, &res)
}

func (p alertPluginProvider) Get(ctx context.Context, id string) (schema.Alert, error) {
    var res schema.Alert
    return res, p.runner.call(ctx, "alert.get", map[string]any{"id": id}, &res)
}
```

#### 7. Update Capability Registry

**Modify** `api/capability.go` - Add capability normalization:

```go
func normalizeCapability(name string) (string, bool) {
    switch strings.ToLower(name) {
    // ... existing cases
    case "alert", "alerts":
        return "alert", true
    // ... existing cases
    }
}
```

**Modify** `api/providers.go` - Add provider listing:

```go
import (
    // ... existing imports
    "github.com/opsorch/opsorch-core/alert"
)

func (s *Server) handleProviders(w http.ResponseWriter, r *http.Request) bool {
    // ... existing code
    switch capability {
    // ... existing cases
    case "alert":
        providers = alert.Providers()
    // ... existing cases
    }
}
```

#### 8. Add API Tests

**Modify** `api/server_test.go`:

1. Add import:
```go
import (
    // ... existing imports
    "github.com/opsorch/opsorch-core/alert"
)
```

2. Add stub provider:
```go
type stubAlertProvider struct{}

func (s stubAlertProvider) Query(ctx context.Context, query schema.AlertQuery) ([]schema.Alert, error) {
    res := []schema.Alert{{ID: "a1", Title: "test alert", Status: "firing", Severity: "critical", CreatedAt: time.Now(), UpdatedAt: time.Now()}}
    if query.Limit > 0 && query.Limit < len(res) {
        return res[:query.Limit], nil
    }
    return res, nil
}

func (s stubAlertProvider) Get(ctx context.Context, id string) (schema.Alert, error) {
    return schema.Alert{ID: id, Title: "test alert", Status: "firing", Severity: "critical", CreatedAt: time.Now(), UpdatedAt: time.Now()}, nil
}
```

3. Add tests:
```go
func TestAlertQuery(t *testing.T) {
    srv := &Server{alert: AlertHandler{provider: stubAlertProvider{}}, corsOrigin: "*"}
    body, _ := json.Marshal(schema.AlertQuery{Limit: 1})
    req := httptest.NewRequest(http.MethodPost, "/alerts/query", bytes.NewReader(body))
    w := httptest.NewRecorder()
    srv.ServeHTTP(w, req)
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", res.StatusCode)
    }
    var out []schema.Alert
    if err := json.NewDecoder(res.Body).Decode(&out); err != nil {
        t.Fatalf("decode response: %v", err)
    }
    if len(out) != 1 || out[0].ID != "a1" {
        t.Fatalf("unexpected alert response: %+v", out)
    }
}

func TestAlertGet(t *testing.T) {
    srv := &Server{alert: AlertHandler{provider: stubAlertProvider{}}, corsOrigin: "*"}
    req := httptest.NewRequest(http.MethodGet, "/alerts/abc", nil)
    w := httptest.NewRecorder()
    srv.ServeHTTP(w, req)
    
    res := w.Result()
    if res.StatusCode != http.StatusOK {
        t.Fatalf("expected 200, got %d", res.StatusCode)
    }
}
```

4. Update `TestMissingProvidersReturn501`:
```go
cases := []struct {
    name   string
    server *Server
    method string
    url    string
}{
    {"alert", &Server{alert: AlertHandler{}}, http.MethodPost, "/alerts/query"},
    // ... existing cases
}
```

#### 9. Update Documentation

**Modify** `README.md`:

1. Add to environment variables list:
```markdown
Environment variables for any capability (`incident`, `alert`, `log`, ...):
```

2. Add to API endpoints list:
```markdown
- Incidents
- Alerts
- Logs
```

3. Add usage example:
```bash
# Query Alerts
curl -s -X POST http://localhost:8080/alerts/query -d '{}'
```

### Verification Checklist

- [ ] Schema defined in `schema/<capability>.go`
- [ ] Provider interface defined in `<capability>/provider.go`
- [ ] Provider tests in `<capability>/provider_test.go`
- [ ] API handler in `api/<capability>_handler.go`
- [ ] Server wired up in `api/server.go`
- [ ] Plugin provider in `api/plugin_providers.go`
- [ ] Capability normalized in `api/capability.go`
- [ ] Provider listing in `api/providers.go`
- [ ] API tests in `api/server_test.go`
- [ ] Documentation updated in `README.md`
- [ ] All tests passing: `go test ./...`

---

# Summary
