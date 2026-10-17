---
type: Feature
title: DNS API
description: Authenticated /api/v1/dns routes to read and change upstreams and interface, manage local domains, blocklist URLs and single blocked domains, writing changes back to the config file.
status: stable
code:
  - internal/web/v1/dns.go
tags: [web, api, dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Manage DNS from the UI, as Pi-hole's admin pages did.

# Behavior
- `GET /dns/config` → `{upstreams, interface, blocklist, blockedDomains}`.
- `PUT /dns/config {interface, upstreams, blockedDomains}` restarts the DNS
  server with them and saves interface and upstreams to the config. The
  running resolver keeps its old upstreams (bug), and blocked domains are
  neither saved nor removed from the resolver.
- `GET /dns/local-domains` → the name → IP map.
- `POST /dns/local-domains {domain, ip}` adds it live and saves. A live add
  error still saves and answers twice (bug: no return).
- `PUT /dns/local-domains/:domain {domain, ip}` renames or re-points it.
- `DELETE /dns/local-domains/:domain` removes it.
- `POST /dns/blocklist {url}` appends the URL and saves, fetches it (fetch
  error → 400, URL stays saved), then flushes and reloads every list. Answers
  an empty 200.
- `DELETE /dns/blocklist/:index` removes the URL, saves, flushes and reloads;
  empty 200.
- `PUT /dns/blockeddomains {url: <domain>}` blocks one domain and saves.
- `DELETE /dns/blockeddomains/:index` unblocks and saves.
- Rollback on a failed save reuses a slice that `append` already changed.

# Decisions
- Fixes agreed 2026-10-17 (plans): apply config changes to the running
  resolver; the handler bugs, bodies, rollback and races.
