---
type: Feature
title: GitHub Actions CI and release
description: On push and PR, CI builds the UI and runs the tests with coverage to Codecov; a pushed tag builds five Linux binaries, pushes the multi-arch GHCR image with provenance and creates a GitHub release.
status: stable
code:
  - .github/workflows/test.yml
  - .github/workflows/release.yml
tags: [ci, release, github]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- `test.yml`: on pushes touching `cmd/`, `internal/`, `pkg/`, `go.mod`,
  `go.sum`, `Makefile` or the workflow, and on every PR: Go 1.25.4,
  Node 24, `make web`, `make test`, coverage uploaded to Codecov with
  `CODECOV_TOKEN`; a second job builds and uploads the UI.
- `release.yml`, on any tag: tests with reports (kept 90 days), the UI,
  binaries for amd64, armv7, armv8, arm64 and 386; then the image
  `ghcr.io/<repo>` for amd64, arm64, arm/v7 and 386, tagged by
  docker/metadata-action, with a build provenance attestation; then a
  GitHub release with the binaries. Uses `GITHUB_TOKEN` only.

# Decisions
- CI uses Go 1.25.4 while `go.mod` says 1.24.6; both build.
- The armv8 binary is built but the image leaves out `arm/v8`.
