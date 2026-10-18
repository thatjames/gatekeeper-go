---
type: Feature
title: DHCP lease pool
description: The pool of addresses from StartAddr to EndAddr with available, offered, active and reserved leases, MAC reservations, and a binary lease file that survives restarts.
status: stable
code:
  - internal/dhcp/lease.go
  - internal/dhcp/lease_test.go
  - internal/dhcp/features/lease_expiry.feature
  - internal/dhcp/features/lease_initialization.feature
  - internal/dhcp/features/reserved_leases.feature
tags: [dhcp, leases]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T22:10:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
Hand out each address to one client at a time, keep fixed addresses for
known devices, and remember leases across restarts.

# Behavior
- The pool holds one lease per address from start to end inclusive.
  Reservations (client ID → IP) live apart from it and may lie inside or
  outside the range.
- `GetLease(client)`: the reservation if any; else the pool lease with that
  client ID (case-insensitive), which is cleared and not returned once
  expired.
- `NextAvailableLease(client)`: the first lease that is available, offered to
  this client, an expired offer, or an expired active lease; it becomes
  offered to the client for 60 s. An address reserved for any client is
  skipped, so a full pool with a reserved IP in range offers nothing for it.
- `AcceptLease` makes the client's pool lease active for the TTL (no-op for
  reservations).
- `ReserveLease(client, ip)` refuses an IP already reserved for another
  client and then changes nothing. Otherwise it clears the client's own pool
  lease and any other client's pool lease on that IP (that client is NAKed at
  its next REQUEST and rediscovers), then stores the reservation.
- `UpdateLease(client, ip)` changes a reservation (clearing any other
  client's pool lease on the new IP, then writing the config file) or a pool
  lease's IP.
- `ReleaseLease` (the client's DHCPRELEASE): a reserved client keeps its
  reservation, in the pool and in the config file, and nothing is written;
  otherwise the client's pool lease is cleared.
- `DeleteLease` (the admin's DELETE in the [DHCP API](/web/dhcp-api.md)):
  for a reservation, deletes it and its config entry and writes the config;
  otherwise clears the pool lease.
- `ActiveLeases()` returns unexpired active leases; `ReservedLeases()` all
  reservations.
- Lease file: one count byte, then per active lease
  `len,clientID,len,hostname,IPv4(4),state`. Written only by `Stop()`;
  more than 255 leases overflow the count. On load each lease gets a fresh
  full TTL and replaces the pool entry with the same IP; a truncated file
  panics.

# Decisions
- The lease file format is not a hard constraint (2026-10-17); the lease
  file plan may change it.
- Release and delete are split (2026-10-17, user): a device's RELEASE at
  shutdown must not cost it its reservation; only an admin removes one.
- Reserved IPs never reach another client (2026-10-17, user): the pool skips
  them and a new reservation evicts a client already on that IP.
- Fixes agreed 2026-10-17 (plans): persist on change, keep real expiry, no
  255 limit, safe load.
