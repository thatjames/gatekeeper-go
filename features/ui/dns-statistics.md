---
type: Feature
title: UI DNS statistics
description: Charts of DNS metrics read from /metrics every 30 s - request time, queries by upstream and result, top blocked domains, queries by client IP.
status: stable
code:
  - internal/web/ui/src/scenes/dns/DNSStatistics.svelte
  - internal/web/ui/src/lib/prometheus/**
  - internal/web/ui/src/components/ApexChart.svelte
tags: [ui, dns, metrics]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:12:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- Fetches the Prometheus text format from `/metrics`, sending the Bearer
  token when there is one and cookies always, and parses histograms and
  counter vectors itself; refreshes every 30 s.
- When `/metrics` answers an error (404 with `Web.Prometheus` off, 401 or
  403) the page shows "Error loading DNS metrics" with the status text.
- Charts `dns_req_time`, `dns_query_count`, `dns_blocked_domain_count` and
  `dns_query_by_ip_count`. The Cache Hit Rate card is the share of
  `dns_query_count` with `upstream="cache"`.

# Decisions
- `/metrics` is behind auth (2026-10-18), so the page sends the same
  credentials as the API client.
- No `dns_cache_hit_count` (2026-10-24): the page asked for it but never
  drew it; cache hits are already counted in `dns_query_count`.
