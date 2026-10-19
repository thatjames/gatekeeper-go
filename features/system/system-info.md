---
type: Feature
title: System info
description: Linux host facts for the home screen - hostname, uptime, memory, and Tx/Rx of the DHCP interface and ppp0 - plus the network interface list.
status: stable
code:
  - internal/system/types.go
  - internal/system/system_linux.go
tags: [system, linux]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T21:59:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Show the state of the box gatekeeper runs on.

# Behavior
- `Hostname`, `Uptime` (`1d:2h:3m:4s`, zero parts left out), `Memory` as
  `free / total` (label says Memory, value starts with free), sizes as
  `1.5G`.
- `LAN Interface/Tx/Rx` from `/sys/class/net/<DHCP.Interface>/statistics`
  when DHCP is configured (read under the config read lock, since the DHCP
  API replaces that section); `WAN Interface/Tx/Rx` from `ppp0` when it
  exists. On atlas (DNS only, no ppp0) neither shows.
- `Version` holds the build version.

# Decisions
- Linux only (hard constraint, 2026-10-17).
- Fix agreed 2026-10-17 (plan): WAN interface not hardcoded to ppp0, memory
  shown as used/total or labelled free.
