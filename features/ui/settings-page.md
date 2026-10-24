---
type: Feature
title: Settings from feature/settings-page
description: What came to main from the feature/settings-page branch - the DNS port in the DNS settings, a login screen shaped by the auth type, and a generated bcrypt htpasswd file on first start; no settings page of its own.
status: stable
code:
  - internal/web/ui/src/scenes/dns/DNSSettings.svelte
  - internal/web/ui/src/scenes/dns/DNSSettingsForm.svelte
  - internal/web/ui/src/scenes/login/Login.svelte
  - internal/web/v1/htpasswd.go
  - internal/web/v1/htpasswd_test.go
tags: [ui, settings, dns, auth]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T16:55:00Z }
verified: { by: human:slimjim, at: 2026-10-24T16:20:00Z }
---

# Intent
The branch (one commit, 2026-09-14) was called a settings page but held
three smaller changes; the user chose to keep those and drop the rest
(2026-10-24).

# Behavior
- DNS port: the DNS settings view shows the port (53 when unset) and its
  edit form has a number field for it; it is sent with `PUT /dns/config`
  ([DNS API](/web/dns-api.md)).
- Login screen ([UI shell](/ui/app-shell.md)): when `/features` lists
  `oidc`, the provider button is the main action and the password form is
  hidden behind a "Log in with a password" link; otherwise only the
  password form shows. There is no automatic redirect to the provider.
- Generated htpasswd ([Login and tokens](/web/auth.md)): at startup, when
  `Web.HTPasswdFile` is set and the file does not exist, gatekeeper makes
  a random 16-character password, writes `admin:<bcrypt hash>` to the file
  with mode 0600, and logs the password once at warn level. From then on
  logins check the file as usual.

# Decisions
- Dropped from the branch (user, 2026-10-24): its CORS block (main has
  `Web.AllowedOrigins`), `/features` as a list of objects (it stays a
  string list), and the `src/lib/settings` stub (deleted).
- No auto-redirect to OIDC (executor, 2026-10-24): after logout the login
  page would send the user straight back through the provider's session.
- The password form stays reachable with OIDC (executor, 2026-10-24): the
  server keeps `POST /login` open, so the admin can still log in when the
  provider is down.
- bcrypt and 0600 instead of the branch's plain-text, world-readable file
  (user, 2026-10-24).
