---
type: Feature
title: DNS API
description: Authenticated /api/v1/dns routes to read and change upstreams and interface, manage local domains, blocklist URLs and single blocked domains, writing changes back to the config file.
status: stable
code:
  - internal/web/v1/dns.go
tags: [web, api, dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T21:59:00Z }
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
- `POST /dns/local-domains {domain, ip}` adds it live and saves; answers the
  map. A live add error → 400, and the config (memory and file) is unchanged.
- `PUT /dns/local-domains/:domain {domain, ip}` renames or re-points it;
  answers the map.
- `DELETE /dns/local-domains/:domain` removes it; answers the map.
- `POST /dns/blocklist {url}` fetches the URL once (fetch error → 400, config
  unchanged), appends it and saves, then adds its entries to the running
  blocklist. The other lists are not fetched again. Answers
  `DNSConfigResponse`.
- `DELETE /dns/blocklist/:index` removes the URL, saves, then flushes and
  reloads the remaining lists. Answers `DNSConfigResponse`.
- `PUT /dns/blockeddomains {url: <domain>}` blocks one domain and saves;
  answers `DNSConfigResponse`.
- `DELETE /dns/blockeddomains/:index` unblocks and saves; answers
  `DNSConfigResponse`.
- Every error answers once, with `{error}` or `ErrorResponse`, and the
  handler stops there.
- A failed save (500) puts the config back as it was: slices and maps are
  copied before they change, and a live change already made to the resolver
  is undone.
- Handlers read the config under the [config lock](/config/config-file.md)
  and hold it for each read-modify-write including the save. Blocklist
  fetches run outside the lock.

# Decisions
- Fixes agreed 2026-10-17 (plans): apply config changes to the running
  resolver; the handler bugs, bodies, rollback and races.
- A blocklist URL is fetched before it is saved (2026-10-18, web handler
  plan): a URL that cannot be fetched is not kept, and the fetch is the only
  one.
