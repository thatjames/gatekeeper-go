---
type: Feature
title: UI shell and login
description: The Svelte 5 single-page admin UI - hash routes, side menu built from the running modules, home screen with system info, login by password or OIDC, error screen - and its Vite build into internal/web/ui/dist.
status: stable
code:
  - internal/web/ui/index.html
  - internal/web/ui/Makefile
  - internal/web/ui/package.json
  - internal/web/ui/README.md
  - internal/web/ui/jsconfig.json
  - internal/web/ui/svelte.config.js
  - internal/web/ui/vite.config.js
  - internal/web/ui/src/App.svelte
  - internal/web/ui/src/main.js
  - internal/web/ui/src/app.css
  - internal/web/ui/src/components/Nav.svelte
  - internal/web/ui/src/components/Footer.svelte
  - internal/web/ui/src/components/SideMenu.svelte
  - internal/web/ui/src/components/SimpleCard.svelte
  - internal/web/ui/src/lib/api/**
  - internal/web/ui/src/lib/auth/**
  - internal/web/ui/src/lib/common/**
  - internal/web/ui/src/lib/stores/**
  - internal/web/ui/src/lib/system/**
  - internal/web/ui/src/scenes/home/**
  - internal/web/ui/src/scenes/login/**
  - internal/web/ui/src/scenes/error/**
tags: [ui, svelte, auth]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:20:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
The admin UI, embedded in the binary, for the one admin.

# Behavior
- Build: `make web` runs `npm ci` and `vite build` (Svelte 5, Tailwind 4,
  Flowbite, ApexCharts, svelte-spa-router, jwt-decode); output goes to
  `dist/`, which the Go binary embeds. Aliases `$lib`, `$scenes`,
  `$components`.
- API base: `http://localhost:8085/api/v1/` in dev, `/api/v1/` in a
  production build. Requests send the Bearer token when there is one and
  cookies always. 401 → reload on `/#/auth/login`; a network error →
  the error screen ("Server Communications Failure").
- On load, a token in `sessionStorage` counts as logged in (its user is
  decoded from the token). Without one the app asks `GET /me` with cookies
  (no reload on 401): an answer sets the user and the page stays where it
  is; anything else goes to login. A token that has expired makes the
  next API call answer 401, which reloads on login.
- Login: username/password → `POST /login`, token kept in
  `sessionStorage`. When `/features` lists `oidc` a provider button goes
  to `/oidc-login` and the password form hides behind a link
  ([settings from the branch](/ui/settings-page.md)), and the login page asks `GET /me` (with cookies, no
  reload on 401): a user there lands on home, otherwise the form stays.
  The `oauth_token` cookie is HttpOnly, so the UI never reads it.
- Logout clears the session, calls `/logout` and reloads.
- `isAdmin()` is true when the logged-in user's `Role` is `admin`. For a
  viewer the DHCP and DNS pages show no Edit, Add, Reserve, Save or Delete
  buttons and opening a lease or local domain shows it read only
  ([OIDC roles](/web/auth-roles.md)).
- The side menu shows DHCP (Leases, Settings) and DNS (Local Domains,
  Statistics, Settings) only for running modules; the footer shows the
  version. Home shows the system info map.

# Interfaces
- [API routes](/web/api-routes.md), [login and tokens](/web/auth.md).

# Decisions
- Fixed 2026-10-24 (UI auth flow plan): the `auth_token` check is
  replaced by the session check above, and `verify()` (which called a
  `/verify` route the server never had) is gone; `GET /me` does its job.
