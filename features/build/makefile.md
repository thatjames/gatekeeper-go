---
type: Feature
title: Make targets and Go module
description: The Makefile builds the UI and the Go binary (native, pipeline, Docker, multi-arch), runs the tests and test reports, and runs locally; go.mod pins Go 1.24.6 and the dependencies.
status: stable
code:
  - Makefile
  - go.mod
tags: [build, make, go]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:22:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- `VERSION` is the latest git tag (or `dev`), linked into `main.version`.
- `make` / `all`: web, test, build. `build`: tests, UI, then
  `bin/gatekeeper` with tag `live`. `build-pipeline`: CGO off, GOOS/GOARCH/
  BIN_NAME from the environment (CI).
- `docker` / `docker-multiarch`: Linux binaries then the image(s)
  `ghcr.io/thatjames/gatekeeper-go` (the name CI pushes); multi-arch pushes
  amd64, arm64, arm/v7 and 386 with buildx.
- `test`: generate-mocks, then the godog feature tests of dhcp and dns and
  the packet test. `test-report`: JUnit reports and total coverage, with a
  dummy `dist` so the embed compiles.
- `local-run`: `go run` with `config/config.yml`, `-trace`, a random
  `GATEKEEPER_JWT_SECRET` and `GATEKEEPER_OIDC_CLIENT_SECRET` from the
  untracked `.client_secret` file.
- `docker-run`: the image with `config/docker-config.yml`, leases in
  `/tmp/leases`, ports 8085, 53 and 67.
- `generate-mocks` / `install-mockgen` and `clean` handle
  `internal/datasource/mocks`, which does not exist; there is no
  `go:generate` directive, so `make test` installs mockgen for nothing.
- `docs_server`: Hugo dev server for the docs.

# Decisions
- The module path is `github.com/thatjames/gatekeeper-go`, the
  repository's own (moved from the GitLab-era
  `gitlab.com/thatjames-go/gatekeeper-go` on 2026-10-24). Agreed fix still
  open (2026-10-17): drop the mock targets.
- Linux only, multi-arch images must keep building (hard constraint).
