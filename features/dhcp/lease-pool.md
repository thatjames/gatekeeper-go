---
type: Feature
title: DHCP lease pool
description: The pool of addresses from StartAddr to EndAddr with available, offered, active and reserved leases, MAC reservations, and a binary lease file that survives restarts.
status: stable
code:
  - internal/dhcp/lease.go
  - internal/dhcp/lease_test.go
  - internal/dhcp/leasefile.go
  - internal/dhcp/leasefile_test.go
  - internal/dhcp/features/lease_expiry.feature
  - internal/dhcp/features/lease_initialization.feature
  - internal/dhcp/features/reserved_leases.feature
tags: [dhcp, leases]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-24T19:30:00Z }
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
  lease's IP. Like `ReserveLease`, it refuses to move a reservation onto an
  IP reserved for another client and then changes nothing.
- `ReleaseLease` (the client's DHCPRELEASE): a reserved client keeps its
  reservation, in the pool and in the config file, and nothing is written;
  otherwise the client's pool lease is cleared.
- `DeleteLease` (the admin's DELETE in the [DHCP API](/web/dhcp-api.md)):
  for a reservation, deletes it and its config entry and writes the config;
  otherwise clears the pool lease.
- Config changes and writes from the pool hold the
  [config lock](/config/config-file.md), taken inside the pool lock.
- `ActiveLeases()` returns unexpired active leases; `ReservedLeases()` all
  reservations.
- `SetHostname(lease, name)` stores a client's hostname under the pool lock.
- Every change that matters to the lease file (`AcceptLease`,
  `ReleaseLease`, `DeleteLease`, `ReserveLease`, `UpdateLease`, a new
  hostname) signals the writer; offers do not. `PersistOnChange(file,
  debounce, stop)` saves at once on the first signal, then waits `debounce`
  (1 s in the server) and saves once more if anything changed meanwhile, so
  a burst of changes costs at most two writes; it returns when `stop`
  closes.
- Lease file, big-endian: magic `GKLF`, version byte `2`, uint32 record
  count; each record is uint16 length + client ID, uint16 length +
  hostname, 4-byte IPv4, state byte (`3` active, `2` reserved), int64 Unix
  expiry in seconds (`0` for none). Any number of leases fit. Unexpired
  active leases and all reservations are written.
- `SaveLeases(file)` creates the directory (0755), writes a temporary file
  beside it, syncs it and renames it over the old one (mode 0600): a crash
  leaves the old file or the new one, never half of one.
- `LoadLeases(file)`: a missing or empty file loads nothing. Every length
  is checked against the data; a wrong magic or version (the old
  one-count-byte format too), a record past the end, an unknown state or
  trailing bytes reject the file: it is renamed `<file>.bad-<UTC time,
  20061102T150405Z>`, an error naming the new path is returned and the
  pool is left as it was. Otherwise active leases keep their saved expiry
  and replace the pool entry with their IP; expired ones, ones outside the
  range, and ones on a reserved IP or for a reserved client are dropped. A
  saved reservation only gives its hostname back to a reservation the pool
  already has for that client on the same IP; the config stays the only
  source of reservations.

# Decisions
- The lease file format is not a hard constraint (2026-10-17); the lease
  file plan may change it.
- Release and delete are split (2026-10-17, user): a device's RELEASE at
  shutdown must not cost it its reservation; only an admin removes one.
- Reserved IPs never reach another client (2026-10-17, user): the pool skips
  them and a new reservation evicts a client already on that IP.
- Lease file format v2 (2026-10-18, user): versioned binary with magic
  `GKLF`, a 32-bit count and real expiry times; an old-format file is moved
  aside, not migrated. Reservations are saved but restore only hostnames,
  so the config remains their one source. 1 s debounce; a rejected file is
  kept as `<file>.bad-<time>` for inspection.
