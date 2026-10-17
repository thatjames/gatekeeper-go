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
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Public documentation at https://thatjames.github.io/gatekeeper-go/.

# Behavior
- `docs.yml` builds `docs/` with the latest extended Hugo (submodules
  checked out) and deploys to GitHub Pages; also by hand.
- Pages: getting started (htpasswd, Docker, DHCP with host networking
  unsupported, binaries), configuration, auth, modules, monitoring, web
  pages with screenshots.
- README: coverage, release and Docker badges and the GitLab migration note.

# Decisions
- The docs disagree with the code on the fallback login (getting started
  says admin/admin, auth says a random password in the logs) and never name
  `GATEKEEPER_JWT_SECRET`; the auth plan updates them.
- The README Docker badge points at Docker Hub, while images are on GHCR:
  agreed fix (2026-10-17).
- `docs/.hugo_build.lock` and `docs/resources/_gen/` are tracked although
  `.gitignore` lists them.
