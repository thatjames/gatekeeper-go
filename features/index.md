---
okf_version: "0.2"
---

# Runtime

* [Startup and shutdown](/runtime/startup.md) - The gatekeeper binary loads the YAML config, sets up logging, registers the DHCP, DNS and web modules that are configured, starts them and stops them on SIGINT/SIGTERM.
* [Service manager](/runtime/service-manager.md) - A process-wide registry of the running modules (dhcp, dns, web keys) that starts and stops them and lets the web API fetch a module by key.
* [Router mode](/runtime/router-mode.md) - draft - gatekeeper as a full router on Linux through netlink; work in progress on the feature/router branch, a long-term goal not used at home.

# Config

* [Config file](/config/config-file.md) - One YAML file (DHCP, DNS, Web, Auth sections) is loaded into a global at start and written back whole when the web API changes settings.

# DNS

* [DNS server](/dns/server.md) - A UDP and TCP DNS server on the configured port that hands each query to the resolver and answers with its records, and can be reconfigured from the API.
* [DNS resolver](/dns/resolver.md) - Resolves a name and type from the blocklist, a TTL cache, local domains, private PTR records or all upstreams in parallel, first answer wins.
* [DNS blocklists](/dns/blocklist.md) - Blocklists in hosts-file format are fetched from HTTP(S) URLs or read from files, in parallel, and every hostname in them is blocked by the resolver.
* [DNS wire format](/dns/wire-format.md) - Parses and marshals DNS messages (header, one question, answers, authorities, additionals, name compression, OPT) for the server and resolver.
* [Pi-hole parity on atlas](/dns/pihole-parity.md) - What gatekeeper's DNS does so it can replace Pi-hole on atlas - TCP and UDP on port 53, truncation and EDNS, conditional forwarding of reverse and local lookups to hecate, interface binding.

# DHCP

* [DHCP server](/dhcp/server.md) - An optional DHCPv4 server on UDP 67 that offers, acknowledges and releases leases from the pool with the configured gateway, mask, DNS servers and domain; off at home, where hecate serves DHCP.
* [DHCP lease pool](/dhcp/lease-pool.md) - The pool of addresses from StartAddr to EndAddr with available, offered, active and reserved leases, MAC reservations, and a binary lease file that survives restarts.
* [DHCP packet format](/dhcp/packet.md) - Reads and writes BOOTP/DHCP messages in place over a byte slice - fixed fields, magic cookie, options - and builds replies.

# Web API

* [HTTP server](/web/http-server.md) - A gin server on Web.Address (default :8085, optional TLS) that serves the API under /api, authenticated Prometheus metrics at /metrics when Web.Prometheus is true, and the embedded UI for every other path, with CORS for WebURL and Web.AllowedOrigins.
* [API v1 routes](/web/api-routes.md) - The /api/v1 route table - open health, version, features and login routes; DHCP, DNS and system routes behind auth, registered only for running modules - with request types and validation.
* [Login and tokens](/web/auth.md) - Users log in with htpasswd credentials or OIDC (Authentik) and get a 7-day HS256 JWT signed with GATEKEEPER_JWT_SECRET, sent as a Bearer header or the oauth_token cookie.
* [OIDC roles](/web/auth-roles.md) - draft - map the OIDC roles claim to gatekeeper roles (admin, user and more) and allow API actions by role, so not every Authentik user can change DNS.
* [DHCP API](/web/dhcp-api.md) - Authenticated /api/v1/dhcp routes to list, release, reserve and change leases and to read and change DHCP options, writing changes back to the config file.
* [DNS API](/web/dns-api.md) - Authenticated /api/v1/dns routes to read and change upstreams and interface, manage local domains, blocklist URLs and single blocked domains, writing changes back to the config file.
* [System API](/web/system-api.md) - /api/v1/version and authenticated /api/v1/system routes for host info, network interfaces and active modules; /features names extra features such as oidc.
* [Unused OIDC config helper](/web/oidc-config-helper.md) - deprecated - NewOIDCConfig and OIDCOptions in auth.go duplicate initOIDC and are never called; to be deleted.

# UI

* [UI shell and login](/ui/app-shell.md) - The Svelte 5 single-page admin UI - hash routes, side menu built from the running modules, home screen with system info, login by password or OIDC, error screen - and its Vite build into internal/web/ui/dist.
* [UI DHCP pages](/ui/dhcp-pages.md) - Lease pages (active and reserved tables, details, reserve, edit, release) and the DHCP settings view and form.
* [UI DNS pages](/ui/dns-pages.md) - DNS settings (upstreams, interface, blocklist URLs, blocked domains) and local domain pages.
* [UI DNS statistics](/ui/dns-statistics.md) - Charts of DNS metrics read from /metrics every 30 s - request time, queries by upstream and result, top blocked domains, queries by client IP.
* [UI settings page](/ui/settings-page.md) - draft - a settings page in the UI, with backing API and web server changes, in progress on the feature/settings-page branch.

# System

* [System info](/system/system-info.md) - Linux host facts for the home screen - hostname, uptime, memory, and Tx/Rx of the DHCP interface and ppp0 - plus the network interface list.
* [Byte size helper](/system/byte-size.md) - util.ByteSize formats a byte count as B, K, M or G; tested but not called by any code (system info has its own copy).
* [Non-Linux stubs](/system/non-linux-stubs.md) - deprecated - darwin and windows stubs of the system package that return "not supported"; gatekeeper is Linux only, the darwin build fails, to be deleted.

# Build and release

* [Make targets and Go module](/build/makefile.md) - The Makefile builds the UI and the Go binary (native, pipeline, Docker, multi-arch), runs the tests and test reports, and runs locally; go.mod pins Go 1.24.6 and the dependencies.
* [Docker image](/build/docker-image.md) - Alpine images that run /app/gatekeeper with prebuilt binaries - a single-arch image and a multi-arch one picking the binary by TARGETPLATFORM.
* [GitHub Actions CI and release](/build/ci.md) - On push and PR, CI builds the UI and runs the tests with coverage to Codecov; a pushed tag builds five Linux binaries, pushes the multi-arch GHCR image with provenance and creates a GitHub release.
* [GitHub repository templates](/build/github-meta.md) - CODEOWNERS (@thatjames) and the bug report, feature request and pull request templates of the GitHub repository.

# Docs

* [Documentation site](/docs/site.md) - Hugo docs with the hugo-book theme (submodule), published to GitHub Pages on pushes to main that touch docs/ - getting started, configuration, auth, modules, monitoring, web pages; plus the README.

# Deployment

* [Atlas deployment](/deploy/atlas.md) - draft - gatekeeper runs on atlas as the LAN DNS server in place of Pi-hole, DNS only (hecate serves DHCP), with OIDC through Authentik; not deployed yet.
