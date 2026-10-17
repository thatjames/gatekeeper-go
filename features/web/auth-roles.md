---
type: Feature
title: OIDC roles
description: draft - map the OIDC roles claim to gatekeeper roles (admin, user and more) and allow API actions by role, so not every Authentik user can change DNS.
status: draft
code: []
tags: [web, auth, oidc, roles]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
The closest feature to done (user, 2026-10-17): roles are parsed from the
claims but not used. Admins change settings; other roles see less.

# Behavior
- The `roles` claim of the ID token decides the role in the app JWT; with
  no matching role the login is refused.
- Read-only routes allow every role; changes need `admin`.

# Decisions
- *open*: which roles beyond `admin` and `user` ("just need to think of
  some more roles") and which routes each may use.
