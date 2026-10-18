---
type: Feature
title: DNS server
description: A UDP and TCP DNS server on the configured port that hands each query to the resolver and answers with its records, and can be reconfigured from the API.
status: stable
code:
  - internal/dns/server.go
  - internal/dns/types.go
  - internal/dns/server_test.go
  - internal/dns/dns_feature_test.go
  - internal/dns/features/dns_request.feature
tags: [dns]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T12:40:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
The LAN's DNS server on atlas, replacing Pi-hole: hecate's dnsmasq forwards
to it ([Pi-hole parity](/dns/pihole-parity.md)).

# Behavior
- Defaults: port 53, upstreams 8.8.8.8 and 1.1.1.1, read deadline 2 s. The
  resolver gets the server's upstreams when its own list is empty. Startup
  passes `DNS.ConditionalForwarders` to the resolver.
- `Start()` adds `BlockedDomains` to the resolver, listens on UDP4 and
  TCP4 `:<port>` on all interfaces (the `Interface` option is not used),
  starts the UDP listener with one receive worker and one response worker
  and the TCP accept loop, then fetches the blocklist URLs
  ([blocklists](/dns/blocklist.md)) before it returns. When either listen
  fails, `Start` closes what it opened and returns the error.
- TCP: each connection is served by its own goroutine. Messages carry a
  2-byte big-endian length prefix (RFC 1035 4.2.2); the connection may send
  several queries, answered in order, and is closed after 10 s with no
  query, on a short read, or on a message that does not parse or has no
  question. TCP replies go through the same resolver and reply building as
  UDP and are never truncated.
- Each datagram (up to 1500 bytes) is parsed; a packet that does not parse
  is logged and dropped with no answer.
- Only the first question is resolved. A query with no question panics the
  receive worker (bug).
- Resolver result → RCODE: success 0 with answers and authorities set;
  `ErrNxDomain` 3; any other error 2 (server failure).
- The reply sets QR and RA, clears TC and drops the query's additional
  records. When the query had an OPT record (EDNS), the reply carries one
  OPT of its own: UDP size 1232, extended RCODE 0, version 0, DO clear, no
  options. A query without OPT gets none. The reply goes back to the
  sender.
- Truncation (UDP only): the client's limit is 512 bytes, or the UDP size
  in the query's OPT record, raised to 512 and capped at 1232. A reply over
  the limit is sent with TC set, the question and OPT kept and all answers
  and authorities removed, so the client asks again over TCP.
- Metrics (defined in `types.go`): `dns_req_time`, `dns_query_count`,
  `dns_blocked_domain_count`, `dns_query_by_ip_count`. There is no cache-hit
  metric; cache hits count in `dns_query_count` with upstream `cache`.
- `queryByIPCounter` per client IP with `result` (always
  `success` in practice: resolution errors are not counted as failed) and
  `reqDuration` in ms.
- `Stop()` closes the UDP listener and waits for it, closes the TCP
  listener and every open TCP connection; the UDP workers keep running, so
  every restart adds another pair (leak).
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
