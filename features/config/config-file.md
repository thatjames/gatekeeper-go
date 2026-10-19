---
type: Feature
title: Config file
description: One YAML file (DHCP, DNS, Web, Auth sections) is loaded into a global at start and written back whole when the web API changes settings.
status: stable
code:
  - internal/config/**
  - config/config.yml
  - config/docker-config.yml
tags: [config, yaml]
generated: { by: claude-code/claude-opus-5-5, at: 2026-10-18T21:59:00Z }
verified: { by: human:slimjim, at: 2026-10-17T21:59:57Z }
---

# Intent
A single human-editable file holds all settings, and the UI can change them
without hand edits.

# Behavior
- `LoadConfig(path)` decodes the YAML into `config.Config` and remembers the
  path. Each top-level section is optional; a missing section turns its
  module off.
- `Auth.AuthType` picks the auth block: `oidc` (IssuerURL, ClientID,
  ClientSecretVar, RedirectURL, Scopes) or `htpasswd` (HTPasswdFile). Any
  other type fails the load. No `Auth:` leaves it nil. `Auth.HTPasswdFile`
  is parsed but never read: password login uses `Web.HTPasswdFile`
  ([auth](/web/auth.md)); the duplicate key is on the backlog.
- `DHCP`: Interface, StartAddr, EndAddr, DomainName, NameServers, LeaseTTL
  (seconds), SubnetMask, Gateway, ReservedAddresses (MAC → IP), LeaseFile.
- `DNS`: UpstreamServers, Interface, LocalDomains (name → IPv4), Port,
  BlockLists (URLs or file paths), BlockedDomains, ConditionalForwarders
  (list of `Zone`, `Server`: IPv4 with optional `:port`;
  [resolver](/dns/resolver.md)).
- `Web`: Address, WebURL, AllowedOrigins (extra CORS origins), TLS
  (PublicKey, PrivateKey paths), HTPasswdFile, Prometheus. `Prometheus` (default false) turns on `/metrics`
  ([HTTP server](/web/http-server.md)).
- `UpdateConfig()` truncates the loaded file and writes the whole global back
  (2-space indent); comments and key order in the file are lost. With no
  file loaded it returns an error.
- One `sync.RWMutex` guards the global after load: `config.Lock()` /
  `Unlock()` around a change and its `UpdateConfig()` (the caller must hold
  `Lock`), `config.RLock()` / `RUnlock()` to read a section the API can
  change (DHCP, DNS). Web and Auth are only set at load and read without it.
- `String()` prints the config as an indented tree for debug logs; nil
  sections are skipped.
- `config/config.yml` is the developer sample (OIDC, DNS only);
  `config/docker-config.yml` is the sample mounted by `make docker-run`.

# Decisions
- Secrets are not in the file: `ClientSecretVar` names an environment
  variable (`GATEKEEPER_OIDC_CLIENT_SECRET` in the sample).
- The sample's `IssuerURL` points at `keycloak.thatjames.nl`, which is stale:
  the identity provider is Authentik on atlas ([OIDC login](/web/auth.md)).
- The config format is not a hard constraint; it may change.
- One global lock rather than one per section (2026-10-18, web handler
  plan): `UpdateConfig` writes every section, so any save must exclude every
  change.
