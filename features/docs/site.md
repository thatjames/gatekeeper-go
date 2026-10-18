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
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T13:15:00Z }
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
  with the fallback password, OIDC and its cookie), modules, monitoring
  (`/metrics` needs a token; Prometheus scrape example), web pages with
  screenshots.
- README: coverage, release and Docker badges and the GitLab migration note.

# Decisions
- The docs follow the code on the htpasswd key (user, 2026-10-18): only
  `Web.HTPasswdFile` is documented, because `Auth.HTPasswdFile` is never
  read ([config file](/config/config-file.md)).
- The README Docker badge points at Docker Hub, while images are on GHCR:
  agreed fix (2026-10-17).
- `docs/.hugo_build.lock` and `docs/resources/_gen/` are tracked although
  `.gitignore` lists them.
