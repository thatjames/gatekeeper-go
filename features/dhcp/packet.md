---
type: Feature
title: DHCP packet format
description: Reads and writes BOOTP/DHCP messages in place over a byte slice - fixed fields, magic cookie, options - and builds replies.
status: stable
code:
  - internal/dhcp/packet.go
  - internal/dhcp/packet_test.go
tags: [dhcp, protocol]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T17:35:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
A hand-written RFC 2131/2132 codec, part of the point of the project.

# Behavior
- `Message` is the raw packet; getters and setters read op, htype, hlen,
  xid, secs, flags, ciaddr, yiaddr, siaddr, giaddr, chaddr, sname, file,
  cookie and options at their fixed offsets.
- `ParseOptions`: pad skipped, end stops, subnet mask and time offset read
  as 4 bytes, every other option as tag-length-value.
- `MarshalOptions` writes tag-length-value in map order and an end option.
- `MakeReply` copies xid, flags, giaddr and chaddr from the request, sets
  yiaddr, lease time and message type, and appends the options.
- `packet_test.go` parses a synthetic DHCPREQUEST it builds itself (MAC
  `02:00:00:00:00:01`, requested IP 10.0.0.40, seven requested parameters).

# Decisions
- The captured `.test-req`, which carried a real device's MAC, was
  replaced by the synthetic packet built in the test (2026-10-24): no real
  device data in the repository.
