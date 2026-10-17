---
type: Feature
title: DHCP packet format
description: Reads and writes BOOTP/DHCP messages in place over a byte slice - fixed fields, magic cookie, options - and builds replies.
status: stable
code:
  - internal/dhcp/packet.go
  - internal/dhcp/packet_test.go
  - internal/dhcp/.test-req
tags: [dhcp, protocol]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
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
- `.test-req` is a captured DHCP request used by `packet_test.go`.

# Decisions
- `.test-req` carries MAC `00:d8:61:39:b5:6a`, which also appears in the
  sample config; it is to be replaced by a synthetic packet (agreed
  2026-10-17).
