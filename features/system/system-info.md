---
type: Feature
title: System info
description: Linux host facts for the home screen - hostname, uptime, used/total memory, and Tx/Rx of the DHCP interface and the WAN (default-route or configured) interface - plus the network interface list.
status: stable
code:
  - internal/system/types.go
  - internal/system/system_linux.go
  - internal/system/system_linux_test.go
  - internal/system/testdata/**
tags: [system, linux]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:45:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Show the state of the box gatekeeper runs on.

# Behavior
- `Hostname`, `Uptime` (`1d:2h:3m:4s`, zero parts left out), sizes as
  `1.5G`.
- `Memory` is `used / total`: total is `MemTotal` and used is `MemTotal -
  MemAvailable` from `/proc/meminfo`, so page cache counts as available.
  When `/proc/meminfo` cannot be read, `sysinfo` total minus free is used.
- `LAN Interface/Tx/Rx` from `/sys/class/net/<DHCP.Interface>/statistics`
  when DHCP is configured (read under the config read lock, since the DHCP
  API replaces that section).
- `WAN Interface/Tx/Rx`: the interface is `Web.WANInterface` when set,
  otherwise the interface of the IPv4 default route in `/proc/net/route`
  (destination and mask 0, route up; the lowest metric wins). No such
  interface, or no statistics for it → no WAN entries.
- `Version` holds the build version.

# Interfaces
- `Web.WANInterface` ([Config file](/config/config-file.md)).
- The paths `/proc/meminfo`, `/proc/net/route` and `/sys/class/net` are
  package variables, so the tests read fixture files in
  `internal/system/testdata/`.

# Decisions
- Linux only (hard constraint, 2026-10-17).
- WAN from the default route with an optional config override (executor
  recommendation, 2026-10-24): atlas has no `ppp0`, the old hardcoded name.
- `MemAvailable` rather than free memory (2026-10-24): free memory leaves
  out the page cache, so "used" would read far too high.
