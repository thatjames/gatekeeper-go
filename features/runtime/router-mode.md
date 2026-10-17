---
type: Feature
title: Router mode
description: draft - gatekeeper as a full router on Linux through netlink; work in progress on the feature/router branch, a long-term goal not used at home.
status: draft
code: []
tags: [runtime, router, netlink]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
A long-term goal: let gatekeeper act as the router of a network, beyond DNS
and DHCP. Not used at home, where hecate (OpenWrt) is the router.

# Behavior
- Not on main. The `feature/router` branch (3 commits, last 2025-11-29) adds
  `internal/netlink` and a router mode in the service manager.

# Decisions
- *open*: what router mode covers (routes, NAT, firewall?) and how it is
  switched on. To settle when the branch is picked up.
