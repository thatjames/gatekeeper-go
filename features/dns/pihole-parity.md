---
type: Feature
title: Pi-hole parity on atlas
description: draft - what gatekeeper's DNS must do before it replaces Pi-hole on atlas - TCP and UDP on port 53, truncation and EDNS, conditional forwarding of reverse and local lookups to hecate, interface binding.
status: draft
code: []
tags: [dns, atlas, pihole]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
gatekeeper replaces the Pi-hole service on atlas (confirmed 2026-10-17).
Hecate's dnsmasq forwards LAN queries to it; DHCP stays on hecate.

# Behavior
- Listens on port 53 over UDP and TCP.
- Replies larger than 512 bytes (or the client's EDNS size) set TC over UDP;
  EDNS OPT is answered, not dropped.
- Conditional forwarding: reverse lookups for the LAN (`10.20.1.0/24`) and
  the local domain go to hecate (`10.20.1.1`), as Pi-hole does today.
- The `Interface` option binds the listener to that interface's address.
- Upstreams as configured (Pi-hole on atlas uses Quad9).

# Decisions
- *open*: which image atlas pulls (GHCR release or registry.thatjames.nl) and
  where the compose entry lives: not decided (2026-10-17).
- The local domain and subnet follow hecate's config at the time of the
  switch; check them then.
