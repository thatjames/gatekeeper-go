---
type: Feature
title: Unused OIDC config helper
description: deprecated - NewOIDCConfig and OIDCOptions in auth.go duplicate initOIDC and are never called; to be deleted.
status: deprecated
code:
  - internal/web/v1/auth.go
tags: [web, auth, dead-code]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
None left: confirmed dead by the user on 2026-10-17.

# Behavior
- `NewOIDCConfig(ctx, OIDCOptions)` builds a provider, verifier and OAuth2
  config; nothing calls it. `initOIDC` in `api.go` does the real work.

# Decisions
- Removed by the dead code plan.
