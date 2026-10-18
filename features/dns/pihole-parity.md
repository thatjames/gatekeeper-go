---
type: Feature
title: Pi-hole parity on atlas
description: What gatekeeper's DNS does so it can replace Pi-hole on atlas - TCP and UDP on port 53, truncation and EDNS, conditional forwarding of reverse and local lookups to hecate, interface binding.
status: stable
code:
  - internal/dns/features/dns_parity.feature
  - internal/dns/parity_feature_test.go
tags: [dns, atlas, pihole]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T14:00:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
gatekeeper replaces the Pi-hole service on atlas (confirmed 2026-10-17).
Hecate's dnsmasq forwards LAN queries to it; DHCP stays on hecate. The
switch must change nothing for the LAN.

# Behavior
- **TCP and UDP.** The server listens on the same port over UDP and TCP
  and both hand queries to the same resolver. TCP messages carry the
  2-byte length prefix (RFC 1035 4.2.2); a connection may send several
  queries and is closed after 10 s idle.
- **Truncation.** A UDP reply bigger than the client's limit (512 bytes,
  or the client's EDNS UDP size, capped at 1232) is sent with TC set, the
  question and OPT kept and all answer and authority records removed; the
  client then asks over TCP and gets the full reply. TCP replies are never
  truncated.
- **EDNS.** A query with an OPT record gets an OPT record back (UDP size
  1232, version 0, no options); other additional records are not returned.
  A query without OPT gets none.
- **Large upstream answers.** Upstream and forwarder queries advertise EDNS
  with UDP size 1232. A reply with TC set is asked again over TCP from the
  same server.
- **Conditional forwarding.** `DNS.ConditionalForwarders` lists zones and
  the server that answers for them. A query for a zone, or for a name under
  it, goes only to that server: the longest matching zone wins, matching
  ignores case and a trailing dot. When the server fails or does not
  answer, the reply is SERVFAIL; the normal upstreams are never asked, so
  LAN names do not leave the LAN. Forwarded answers are cached like
  upstream ones. Blocklist, cache and local domains (and a local-domain
  PTR match) still answer first.
- **Interface binding.** When `DNS.Interface` is set, UDP and TCP listen
  on that interface's first IPv4 address only; a missing interface or one
  without an IPv4 address fails `Start`. Empty: all addresses.
- Upstreams as configured (Pi-hole on atlas uses Quad9).

# Interfaces
Config for atlas (zones written out, not derived from subnets):

```yaml
DNS:
  Interface: <atlas LAN interface>
  UpstreamServers: [<the Quad9 addresses Pi-hole uses>]
  ConditionalForwarders:
    - Zone: styx
      Server: 10.20.1.1
    - Zone: 1.20.10.in-addr.arpa
      Server: 10.20.1.1
```

`Server` is an IPv4 address, optionally with `:port` (default 53).
`Interface` names an interface as gatekeeper sees it: with Docker's bridge
network that is the container's `eth0`, so binding to atlas's LAN NIC needs
host networking (or leave `Interface` empty).
The godog scenarios in `dns_parity.feature` run a real server on
127.0.0.1 against fake forwarders. Built in [DNS server](/dns/server.md), [DNS resolver](/dns/resolver.md),
[DNS wire format](/dns/wire-format.md) and [Config file](/config/config-file.md).

# Decisions
- Config shape: a list of `{Zone, Server}` with zones written out, not a
  domain plus subnets (user, 2026-10-18).
- Interface binding by address, not `SO_BINDTODEVICE`: portable and needs
  no extra capability in the container (user, 2026-10-18).
- A failed forwarder gives SERVFAIL with no fallback to upstreams, as
  dnsmasq's `server=/zone/` does (user, 2026-10-18).
- Upstream queries use EDNS 1232 plus a TCP retry on TC, so large answers
  reach the client in full (user, 2026-10-18). 1232 is the DNS Flag Day
  2020 size that avoids IP fragmentation.
- *open*: which image atlas pulls (GHCR release or registry.thatjames.nl) and
  where the compose entry lives: not decided (2026-10-17).
- The local domain and subnet follow hecate's config at the time of the
  switch; check them then.
