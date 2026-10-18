---
type: Feature
title: DNS server
description: A UDP DNS server on the configured port that hands each query to the resolver and answers with its records, and can be reconfigured from the API.
status: stable
code:
  - internal/dns/server.go
  - internal/dns/types.go
  - internal/dns/server_test.go
  - internal/dns/dns_feature_test.go
  - internal/dns/features/dns_request.feature
tags: [dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T10:00:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
The LAN's DNS server on atlas, replacing Pi-hole: hecate's dnsmasq forwards
to it ([Pi-hole parity](/dns/pihole-parity.md)).

# Behavior
- Defaults: port 53, upstreams 8.8.8.8 and 1.1.1.1, read deadline 2 s. The
  resolver gets the server's upstreams when its own list is empty.
- `Start()` adds `BlockedDomains` to the resolver, listens on UDP4 `:<port>`
  on all interfaces (the `Interface` option is not used), starts one
  listener, one receive worker and one response worker, then fetches the
  blocklist URLs ([blocklists](/dns/blocklist.md)) before it returns.
- Each datagram (up to 1500 bytes) is parsed; a packet that does not parse
  is logged and dropped with no answer.
- The receive worker resolves each query in its own goroutine, at most 256
  at a time; with 256 in flight it waits for one to finish before it takes
  the next packet. Answers may leave in a different order than the queries
  came in.
- Only the first question is resolved. A query with no question panics the
  receive worker (bug).
- Resolver result → RCODE: success 0 with answers and authorities set;
  `ErrNxDomain` 3 with the resolver's authorities (the SOA); any other
  error 2 (server failure).
- The reply sets QR and RA, drops all additional records (EDNS OPT too) and
  goes back to the sender. There is no TCP and no truncation.
- Metrics (defined in `types.go`): `dns_req_time`, `dns_query_count`,
  `dns_blocked_domain_count`, `dns_query_by_ip_count`. There is no cache-hit
  metric; cache hits count in `dns_query_count` with upstream `cache`.
- `queryByIPCounter` per client IP with `result` (always
  `success` in practice: resolution errors are not counted as failed) and
  `reqDuration` in ms.
- `Stop()` closes the listener and waits for it; the workers keep running,
  so every restart adds another pair (leak).
- API helpers: add/delete local domain, add/delete blocked domain, flush the
  blocklist, add a blocklist from a URL, reload all blocklist URLs,
  `Options()` / `UpdateOptions()`.

# Interfaces
- [Resolver](/dns/resolver.md), [wire format](/dns/wire-format.md),
  [DNS API](/web/dns-api.md).

# Decisions
- Godog scenarios in `features/*.feature` drive the feature tests
  (`make test` runs `-run TestFeatures`).
- Fixes agreed 2026-10-17: empty-question panic, the failed metric and the
  worker leak (plans), TCP/truncation/EDNS/interface binding under
  [Pi-hole parity](/dns/pihole-parity.md).
- Parallel receive capped at 256 queries in flight: one slow lookup no
  longer holds up the LAN, and a flood of queries cannot start unbounded
  goroutines (user, 2026-10-18).
