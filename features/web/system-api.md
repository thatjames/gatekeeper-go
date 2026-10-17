---
type: Feature
title: System API
description: /api/v1/version and authenticated /api/v1/system routes for host info, network interfaces and active modules; /features names extra features such as oidc.
status: stable
code:
  - internal/web/v1/system.go
tags: [web, api, system]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- `GET /version` → `{version}`.
- `GET /features` → `["oidc"]` or `[]`; 500 (recovered panic) when no
  `Auth:` is configured.
- `GET /system/info` → [system info](/system/system-info.md) map.
- `GET /system/interfaces` → names of up, non-loopback, non-point-to-point
  interfaces with a MAC, minus docker, bridge, veth, virbr and Hamachi ones.
- `GET /system/modules` → the running module keys.
- `/health` → `{status: "healthy", timestamp}`.

# Decisions
- The `/features` panic is fixed by the UI auth flow plan.
