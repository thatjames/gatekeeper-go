---
type: Feature
title: DNS wire format
description: Parses and marshals DNS messages (header, one question, answers, authorities, additionals, name compression, OPT) for the server and resolver.
status: stable
code:
  - internal/dns/types.go
  - internal/dns/types_test.go
  - internal/dns/features/dns_packet.feature
tags: [dns, protocol]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T13:00:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
A hand-written RFC 1035 codec, part of the point of the project.

# Behavior
- `ParseDNSMessage`: under 12 bytes → `ErrDNSPacketTooShort`; more than one
  question → `ErrDNSTooManyQuestions` (the server then drops the query; the
  code comment's FORMERR is not sent); zero questions parse fine.
- Names: labels up to 63 bytes, compression pointers followed and checked
  against the packet length; parse errors name the offset.
- OPT (EDNS) records are parsed with their own RDATA bounds check; the
  UDP size lands in `Class` and the extended RCODE, version and flags in
  `TTL`. An OPT record with no `Name` is marshaled with the root name
  (one zero byte), so parsed OPT records marshal back unchanged.
  `NewOPTRecord(udpSize)` builds one with version 0 and no options.
- `MarshalDNSMessage` writes the header counts from the slices; answers use
  the `0xc00c` pointer to the question name.
- Header helpers set QR, opcode, AA, TC, RD, RA, Z and RCODE, and read TC
  and RCODE.
- Named types: A, NS, CNAME, SOA, PTR, MX, TXT, AAAA, OPT; others pass
  through as numbers.
