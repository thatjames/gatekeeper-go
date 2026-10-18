---
type: Feature
title: DNS resolver
description: Resolves a name and type from the blocklist, a TTL cache, local domains, a conditional forwarder for its zone, private PTR records or all upstreams in parallel, first answer wins.
status: stable
code:
  - internal/dns/resolver.go
  - internal/dns/forwarder.go
  - internal/dns/forwarder_test.go
  - internal/dns/fake_server_test.go
  - internal/dns/resolver_test.go
  - internal/dns/features/dns_resolver.feature
tags: [dns, cache]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T13:10:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Ad-blocking, local-name-aware recursion for the home LAN, as Pi-hole did.

# Behavior
- Order: blocklist → cache → local domains → conditional forwarder →
  (PTR for private IPs) → upstreams.
- Blocked name: one answer of the asked type, `0.0.0.0` for A, `::` for
  AAAA, empty RDATA for other types; TTL = cache TTL; counts
  `blockedDomainCounter`. Matching is exact: subdomains are not blocked.
- Cache key `name|type`; a hit before expiry returns the cached answers
  (authorities are not cached); an expired entry is deleted on lookup. There
  is no other eviction, so the cache grows without bound.
- Local domain: A query → its IPv4 with the cache TTL; any other type →
  empty NOERROR answer.
- PTR `a.b.c.d.in-addr.arpa` for a private IP: only the first local domain
  in map order is checked; a match answers its name, otherwise NXDOMAIN
  without asking upstream (bug). With no local domains the query goes
  upstream. A trailing `.` is not stripped (typo: the second TrimSuffix
  uses the original name).
- Conditional forwarders (`ConditionalForwarders`, zone → server): a name
  equal to a zone or under it goes to that zone's server only; the longest
  matching zone wins; zone and name are compared lower-case without a
  trailing dot. For a PTR query, a local domain whose IPv4 matches the
  reverse name is answered first (every local domain is checked; when
  several share the address, the alphabetically first name). The
  forwarder is asked like an upstream (new random ID, RD, dial and read
  timeouts); its answers and authorities are returned and cached like
  upstream ones, NXDOMAIN is passed on, NOERROR with no records is an empty
  success, and any other failure is `ErrDNSServerFailure` (SERVFAIL). The
  upstreams are never asked for a forwarded zone. A forwarder whose server
  is not IPv4[:port] or whose zone is empty is logged and skipped. Counted
  in `queryCounter` with the forwarder's `ip:port` as upstream.
- Asking a server (upstream or forwarder): a new random ID, RD set and an
  EDNS OPT record with UDP size 1232; the reply is read over UDP (up to
  64 KiB). When it has TC set, the same query goes to the same server over
  TCP (2-byte length prefix, its own dial and read timeouts) and that reply
  is used. Additional records in the reply are not returned.
- Upstreams: every upstream (IPv4 only, port 53) is asked at once; dial
  timeout 2 s, read timeout 2 s. The first
  answer with records wins and is cached with the first record's TTL.
  Each wait has its own 5 s upstream timeout. All failed → last error, or
  `ErrNxDomain` when none returned records.
- Upstream RCODE 1 → `ErrDNSFormatError`, 3 → `ErrNxDomain`, 2 →
  `ErrDNSServerFailure`.
- Defaults: cache TTL 5 min, upstreams 1.1.1.1 and 9.9.9.9, no
  conditional forwarders.
- `Resolve` holds one exclusive lock for the whole call, upstream wait
  included, so queries are answered one at a time (bug).
- Local domains reject `0.0.0.0` and non-IPv4 addresses.
- Metrics: `queryCounter` per domain, upstream (or `cache`,
  `local-domain`) and result.

# Interfaces
- `Resolver` interface: Resolve, Add/DeleteLocalDomain,
  AddBlocklistEntries, DeleteBlocklistEntry, FlushBlocklist.
- [Wire format](/dns/wire-format.md).

# Decisions
- Fixes agreed 2026-10-17 (plans): per-query locking and a bounded cache;
  the PTR bug, folded together with the `feature/ptr_records` branch.
