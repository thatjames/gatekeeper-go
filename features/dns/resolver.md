---
type: Feature
title: DNS resolver
description: Resolves a name and type from the blocklist, a bounded TTL cache, local domains, private PTR records or all upstreams in parallel, first answer wins; queries resolve concurrently.
status: stable
code:
  - internal/dns/resolver.go
  - internal/dns/resolver_test.go
  - internal/dns/cache.go
  - internal/dns/cache_test.go
  - internal/dns/features/dns_resolver.feature
tags: [dns, cache]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T10:00:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Ad-blocking, local-name-aware recursion for the home LAN, as Pi-hole did.

# Behavior
- Order: blocklist → cache → local domains → (PTR for private IPs) →
  upstreams.
- Blocked name: one answer of the asked type, `0.0.0.0` for A, `::` for
  AAAA, empty RDATA for other types; TTL = cache TTL; counts
  `blockedDomainCounter`. Matching is exact: subdomains are not blocked.
- Cache key `name|type`; an entry holds the answers and the authorities.
  A hit before expiry returns both; an expired entry is a miss.
- The cache holds at most 10 000 entries. Adding a new key to a full cache
  first evicts the entry nearest expiry. Once a minute a sweep deletes
  expired entries; `Close()` stops the sweep. An entry whose TTL is 0 is
  not cached.
- Local domain: A query → its IPv4 with the cache TTL; any other type →
  empty NOERROR answer.
- PTR `a.b.c.d.in-addr.arpa` for a private IP: only the first local domain
  in map order is checked; a match answers its name, otherwise NXDOMAIN
  without asking upstream (bug). With no local domains the query goes
  upstream. A trailing `.` is not stripped (typo: the second TrimSuffix
  uses the original name).
- Upstreams: every upstream (IPv4 only, port 53, UDP) is asked at once with
  a new random ID and RD set; dial timeout 2 s, read timeout 2 s. The first
  answer with records wins and is cached with the TTL of its first answer
  (first authority when there are no answers). The whole fan-out has one
  upstream deadline (5 s) from the moment the queries go out; replies after
  it are dropped. All failed → the last error to arrive, or `ErrNxDomain`
  when none returned records or the deadline passed (bug: a timeout
  answers NXDOMAIN, not SERVFAIL).
- Negative caching: when the result is `ErrNxDomain` and an upstream's
  NXDOMAIN carried an SOA in its authority section, Resolve returns that
  SOA as the authority with the error, and caches the NXDOMAIN for
  min(SOA TTL, SOA minimum, 60 s). A hit returns `ErrNxDomain` with the
  SOA. An NXDOMAIN without an SOA is not cached.
- The cache is checked before local domains, so a name cached before it
  became a local domain keeps its cached result until the entry expires.
- Upstream RCODE 1 → `ErrDNSFormatError`, 3 → `ErrNxDomain`, 2 →
  `ErrDNSServerFailure`.
- Defaults: cache TTL 5 min, upstreams 1.1.1.1 and 9.9.9.9.
- Locking: the blocklist and local domains sit behind a read-write lock
  (reads for lookups, writes for changes); the cache has its own. No lock
  is held across an upstream lookup, so queries resolve in parallel and a
  cached name answers while other lookups are in flight. Parallel misses
  for the same name each go upstream (no coalescing).
- Local domains reject `0.0.0.0` and non-IPv4 addresses.
- Metrics: `queryCounter` per domain, upstream (or `cache`,
  `local-domain`) and result. NXDOMAIN is not counted, from an upstream or the
  cache.

# Interfaces
- `Resolver` interface: Resolve, Add/DeleteLocalDomain,
  AddBlocklistEntries, DeleteBlocklistEntry, FlushBlocklist.
- [Wire format](/dns/wire-format.md).

# Decisions
- Fixes agreed 2026-10-17 (plans): the PTR bug, folded together with the
  `feature/ptr_records` branch.
- Cache limit 10 000 entries, nearest expiry evicted first: bounded memory
  on atlas, and the entry that is evicted would have been the next to
  expire anyway (user, 2026-10-18).
- NXDOMAIN cached for at most 60 s: repeated lookups of a missing name stop
  going upstream, and a name that starts to exist shows up within a minute
  (user, 2026-10-18).
