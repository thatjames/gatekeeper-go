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
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-17T21:52:15Z }
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
  offered to the client for 60 s. Reserved addresses inside the range are
  not skipped (bug: two clients can get one IP).
- `AcceptLease` makes the client's pool lease active for the TTL (no-op for
  reservations).
- `ReserveLease(client, ip)` clears the client's pool lease and refuses an IP
  already reserved for another client.
- `UpdateLease(client, ip)` changes a reservation (then writes the config
  file) or a pool lease's IP.
- `ReleaseLease`: for a reservation, deletes it and its config entry and
  writes the config; otherwise clears the pool lease.
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
- Fixes agreed 2026-10-17 (plans): skip reserved IPs; persist on change,
  keep real expiry, no 255 limit, safe load.
