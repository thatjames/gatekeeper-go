# GateKeeper

GateKeeper is a configurable DHCP and DNS server written in Go

## Installation

Install gatekeeper from source with `go install gitlab.com/thatjames-go/gatekeeper-go@latest` or download one of the binaries under the [releases](https://gitlab.com/thatjames-go/gatekeeper-go/-/releases) page.


## Running

The binary will look for a `config.yml` file in the PWD, or it can be passed with the `-c <file>` flag.

## Configuration

Example config:

```yaml
DHCP:
  DomainName: GateKeeper
  Interface: eth0
  StartAddr: 10.0.0.2
  EndAddr: 10.0.0.99
  LeaseTTL: 86400
  SubnetMask: 255.255.255.0
  Router: 10.0.0.1
  NameServers:
    - 8.8.8.8
    - 1.1.1.1
  ReservedAddresses:
    bc:5f:f4:ac:ea:c4: 10.0.0.100
```
