---
type: Feature
title: Login and tokens
description: Users log in with htpasswd credentials or OIDC (Authentik) and get a 7-day HS256 JWT signed with GATEKEEPER_JWT_SECRET, sent as a Bearer header or the oauth_token cookie.
status: stable
code:
  - internal/web/v1/api.go
  - internal/web/v1/jwt.go
  - internal/web/v1/middleware.go
  - internal/web/v1/auth_test.go
tags: [web, auth, oidc, jwt]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:20:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Only the admin (the user, alone) can change DNS and DHCP. The identity
provider is Authentik on atlas (confirmed 2026-10-17).

# Behavior
- `POST /login {username, password}` logs the attempt and its outcome with
  the username only; the password and the request body are never logged.
  The route stays open when OIDC is configured.
- At startup, when `Web.HTPasswdFile` is set and the file does not exist,
  gatekeeper writes one: `admin` with a random 16-character password
  (`crypto/rand`), bcrypt-hashed, mode 0600, the password logged once at
  warn level ([settings from the branch](/ui/settings-page.md)).
- Then the file is read once to see whether it can be used. When it can,
  each login is checked against it (any htpasswd hash, re-read per login so
  edits apply). When it cannot (unset, unreadable, or the write above
  failed), gatekeeper makes one random 16-character password, logs it once
  at warn level, and accepts `admin` with that password, compared in
  constant time, for the life of the process.
- Success → `{token}`: HS256 JWT, subject and username = login name, role
  `admin`, issuer `gatekeeper`, valid 7 days. Wrong credentials → 401.
- The JWT secret is `GATEKEEPER_JWT_SECRET`, loaded at startup; with `Web:`
  configured and the variable unset or empty, gatekeeper refuses to start
  ([startup](/runtime/startup.md)). Creating or parsing a token without a
  loaded secret returns an error; it never exits the process.
- OIDC (`AuthType: oidc`): discovery from `IssuerURL` at startup (on failure
  the OIDC routes are not added). `/oidc-login` sets a random state cookie
  (10 min, Secure, HttpOnly) and redirects to the provider. `/auth/callback`
  checks the state, exchanges the code, verifies the ID token for
  `ClientID`, maps its role claim to a gatekeeper role (none → 403, no
  cookie; [OIDC roles](/web/auth-roles.md)), and issues the app JWT for the
  token's email with that role in the `oauth_token` cookie, then redirects
  to `WebURL`.
- The `oauth_token` cookie: path `/`, HttpOnly, SameSite Lax, Secure when
  `WebURL` starts with `https://`, max age 7 days (the JWT life). Scripts
  cannot read it; the UI learns the user from `GET /me`.
- The client secret comes from the variable named by `ClientSecretVar`.
- Middleware: token from `Authorization: Bearer`, else the `oauth_token`
  cookie. None → 401. Invalid or expired → 403 and the cookie is cleared.
  It guards the protected API routes and `/metrics`
  ([HTTP server](/web/http-server.md)); on the API routes the role check
  follows it ([OIDC roles](/web/auth-roles.md)).
- `GET /me` (behind the middleware) answers `{Username, Role}` of the token.
- `/logout` clears the `oauth_token` cookie (same attributes, max age -1).

# Decisions
- A missing JWT secret fails closed, at startup (user, 2026-10-17): a web
  server that cannot sign tokens is useless, and dying at the first login
  took DNS down with it.
- The fallback password is per process and logged once (2026-10-17), so a
  fresh install can log in without an htpasswd file, and the log shows the
  password only at startup, never per request. With a configured path the
  generated file replaces it (2026-10-24), so the password survives
  restarts.
- The cookie is HttpOnly and the UI asks `GET /me` who is logged in (user,
  2026-10-18); `/me` also stands in for the `/verify` the UI calls
  ([UI shell](/ui/app-shell.md)).
- Authorization by role: [OIDC roles](/web/auth-roles.md).
