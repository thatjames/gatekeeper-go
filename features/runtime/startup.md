---
type: Feature
title: Startup and shutdown
description: The gatekeeper binary loads the YAML config, sets up logging, registers the DHCP, DNS and web modules that are configured, starts them and stops them on SIGINT/SIGTERM.
status: stable
code:
  - cmd/gatekeeper/**
tags: [runtime, logging]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
One process that runs only the modules the config file names, so the same
binary serves DNS on atlas (with DHCP left to hecate) or DNS+DHCP elsewhere.

# Behavior
- Flags: `-config <file>` (default `config.yml`), `-debug`, `-trace`.
  `-debug` wins over `-trace` when both are given (level is Debug), but
  `-trace` still switches the log line format to include the caller.
- A config file that cannot be opened or parsed panics before anything starts.
- Log lines: `YYYY-MM-DD HH:MM:SS LEVEL - message key: value ...`; with
  `-trace`, the calling function sits after the timestamp.
- `version` is set at build time with `-X main.version` (default
  `development-build`) and copied to `system.Version`.
- `DHCP:` present → the DHCP server is built from it and registered.
- `DNS:` present → `LocalDomains` values that are not IPv4 are skipped with
  a warning; the DNS server is registered with interface, port, upstreams,
  blocklist URLs, blocked domains and local domains from the config.
- `Web:` present → the web server starts in its own goroutine; an error
  there is logged and does not stop DNS or DHCP.
- Registered services start in map order; the first start error is fatal.
- On SIGINT or SIGTERM every service is stopped; a stop error is logged as
  an unclean exit.
- Built with tag `live` (all Makefile and CI builds), gin runs in release
  mode.

# Interfaces
- [Service manager](/runtime/service-manager.md), [config file](/config/config-file.md),
  [DNS server](/dns/server.md), [DHCP server](/dhcp/server.md),
  [HTTP server](/web/http-server.md).

# Decisions
- A commented-out routing manager in `main.go` belongs to router mode
  ([router mode](/runtime/router-mode.md)), which lives on `feature/router`.
