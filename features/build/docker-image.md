---
type: Feature
title: Docker image
description: Alpine images that run /app/gatekeeper with prebuilt binaries - a single-arch image and a multi-arch one picking the binary by TARGETPLATFORM.
status: stable
code:
  - Dockerfile
  - Dockerfile.multiarch
tags: [build, docker]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- Both: `alpine:latest` with CA certificates and tzdata, workdir `/app`,
  entrypoint `/app/gatekeeper` (flags such as `-config` passed as args;
  default config path `/app/config.yml`).
- `Dockerfile` copies `bin/gatekeeper`, exposes 8085 and 53.
- `Dockerfile.multiarch` copies `bin/gatekeeper-linux-*` and keeps the one
  for `linux/amd64`, `arm64`, `arm/v7`, `arm/v8` or `386`; any other
  platform fails the build. Exposes 8085, 53 and 67.
- Only UDP is exposed in practice; Docker `EXPOSE` lines name no protocol.

# Decisions
- *open*: which image atlas runs (GHCR or registry.thatjames.nl), see
  [atlas deployment](/deploy/atlas.md).
