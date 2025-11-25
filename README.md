# OpsOrch Core

OpsOrch Core is a stateless, open-source orchestration layer that unifies incident, log, metric, dashboard, ticket, and messaging workflows behind a single, provider-agnostic API. 
It does not store operational data, and it does not include any built-in vendor integrations.  
External adapters implement provider logic and are loaded dynamically by OpsOrch Core.

OpsOrch Core provides:
- Unified API surfaces
- Secret management
- Capability registry
- Schema boundaries (evolving)
- Routing and request orchestration

Adapters live in separate repos such as:
- opsorch-pagerduty-adapter
- opsorch-incidentio-adapter
- opsorch-prometheus-adapter
- opsorch-elasticsearch-adapter
- opsorch-grafana-adapter
- opsorch-jira-adapter
- opsorch-slack-adapter

## Adapter Loading Model

OpsOrch Core never links vendor logic directly. Each capability is wired via an **in-process provider** that you import into the binary. The provider registers itself (e.g., `incident.RegisterProvider("pagerduty", pagerduty.New)`) and is selected with env `OPSORCH_<CAP>_PROVIDER`.

Environment variables for any capability (`incident`, `log`, `metric`, `dashboard`, `ticket`, `messaging`, `secret`):
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
- Timelines
- Logs
- Metrics
- Dashboards
- Tickets
- Messaging

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
OpsOrch loads integration credentials through the secret provider interface. Select a provider via `OPSORCH_SECRET_PROVIDER=<name>` and pass any constructor options in `OPSORCH_SECRET_CONFIG=<json>`.

Supported:
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

## Creating an Adapter

See `agent.md` for the full guide.

Summary:
1. Create a repo `opsorch-adapter-<provider>`
2. `go get github.com/opsorch/opsorch-core`
3. Implement capability interfaces
4. Export constructor `New(config map[string]any)`
5. Register provider
6. Add tests

## License
Apache 2.0
