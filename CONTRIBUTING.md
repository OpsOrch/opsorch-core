# Contributing to OpsOrch Core

Thank you for considering contributing!

OpsOrch Core is intentionally minimal.  
It focuses on routing, schemas, secret management, and capability interfaces.  
All provider integrations live in **external repositories**.

## How You Can Contribute

### 1. Core Contributions
You may contribute to:
- API layer
- Schema definitions
- Registry and loading system
- Secret provider implementations
- Documentation
- Bug fixes

### 2. Adapter Contributions
Adapters must live in separate repositories.  
If you want to contribute an adapter, create a repo named:

```
opsorch-adapter-<provider>
```

Follow the guide in `README.md`.

### 3. Coding Standards
- Go 1.22+
- Follow idiomatic Go practices
- Use interfaces for capability boundaries
- No provider SDKs in core
- No secrets or plaintext logs

### 4. Pull Requests
- Include tests
- Keep changes small and focused
- Update documentation if needed
- Ensure backwards compatibility or note breaking changes

### 5. Reporting Issues
Please include:
- Steps to reproduce
- Expected behavior
- Actual behavior
- Environment details

### 6. License
All contributions must be licensed under Apache 2.0.
