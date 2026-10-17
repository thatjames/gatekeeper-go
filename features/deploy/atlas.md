---
type: Feature
title: Atlas deployment
description: draft - gatekeeper runs on atlas as the LAN DNS server in place of Pi-hole, DNS only (hecate serves DHCP), with OIDC through Authentik; not deployed yet.
status: draft
code: []
tags: [deploy, atlas, dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
The home deployment (user, 2026-10-17): gatekeeper replaces the Pi-hole on
atlas, "similar to how pihole does". Not running anywhere today.

# Behavior
- A container on atlas serves DNS on port 53 (TCP and UDP) for the LAN;
  hecate's dnsmasq forwards to it.
- `DHCP:` is left out of the config; hecate stays the only DHCP server.
- The web UI logs in through Authentik on atlas (OIDC client `gatekeeper`);
  secrets come from `GATEKEEPER_JWT_SECRET` and
  `GATEKEEPER_OIDC_CLIENT_SECRET`.
- Before the switch: [Pi-hole parity](/dns/pihole-parity.md).

# Decisions
- *open*: GHCR release image or registry.thatjames.nl, and where its compose
  entry lives (nix-flake's atlas compose holds Pi-hole today).
