---
type: Feature
title: DNS blocklists
description: Blocklists in hosts-file format are fetched from HTTP(S) URLs or read from files, in parallel, and every hostname in them is blocked by the resolver.
status: stable
code:
  - internal/dns/blocklist.go
  - internal/dns/blocklist_test.go
  - internal/util/hosts.go
tags: [dns, blocklist]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Pi-hole-style ad and tracker blocking from public lists such as StevenBlack.

# Behavior
- A source starting with `http` is fetched over HTTP (15 s timeout); any
  other source is read as a local file. Non-200 is an error.
- Hosts format: blank and `#` lines are skipped; every other line needs at
  least two fields, and only the second field is taken. One bad line rejects
  the whole list (`ErrInvalidBlocklistFormat`). Inline `#` comments after
  the hostname are fine (ignored as extra fields).
- Header entries such as `localhost`, `broadcasthost` and `0.0.0.0` are
  blocked like any other name (bug).
- All sources load in parallel at DNS start; a failing source is logged and
  skipped; the rest are added in one call.
- The developer sample `config/config.yml` has a trailing space after the
  StevenBlack URL.

# Decisions
- `FileBlocklistFetcher` is dead (the HTTP fetcher already reads files):
  confirmed 2026-10-17, removed by the dead code plan.
- Fix agreed 2026-10-17: skip hosts-file header entries, decide subdomain
  matching (blocklist plan).
