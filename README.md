# OpsOrch Core

OpsOrch Core is a stateless, open-source orchestration layer that unifies incident, log, metric, ticket, and messaging workflows behind a single, provider-agnostic API. 
It does not store operational data, and it does not include any built-in vendor integrations.  
External adapters implement provider logic and are loaded dynamically by OpsOrch Core.

OpsOrch Core provides:
- Unified API surfaces
- Secret management
- Capability registry
- Schema boundaries (evolving)
- Routing and request orchestration

Adapters live in separate repos such as:
- [`opsorch-pagerduty-adapter`](https://github.com/OpsOrch/opsorch-pagerduty-adapter) - PagerDuty incidents and services
- [`opsorch-jira-adapter`](https://github.com/OpsOrch/opsorch-jira-adapter) - Jira ticket management
- [`opsorch-prometheus-adapter`](https://github.com/OpsOrch/opsorch-prometheus-adapter) - Prometheus metrics
- [`opsorch-mock-adapters`](https://github.com/OpsOrch/opsorch-mock-adapters) - Demo/testing providers for all capabilities
- [`opsorch-adapter`](https://github.com/OpsOrch/opsorch-adapter) - Starter template for new adapters

## Adapter Loading Model

OpsOrch Core never links vendor logic directly. Each capability is wired via an **in-process provider** that you import into the binary. The provider registers itself (e.g., `incident.RegisterProvider("pagerduty", pagerduty.New)`) and is selected with env `OPSORCH_<CAP>_PROVIDER`.

Environment variables for any capability (`incident`, `alert`, `log`, `metric`, `ticket`, `messaging`, `service`, `secret`):
- `OPSORCH_<CAP>_PROVIDER=<registered name>`
- `OPSORCH_<CAP>_CONFIG=<json>`

### Using an in-process provider
1) Add the adapter dependency to the core binary you are building:
```bash
go get github.com/opsorch/opsorch-adapter-pagerduty
```
2) Import the adapter for side effects so it registers itself (create `cmd/opsorch/providers.go` if you prefer to keep imports separate):
```go
package main

import (
    _ "github.com/opsorch/opsorch-adapter-pagerduty/incident" // registers with incident registry
    _ "github.com/opsorch/opsorch-adapter-elasticsearch/log"  // registers with log registry
)
```
3) Select the provider via env and pass its config:
```bash
OPSORCH_INCIDENT_PROVIDER=pagerduty OPSORCH_INCIDENT_CONFIG='{"apiKey":"...","routingKey":"..."}' \
OPSORCH_LOG_PROVIDER=elasticsearch OPSORCH_LOG_CONFIG='{"url":"http://..."}' \
go run ./cmd/opsorch
```

### Quick start: run locally and curl

Run the server (defaults to :8080) with a registered provider and its config:
```bash
OPSORCH_INCIDENT_PROVIDER=<registered> OPSORCH_INCIDENT_CONFIG='{"token":"..."}' go run ./cmd/opsorch
```

Hit the API:
```bash
curl -s -X POST http://localhost:8080/incidents/query -d '{}'
curl -s -X POST http://localhost:8080/incidents \
  -H "Content-Type: application/json" \
  -d '{"title":"test","status":"open","severity":"sev3"}'
  -d '{"title":"test","status":"open","severity":"sev3"}'

# Query Alerts
curl -s -X POST http://localhost:8080/alerts/query -d '{}'

# Query Metrics
curl -s -X POST http://localhost:8080/metrics/query \
  -H "Content-Type: application/json" \
  -d '{
    "expression": {
      "metricName": "http_requests_total",
      "aggregation": "sum"
    },
    "start": "2023-10-01T00:00:00Z",
    "end": "2023-10-01T01:00:00Z",
    "step": 60
  }'

# Discover Metrics
curl -s "http://localhost:8080/metrics/describe?service=api"
```

### TLS

To terminate HTTPS directly in opsorch-core, set both env vars:

```bash
OPSORCH_TLS_CERT_FILE=/path/to/server.crt OPSORCH_TLS_KEY_FILE=/path/to/server.key go run ./cmd/opsorch
```

If only one is provided the server will refuse to start.

### Docker image

#### Using Published Images

Pre-built multi-platform Docker images (linux/amd64, linux/arm64) are automatically published to GitHub Container Registry (GHCR) on every release.

**Pull and run the latest version:**
```bash
docker pull ghcr.io/opsorch/opsorch-core:latest
docker run --rm -p 8080:8080 ghcr.io/opsorch/opsorch-core:latest
```

**Pull a specific version:**
```bash
docker pull ghcr.io/opsorch/opsorch-core:v0.1.0
docker run --rm -p 8080:8080 ghcr.io/opsorch/opsorch-core:v0.1.0
```

Published images contain only the core binary (no bundled plugins). Load adapters via in-process providers or external plugin binaries as documented above.

#### Creating a Release

Releases are automated via GitHub Actions. To create a new release:

1. Go to the [Actions tab](https://github.com/OpsOrch/opsorch-core/actions)
2. Select the "Release" workflow
3. Click "Run workflow"
4. Choose the version bump type:
   - **patch** (default): Bug fixes, minor changes (e.g., `v0.1.0` → `v0.1.1`)
   - **minor**: New features, backward compatible (e.g., `v0.1.0` → `v0.2.0`)
   - **major**: Breaking changes (e.g., `v0.1.0` → `v1.0.0`)
5. Click "Run workflow"

The workflow will:
- Run all tests and linting checks
- Automatically calculate and create the new version tag
- Build and push multi-platform Docker images to GHCR
- Create a GitHub Release with changelog

#### Building Locally

A Dockerfile is provided that builds the core binary and bundles the mock adapter plugins at `/opt/opsorch/plugins`. You can also build a core-only base image and layer plugins later.

Build an image (override `IMAGE` to change the tag; default is `opsorch-core:latest`).
The `PLUGINS` build arg controls which plugin directories under `./plugins` are built and bundled (defaults to `incidentmock logmock secretmock`).

```bash
make docker-build IMAGE=opsorch-core:dev PLUGINS="incidentmock logmock secretmock"
```

Build a core-only base image (no plugins included):

```bash
make docker-build-base BASE_IMAGE=opsorch-core-base:dev
```

Run with the packaged plugin binaries:

```bash
docker run --rm -p 8080:8080 \
  -e OPSORCH_INCIDENT_PLUGIN=/opt/opsorch/plugins/incidentmock \
  -e OPSORCH_LOG_PLUGIN=/opt/opsorch/plugins/logmock \
  -e OPSORCH_SECRET_PLUGIN=/opt/opsorch/plugins/secretmock \
  opsorch-core:dev
```

If you built without overriding `IMAGE`, run with `opsorch-core:latest` instead of `opsorch-core:dev`.

Mount or copy additional adapter binaries and point `OPSORCH_<CAP>_PLUGIN` env vars at them to swap providers.

To bundle your own plugins, add their source under `plugins/<name>` (or vendor them in), then include them in `PLUGINS` when building: `make docker-build PLUGINS="incidentmock logmock secretmock myprovider"`.

## Key Concepts

### Unified API Layer
OpsOrch exposes API endpoints for:
- Incidents
- Alerts
- Timelines
- Logs
- Metrics
- Tickets
- Messaging
- Services

Schemas live under `schema/` and evolve as the system matures.

### Shared Query Scope
All query payloads accept `schema.QueryScope`, a minimal set of filters adapters can map into their native query languages. Fields are optional and may be ignored by providers:
- `service`: canonical OpsOrch service ID (map to tags, project IDs, components)
- `team`: owner/team (map to escalation policies, components, tags)
- `environment`: coarse env such as `prod`, `staging`, `dev` (map to env labels)

### Structured Queries
OpsOrch uses structured expressions for querying logs and metrics, replacing free-form strings to ensure validation and consistency.

**Metric Queries**:
- **Structure**: `MetricName`, `Aggregation` (sum, avg, etc.), `Filters` (label-based), `GroupBy`.
- **Discovery**: Use `GET /metrics/describe` to find available metrics and their labels.
- **Example**:
  ```json
  {
    "expression": {
      "metricName": "http_requests_total",
      "aggregation": "sum",
      "filters": [{"label": "status", "operator": "=", "value": "500"}]
    }
  }
  ```

**Log Queries**:
- **Structure**: `Search` (text), `Filters` (field-based), `SeverityIn`.
- **Example**:
  ```json
  {
    "expression": {
      "search": "connection timeout",
      "severityIn": ["error", "critical"]
    }
  }
  ```

### Adapter Architecture
OpsOrch Core contains **no provider logic**.  
Adapters implement capability interfaces in their own repos and register with the registry.

### No Data Storage
OpsOrch Core does not store operational data such as incidents or logs.  
It stores only:
- encrypted integration configs  
- minimal integration metadata  
- optional audit logs (structured JSON with actions like `incident.created`, `incident.query`)  

### Secure Secret Management
OpsOrch loads integration credentials through the secret provider interface.

#### Secret Provider Loading Priority
The secret provider is loaded in the following order of precedence:
1. **Plugin mode**: If `OPSORCH_SECRET_PLUGIN` is set, OpsOrch spawns that plugin binary (highest priority)
2. **In-process provider**: If `OPSORCH_SECRET_PROVIDER` is set, OpsOrch uses the registered provider by that name
3. **No secret provider**: If neither is set, the system runs without secret management (providers must be configured via environment variables only)

Configuration is always passed via `OPSORCH_SECRET_CONFIG=<json>`.

Supported providers:
- HashiCorp Vault
- AWS KMS
- GCP/Azure KMS
- Local AES-256-GCM
- JSON file store (built in, local/dev convenience)

#### JSON file provider (built in)
The repo ships with a simple JSON-backed provider that is handy for demos and local development. Point the secret subsystem at a file that contains logical keys such as `providers/<capability>/default` and raw JSON strings for the stored configs:

```json
{
  "providers/incident/default": "{\"provider\":\"incidentmock\",\"config\":{\"token\":\"abc\"}}",
  "providers/log/default": "{\"provider\":\"logmock\",\"config\":{\"url\":\"http://localhost:9200\"}}"
}
```

Start OpsOrch with environment variables pointing at that file:

```bash
OPSORCH_SECRET_PROVIDER=json OPSORCH_SECRET_CONFIG='{"path":"/tmp/opsorch-secrets.json"}' go run ./cmd/opsorch
```

The JSON provider keeps changes in memory only; edit the file yourself (or rebuild it) if you want to persist new configs across restarts.

#### Applying capability configs
Each capability can be configured in two ways:
- **Environment variables at startup**: supply `OPSORCH_<CAP>_PROVIDER` and `OPSORCH_<CAP>_CONFIG` (and optionally `OPSORCH_<CAP>_PLUGIN`) every time you launch the server.
- **Persisted configs via the secret store**: once a secret provider (such as the JSON file provider) is set, POST `{"provider":"name","config":{...}}` to `/providers/<capability>` and OpsOrch will persist that payload under the logical key `providers/<capability>/default`. Future restarts automatically reload the stored values, so setting the env vars again is optional.

OpsOrch never returns secrets or logs them.

## Architecture Overview

```
                 +------------------------+
                 |     OpsOrch Core       |
                 |  (routing + schemas)   |
                 +-----------+------------+
                             |
                   Capability Registry
                             |
           -------------------------------------
           |                 |                 |
     Incident Adapter   Log Adapter      Metric Adapter
  (external repo)     (external repo)   (external repo)
           |                 |                 |
           +-----------------+-----------------+
                             |
                   External Providers
```

## Installation

```bash
git clone https://github.com/opsorch/opsorch-core
cd opsorch-core
make build
./opsorch
```

## Configuration

```
SECRET_BACKEND=vault|kms|local
VAULT_ADDR=http://127.0.0.1:8200
VAULT_TOKEN=xxxx
VAULT_TRANSIT_KEY=opsorch
```

## Extending OpsOrch

### 1. Creating a New Adapter

To build a new adapter for an existing capability (e.g., a new Incident provider):

1. Create a repo `opsorch-adapter-<provider>`
2. Add dependency:
   ```bash
   go get github.com/opsorch/opsorch-core
   ```
3. Implement capability interfaces relevant to the provider
4. Map provider responses → current OpsOrch schemas
5. Add unit tests
6. Add provider usage docs
7. Ensure the adapter registers itself with the right provider name

See [AGENTS.md](AGENTS.md) for detailed interface definitions and normalization rules.

### 2. Adding a New Capability to OpsOrch Core

This section explains how to add a new capability (like `alert`, `dashboard`, etc.) to OpsOrch Core itself.

#### Prerequisites
- The capability has a clear, distinct purpose from existing capabilities
- You understand the provider pattern (read-only vs read-write)
- You have identified common fields across major providers in this domain

#### Step-by-Step Guide

**1. Define the Schema**

Create schema file: `schema/<capability>.go`

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

**2. Define the Provider Interface**

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

**3. Create API Handler**

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
        return AlertHandler{provider: newAlertPluginProvider(pluginPath, cfg)}, nil
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
    // ... implementation ...
}
```

**4. Wire Up the Server**

Modify `api/server.go` to add the handler field, initialize it in `NewServerFromEnv`, and dispatch in `ServeHTTP`.

**5. Implement Plugin Provider**

Modify `api/plugin_providers.go`:

1. Define the plugin provider struct:
```go
type alertPluginProvider struct {
    runner *pluginRunner
}

func newAlertPluginProvider(path string, cfg map[string]any) alertPluginProvider {
    return alertPluginProvider{runner: newPluginRunner(path, cfg)}
}
```

2. Implement the provider interface methods, delegating to `runner.call`.

**6. Update Capability Registry**

Modify `api/capability.go` to add capability normalization, and `api/providers.go` to add provider listing.

**7. Add API Tests**

Add tests in `api/server_test.go`.

**8. Update Documentation**

Update `README.md` to list the new capability and endpoints.

## License
Apache 2.0
