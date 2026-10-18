---
type: Feature
title: DHCP API
description: Authenticated /api/v1/dhcp routes to list, release, reserve and change leases and to read and change DHCP options, writing changes back to the config file.
status: stable
code:
  - internal/web/v1/dhcp.go
tags: [web, api, dhcp]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T22:10:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Manage leases and DHCP settings from the UI.

# Behavior
- `GET /dhcp/leases` → `{active, reserved}` lists of
  `{clientId, hostname, ip, state, expiry}`.
- `DELETE /dhcp/leases/:clientId` deletes the lease through the pool's
  `DeleteLease` (a reservation is deleted, also from the config, unlike a
  client's RELEASE); unknown client → 400. Answers the lists.
- `POST /dhcp/leases/reserve {clientId, ip}` reserves, saves the config and
  answers the lists; an IP reserved for another client → 400.
- `PUT /dhcp/leases {clientId, ip}` changes a reservation or lease IP and
  saves the config.
- `GET /dhcp/options` → interface, range, TTL, gateway, mask, domain, lease
  file, DNS servers.
- `PUT /dhcp/options` validates, swaps the server options, restarts the
  server (old options back on a failed stop or start), then replaces the
  config's DHCP section and saves it. The new range does not reach the pool
  until a process restart ([DHCP server](/dhcp/server.md)).

# Interfaces
- [Lease pool](/dhcp/lease-pool.md), [config file](/config/config-file.md).
