---
type: Feature
title: UI DNS pages
description: DNS settings (upstreams, interface, blocklist URLs, blocked domains) and local domain pages.
status: stable
code:
  - internal/web/ui/src/scenes/dns/DNSLayout.svelte
  - internal/web/ui/src/scenes/dns/DNSSettings.svelte
  - internal/web/ui/src/scenes/dns/DNSSettingsForm.svelte
  - internal/web/ui/src/scenes/dns/DNSLocalDomains.svelte
  - internal/web/ui/src/lib/dns/**
tags: [ui, dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- `#/dns/settings`: upstreams, interface, blocklist URLs and blocked domains;
  "Add Blocklist" and "Add Blocked Domain" add one; each list item can be
  removed by index; the edit form saves interface and upstreams.
- `#/dns/local-domains`: the name → IP table with add, edit and delete.

# Interfaces
- [DNS API](/web/dns-api.md).
