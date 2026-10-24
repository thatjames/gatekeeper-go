---
type: Feature
title: DNS API
description: Authenticated /api/v1/dns routes to read and change upstreams and interface, manage local domains, blocklist URLs and single blocked domains, writing changes back to the config file.
status: stable
code:
  - internal/web/v1/dns.go
  - internal/web/v1/dns_test.go
tags: [web, api, dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T16:55:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Manage DNS from the UI, as Pi-hole's admin pages did.

# Behavior
- `GET /dns/config` → `{upstreams, interface, port, blocklist,
  blockedDomains}`; `port` is 53 when the config leaves it unset.
- `PUT /dns/config {interface, port, upstreams, blockedDomains}` restarts
  the DNS server with them and saves interface, port and upstreams to the
  config. `port` 0 or absent keeps the current port; outside 1-65535 → 400
  on the `port` field. A failed restart puts the old interface, port and
  upstreams back. The
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
  copied before they change, never changed in place. For local domains the
  resolver change is undone too; blocked domains and blocklists reach the
  resolver only after the save. (`PUT /dns/config` restores the config
  only; the server keeps the restart.)
- Handlers read the config under the [config lock](/config/config-file.md)
  and hold it for each read-modify-write, through the save and the answer.
  `POST /dns/blocklist` fetches before it takes the lock; `DELETE
  /dns/blocklist` reloads the remaining lists under it, so no other change
  lands between the flush and the reload.
- `dns_test.go` drives the handlers against a fake resolver and fetcher and
  a temp config file, and runs them concurrently for `go test -race`.

# Decisions
- Fix agreed 2026-10-17 (plans): apply config changes to the running
  resolver. The handler errors, bodies, rollback and races were fixed
  2026-10-18.
- A blocklist URL is fetched before it is saved (2026-10-18): a URL that
  cannot be fetched is not kept, and the fetch is the only
  one.
