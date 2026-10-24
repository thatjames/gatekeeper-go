---
type: Feature
title: OIDC roles
description: An OIDC login gets the gatekeeper role (admin or viewer) its groups claim maps to, or is refused; viewers may use every GET route and get 403 on changes; htpasswd users are admin.
status: stable
code:
  - internal/web/v1/roles.go
  - internal/web/v1/roles_test.go
tags: [web, auth, oidc, roles]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:20:00Z }
verified: { by: human:slimjim, at: 2026-10-24T16:20:00Z }
---

# Intent
Not every Authentik user may change DNS. Admins change settings; viewers
only look (user, 2026-10-24).

# Behavior
- Two roles: `admin` (everything) and `viewer` (read only).
- OIDC config keys under `Auth:`:
  - `RoleClaim` - the ID token claim that names the user's groups;
    default `groups` (Authentik sends it without a property mapping).
    Its value may be a list of strings or one string.
  - `Roles` - a map from a claim value to a gatekeeper role, for example
    `{gatekeeper-admins: admin, gatekeeper-viewers: viewer}`. Without it,
    the claim values `admin` and `viewer` map to themselves.
- At the OIDC callback the claim's values are mapped; when several map,
  `admin` wins. No value maps → the login is refused: 403
  `{error: "No gatekeeper role for this user"}`, a warning with the email
  and the claim name, and no token cookie.
- htpasswd logins (and the fallback admin) get `admin`.
- The role goes into the app JWT ([Login and tokens](/web/auth.md)).
- On the protected routes, after the token check: a token whose role is
  neither `admin` nor `viewer` → 403. `GET` and `HEAD` are allowed for both
  roles; any other method needs `admin`, else 403
  `{error: "Forbidden"}`.
- The UI hides the actions a viewer cannot use
  ([UI shell](/ui/app-shell.md)).

# Interfaces
- `config.OIDCAuth.RoleClaim`, `config.OIDCAuth.Roles`
  ([Config file](/config/config-file.md)).
- `RoleFromClaims(claims map[string]any, claim string, roles map[string]string) (Role, bool)`,
  `RoleMiddleware()` in `internal/web/v1/roles.go`.

# Decisions
- Role set admin + viewer (user, 2026-10-24).
- Claim name configurable, default `groups` (executor recommendation,
  2026-10-24): Authentik sends `groups` by default; a `roles` claim needs
  a property mapping.
- Read vs change is decided by the HTTP method, not per route: every
  change route is a non-GET method, so a new route is covered by default.
- Tokens with the old role `user` (from before roles) are refused with
  403; nothing was deployed, so no one is locked out.
