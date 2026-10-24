---
type: Feature
title: Documentation site
description: Hugo docs with the hugo-book theme (submodule), published to GitHub Pages on pushes to main that touch docs/ - getting started, configuration, auth, modules, monitoring, web pages; plus the README.
status: stable
code:
  - docs/**
  - .gitmodules
  - .github/workflows/docs.yml
  - README.md
tags: [docs, hugo]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:30:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Public documentation at https://thatjames.github.io/gatekeeper-go/.

# Behavior
- `docs.yml` builds `docs/` with the latest extended Hugo (submodules
  checked out) and deploys to GitHub Pages; also by hand.
- Pages: getting started (htpasswd, `GATEKEEPER_JWT_SECRET`, Docker with
  the secret passed in, DHCP with host networking unsupported, binaries),
  configuration, auth (JWT secret, password login from `Web.HTPasswdFile`
  with the generated file on first start and the fallback password, OIDC
  and its cookie, roles admin and viewer), modules, monitoring
  (`/metrics` needs a token; Prometheus scrape example), web pages with
  screenshots.
- README: coverage, release and container image badges and the GitLab
  migration note. The image badge shows the latest release (CI pushes an
  image per tag) and links to the `ghcr.io/thatjames/gatekeeper-go`
  package.

# Decisions
- The docs follow the code on the htpasswd key (user, 2026-10-18): only
  `Web.HTPasswdFile` is documented, because `Auth.HTPasswdFile` is never
  read ([config file](/config/config-file.md)).
- The image badge shows the release version (2026-10-24): shields.io has
  no GHCR pull count.
- `docs/.hugo_build.lock` and `docs/resources/_gen/` are Hugo build output
  and untracked, as `.gitignore` says (2026-10-24).
