---
type: Feature
title: Login and tokens
description: Users log in with htpasswd credentials or OIDC (Authentik) and get a 7-day HS256 JWT signed with GATEKEEPER_JWT_SECRET, sent as a Bearer header or the oauth_token cookie.
status: stable
code:
  - internal/web/v1/api.go
  - internal/web/v1/jwt.go
  - internal/web/v1/middleware.go
tags: [web, auth, oidc, jwt]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Only the admin (the user, alone) can change DNS and DHCP. The identity
provider is Authentik on atlas (confirmed 2026-10-17).

# Behavior
- `POST /login {username, password}` checks the `Web.HTPasswdFile` file
  (any htpasswd hash). It logs the whole request, password included (bug).
  Without a readable file it makes a new random 16-character password for
  every request and logs it, so no login can succeed (bug). The route stays
  open when OIDC is configured.
- Success → `{token}`: HS256 JWT, subject and username = login name, role
  `user`, issuer `gatekeeper`, valid 7 days.
- OIDC (`AuthType: oidc`): discovery from `IssuerURL` at startup (on failure
  the OIDC routes are not added). `/oidc-login` sets a random state cookie
  (10 min, Secure, HttpOnly) and redirects to the provider. `/auth/callback`
  checks the state, exchanges the code, verifies the ID token for
  `ClientID`, and issues the app JWT for the token's email in the
  `oauth_token` cookie (1 h, not Secure, not HttpOnly), then redirects to
  `WebURL`. Any user the provider accepts gets in; the `roles` claim is read
  but unused.
- The client secret comes from the variable named by `ClientSecretVar`.
- Middleware: token from `Authorization: Bearer`, else the `oauth_token`
  cookie. None → 401. Invalid or expired → 403 and the cookie is cleared.
- `/logout` clears the `oauth_token` cookie.
- `GATEKEEPER_JWT_SECRET` unset → `log.Fatal` at the first token create or
  parse, which takes DNS and DHCP down with the web server.

# Decisions
- Fixes agreed 2026-10-17 (auth plan): never log passwords; a usable
  fallback login (one password per process, logged once); a missing JWT
  secret fails at startup with a clear message (fail-closed is intended);
  Secure/HttpOnly token cookie whose life matches the JWT; `/metrics` behind
  auth.
- Authorization by role: [OIDC roles](/web/auth-roles.md).
