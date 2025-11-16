# OpsOrch Adapter Development Guide

This document explains how to build **external provider adapters** for OpsOrch Core.

OpsOrch Core is a **stateless orchestration layer**.  
It provides unified APIs, routing, secret management, and schema boundaries — but **does not define or enforce the exact shape of incidents, logs, metrics, dashboards, tickets, or messages**.

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
  log/
  metric/
  dashboard/
  ticket/
  messaging/
```

- **opsorch-core**: routing, secret management, registry, HTTP layer.
- **adapter repos**: implement the capability interfaces.

Adapters do not live inside opsorch-core.

---

## 2. Capability Interfaces (Shape Agnostic)

OpsOrch defines **interfaces only**, not the schema contents.

Example (simplified):

```go
type IncidentProvider interface {
    List(ctx context.Context) ([]schema.Incident, error)
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
```

Providers can be:

- built-in for OSS
- dynamically loaded
- injected through config

---

## 6. Normalization Responsibilities

Adapters must normalize provider data into the **current schema version** of:

- Incident
- TimelineEntry
- LogEntry
- MetricSeries
- DashboardView
- Ticket
- MessageResult

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

1. Create repo `opsorch-adapter-<provider>`
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

