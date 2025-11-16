# OpsOrch Core

**OpsOrch Core is a stateless, open-source orchestration layer that unifies incident, log, metric, dashboard, ticket, and messaging workflows behind a single, provider-agnostic API.**  
It does not store operational data, and it does not include any built-in vendor integrations.  
External adapters implement provider logic and are loaded dynamically by OpsOrch Core.

OpsOrch Core provides:
- Unified API surfaces
- Secret management (Vault, KMS, AES-256 for dev)
- Capability registry
- Schema boundaries (evolving)
- Routing and request orchestration

Adapters live in separate repos such as:
- opsorch-adapter-pagerduty
- opsorch-adapter-incidentio
- opsorch-adapter-prometheus
- opsorch-adapter-elasticsearch
- opsorch-adapter-grafana
- opsorch-adapter-jira
- opsorch-adapter-slack

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

### Adapter Architecture
OpsOrch Core contains **no provider logic**.  
Adapters implement capability interfaces in their own repos and register with the registry.

### No Data Storage
OpsOrch Core does not store operational data such as incidents or logs.  
It stores only:
- encrypted integration configs  
- minimal integration metadata  
- optional audit logs  

### Secure Secret Management
Supported:
- HashiCorp Vault
- AWS KMS
- GCP/Azure KMS
- Local AES-256-GCM

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
