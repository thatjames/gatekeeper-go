---
type: Feature
title: HTTP server
description: A gin server on Web.Address (default :8085, optional TLS) that serves the API under /api, authenticated Prometheus metrics at /metrics when Web.Prometheus is true, and the embedded UI for every other path, with CORS for WebURL and Web.AllowedOrigins.
status: stable
code:
  - internal/web/web.go
  - internal/web/embedfs.go
  - internal/web/web_test.go
tags: [web, http, metrics]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T21:59:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
One port for the admin UI, its API and the metrics.

# Behavior
- `Address` defaults to `:8085`, `WebURL` to `http://localhost:5173`.
- With `TLS` set it serves HTTPS from the PublicKey/PrivateKey files, else
  plain HTTP.
- CORS allows `WebURL` plus each origin in `Web.AllowedOrigins` (empty by
  default), with credentials and the Authorization header. A request from
  any other origin gets no CORS headers. The embedded UI is same-origin and
  needs none; `AllowedOrigins` is for a UI served elsewhere, such as the
  Vite dev server on `http://localhost:5173`.
- `/metrics` serves the Prometheus registry only when `Web.Prometheus` is
  true; otherwise it is not registered and answers 404. It is behind the
  [auth middleware](/web/auth.md) (401 without a token, 403 with a bad
  one), because it shows client IPs, MACs, hostnames and queried domains.
- Paths outside `/api/` and `/metrics` with no `.` are rewritten to `/`, so
  the single-page app handles them; files come from `ui/dist` embedded at
  build time (the build fails without `internal/web/ui/dist`).
- Panics in handlers are recovered as 500.

# Interfaces
- [API routes](/web/api-routes.md), [UI shell](/ui/app-shell.md).

# Decisions
- `/metrics` always needs auth and is off unless `Prometheus: true` (user,
  2026-10-18): it leaks LAN detail, the bool's zero value is the documented
  default, and a scraper can send a token.
- CORS from `WebURL` and an optional `AllowedOrigins` list (user,
  2026-10-18): `WebURL` is also the OIDC redirect target and sets the
  cookie's `Secure` flag, so a dev UI cannot just take it over; the
  hardcoded `gatekeeper.thatjames.nl` is gone.
