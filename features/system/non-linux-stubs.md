---
type: Feature
title: Non-Linux stubs
description: deprecated - darwin and windows stubs of the system package that return "not supported"; gatekeeper is Linux only, the darwin build fails, to be deleted.
status: deprecated
code:
  - internal/system/system_darwin.go
  - internal/system/system_windows.go
tags: [system, dead-code]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
None left: Linux only, confirmed 2026-10-17.

# Behavior
- darwin lacks `GetNetworkInterfaces`, so `GOOS=darwin go build` fails.
- windows returns different types than Linux.
- `GetDHCPInterfaces` exists only here and is never called.

# Decisions
- Removed by the dead code plan, together with any build tags they need.
