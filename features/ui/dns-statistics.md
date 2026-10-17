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
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Behavior
- Fetches the Prometheus text format from `/metrics` and parses histograms
  and counter vectors itself; refreshes every 30 s.
- Charts `dns_req_time`, `dns_query_count`, `dns_blocked_domain_count` and
  `dns_query_by_ip_count`. It also asks for `dns_cache_hit_count`, which the
  server does not export, so that chart stays empty.

# Decisions
- Depends on `/metrics` being readable from the browser: putting `/metrics`
  behind auth (auth plan) must keep this page working.
