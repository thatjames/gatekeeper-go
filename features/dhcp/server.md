---
type: Feature
title: DHCP server
description: An optional DHCPv4 server on UDP 67 that offers, acknowledges and releases leases from the pool with the configured gateway, mask, DNS servers and domain; off at home, where hecate serves DHCP.
status: stable
code:
  - internal/dhcp/server.go
  - internal/dhcp/server_test.go
  - internal/dhcp/dhcp_feature_test.go
  - internal/dhcp/features/lease_offering.feature
tags: [dhcp]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T19:30:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
DHCP for networks where gatekeeper is the DHCP server. Kept as a supported
optional module (confirmed 2026-10-17), not used at home: hecate serves DHCP.

# Behavior
- `Start()` needs the configured interface to exist and support broadcast;
  it takes the interface's first IP network as server address and computes
  its broadcast address. It adds the configured reservations (copied under
  the [config lock](/config/config-file.md), since the API changes that
  map), then loads the lease file (a rejected file is a warning and the
  pool starts empty), listens on UDP4 `:67` on all interfaces, starts one
  listener with 10 receive and 10 response workers, and starts the lease
  writer (`PersistOnChange`, 1 s debounce).
- The lease file path is `LeaseFile`, or `/var/lib/gatekeeper/leases` when
  it is empty. It is resolved when the server is made and again in
  `UpdateOptions()`; load, writer and `Stop()` all use it.
- Only BOOTREQUEST packets of at least 240 bytes are handled. A packet with
  no option 53 panics the worker (bug).
- DISCOVER: the client's existing lease (reserved or active) or the next free
  one is offered; an empty pool sends nothing. The OFFER carries server ID,
  lease time, router, subnet mask, DNS servers and domain name.
- REQUEST: ignored when it names another server ID. The requested address
  (option 50, else ciaddr) must equal the client's lease: then ACK (offered →
  active, active → renewed, reserved → stays reserved), else NAK. A client
  with no lease gets NAK. The hostname option is stored on the lease
  (`SetHostname`, under the pool lock).
- RELEASE: frees the client's pool lease. A reserved client keeps its
  reservation (pool and config file untouched), so it gets the same IP when
  it rediscovers.
- DECLINE and INFORM are not handled.
- Replies go to the sender; when the sender is `0.0.0.0` they go to the
  subnet broadcast address on the same port.
- `Stop()` stops the lease writer and waits for it, closes the listener and
  saves the lease file once more. Workers are not stopped, so each restart
  adds 20 more (leak).
- A lease is on disk about a second after the change that made it, so a
  crash or SIGKILL loses at most that second.
- `UpdateOptions()` swaps the options; the pool keeps its old range until
  the process restarts (bug).
- Metrics: `dhcp_req_time` (ms), `dhcp_op_counter` by op, client MAC and
  hostname, `dhcp_active_lease_count` (drifts: RELEASE decrements before
  checking the lease, reservations count as active).

# Interfaces
- [Lease pool](/dhcp/lease-pool.md), [packet format](/dhcp/packet.md),
  [DHCP API](/web/dhcp-api.md).

# Decisions
- RELEASE keeps a reservation (2026-10-17, user): many devices release at
  shutdown; removing a reservation is an admin action in the API.
- Fixes agreed 2026-10-17 (plans): live option changes, worker leaks, binding to the interface, DECLINE and
  INFORM, the option 53 panic and the gauge.
