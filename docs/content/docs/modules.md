---
title: 'Modules'
weight: 2
---

# GateKeeper Modules

GateKeeper is built up of a number of modules, each of which is responsible for a specific task. Some modules can be enabled/disabled using config keys in YAML, but the Router module requires the use of a certain binary/docker-image.

This is because the Router module is heavily reliant on the Netlink API, which is only available on Linux. Separating the Router module into a separate binary/docker-image allows for the use of GateKeeper on other platforms as a DHCP/DNS server while also making it easier to maintain the router module without impacting other features.

## DNS

The DNS module is a simple recursive DNS server that can be used to resolve DNS locally. It includes a local domain list that can be used to resolve local DNS entries (such as router, mediapc, etc) or just a way to override DNS entries as this local list is considered before a recursive request is made upstream.

### Configuration

The DNS module is loaded if the config file contains the `DNS` key.

Below are the list of configuration options for the DNS module.

| Key             | Description                                                                | Default          |
| --------------- | -------------------------------------------------------------------------- | ---------------- |
| Interface       | The interface the UDP socket will bind to                                  | eth0             |
| LocalDomains    | A map of DNS names to IP addresses                                         |                  |
| UpstreamServers | The upstream DNS servers to use                                            | 8.8.8.8, 1.1.1.1 |
| Blocklists      | A list of host file formated files that will be used to block DNS requests |                  |
| BlockedDomains  | A list of domains to outright block                                        |                  |

## DHCP

The DHCP module is a simple DHCP server that can be used to provide DHCP to devices on the network.

### Configuration

The DHCP module is loaded if the config file contains the `DHCP` key.

| Key               | Description                                                                                                   | Default                    |
| ----------------- | ------------------------------------------------------------------------------------------------------------- | -------------------------- |
| Interface         | Network interface for DHCP server                                                                             | eth0                       |
| StartAddr         | Starting IP address for DHCP pool                                                                             | 10.0.0.2                   |
| EndAddr           | Ending IP address for DHCP pool                                                                               | 10.0.0.99                  |
| DomainName        | Domain name provided to DHCP clients                                                                          |                            |
| NameServers       | DNS servers provided to clients. If you have the DNS module enabled, you should include this device's IP here |                            |
| LeaseTTL          | DHCP lease time-to-live in seconds                                                                            | 300                        |
| SubnetMask        | Subnet mask for the network                                                                                   | 255.255.255.0              |
| Gateway           | Default gateway IP address                                                                                    | 10.0.0.1                   |
| ReservedAddresses | MAC to IP address mappings for static reservations                                                            |                            |
| LeaseFile         | Path to file storing DHCP lease information                                                                   | /var/lib/gatekeeper/leases |

## Web

The Web module is a simple HTTP server that can be used to provide a web interface for GateKeeper.

### Configuration

The Web module is loaded if the config file contains the `Web` key.

| Key          | Description                                                                                                                             | Default |
| ------------ | --------------------------------------------------------------------------------------------------------------------------------------- | ------- |
| Address      | Address and port for web server to listen on                                                                                            | :8085   |
| TLS          | TLS configuration for HTTPS                                                                                                             |         |
| HTPasswdFile | Path to htpasswd file for authentication. If this is not included, the module defaults to username:password admin:admin                 |         |
| Prometheus   | Enable Prometheus metrics endpoint. This is the same endpoint the DNS Statistics page uses, so disabling it here will disable that page | false   |
