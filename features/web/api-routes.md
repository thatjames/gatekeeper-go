---
type: Feature
title: API v1 routes
description: The /api/v1 route table - open health, version, features and login routes; DHCP, DNS and system routes behind auth, registered only for running modules - with request types and validation.
status: stable
code:
  - internal/web/v1/api.go
  - internal/web/v1/api_test.go
  - internal/web/v1/types.go
  - internal/web/v1/types_test.go
tags: [web, api]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T12:50:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
The contract between the embedded UI and the server.

# Behavior
- Open: `GET /api/health`, `GET /api/v1/health`, `GET /api/v1/version`,
  `GET /api/v1/features`, `POST /api/v1/login`; with OIDC also
  `GET /api/v1/oidc-login`, `GET /api/v1/auth/callback`,
  `POST /api/v1/logout` ([auth](/web/auth.md)).
- Behind the auth middleware and request logging:
  `GET /api/v1/me` ([auth](/web/auth.md)), `/dhcp/*` when DHCP runs
  ([DHCP API](/web/dhcp-api.md)), `/dns/*` when DNS runs
  ([DNS API](/web/dns-api.md)), and `/system/*`
  ([system API](/web/system-api.md)).
- Validation errors answer 400 `{error, fields: [{field, message}]}`; a body
  that does not bind answers 400 `{error: "Invalid request format"}`.
- Validators: lease request (client ID, IPv4); local domain (name, IPv4);
  DHCP options (interface, IPv4 range, mask, gateway, DNS servers, TTL);
  DNS config (interface, at least one IPv4 upstream); blocklist (non-empty
  `url`).
- There is no `/verify` route although the UI calls it; `/me` is its
  replacement (UI auth flow plan).

# Decisions
- The API shape is not a hard constraint (2026-10-17): only the embedded UI
  uses it.
