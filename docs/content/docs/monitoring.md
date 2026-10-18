---
title: 'Monitoring'
weight: 3
---

# Monitoring

GateKeeper includes a Prometheus metrics endpoint that can be used to monitor the performance of the system.

This is available from the `/metrics` endpoint on the web server, provided the `prometheus` property is set to `true` in the configuration file's web module.

## Authentication

The endpoint needs a login token, because the metrics include client IP addresses, hostnames and queried domains. Without one it answers `401`.

A scraper sends a token as a bearer token. Get one by logging in:

```bash
curl -s -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"<password>"}' \
  http://gatekeeper:8085/api/v1/login
```

and configure Prometheus with it:

```yaml
scrape_configs:
  - job_name: gatekeeper
    authorization:
      credentials: <token>
    static_configs:
      - targets: ['gatekeeper:8085']
```

Tokens last 7 days, so a long-running scraper needs a fresh token each week.

## Available Metrics

Below is a list of the available metrics that can be scraped from the Prometheus endpoint.

| Metric Category | Metric                   | Type      | Description                                                                                          |
| --------------- | ------------------------ | --------- | ---------------------------------------------------------------------------------------------------- |
| **DHCP**        | dhcp_active_lease_count  | gauge     | Count of currently active DHCP leases                                                                |
|                 | dhcp_req_time            | histogram | DHCP request processing time in milliseconds                                                         |
| **DNS**         | dns_blocked_domain_count | counter   | DNS queries blocked per domain                                                                       |
|                 | dns_query_by_ip_count    | counter   | DNS queries grouped by source IP address and result status                                           |
|                 | dns_query_count          | counter   | DNS queries by domain, result status, and upstream server (cache, 1.1.1.1, 9.9.9.9, or local-domain) |
|                 | dns_req_time             | histogram | DNS request processing time in milliseconds, tracking latency distribution                           |
