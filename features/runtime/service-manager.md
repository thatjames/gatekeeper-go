---
type: Feature
title: Service manager
description: A process-wide registry of the running modules (dhcp, dns, web keys) that starts and stops them and lets the web API fetch a module by key.
status: stable
code:
  - internal/service/**
tags: [runtime]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Lets the API reach the live DHCP and DNS servers and tell the UI which
modules and extra features are on.

# Behavior
- `Register(service, key)` stores a service under `dhcp`, `dns` or `web`;
  a second register under the same key replaces the first.
- `Start()` / `Stop()` call each service in turn and return the first
  error; later services are not started or stopped after an error.
- `GetService[T](key)` returns the service cast to `T`; an unregistered
  key or wrong type panics.
- `IsRegistered(key)`, `ActiveServices()` (the keys, unordered).
- `ExtraFeatures()` returns `["oidc"]` when the auth type is `oidc`,
  otherwise `[]`. When no `Auth:` is configured it panics (nil interface):
  a known bug, see [UI auth flow](/ui/app-shell.md).
- `Register` takes no lock; registration only happens before the web
  server starts.

# Interfaces
- Used by [startup](/runtime/startup.md) and the [web API](/web/api-routes.md).
