# GateKeeper

GateKeeper is a configurable DHCP server written in Go

## Installation

Install gatekeeper from source with `go install gitlab.com/thatjames-go/gatekeeper-go@latest` or download one of the binaries under the [releases](https://gitlab.com/thatjames-go/gatekeeper-go/-/releases) page.


## Running

The binary will look for a `config.yml` file in the PWD, or it can be passed with the `-c <file>` flag.


### Webserver

If the webserver is enabled, but no `htpasswd` file has been specified, then the default login credentials are admin/admin.

In order to secure the webserver, you can create an `htpasswd` file and specify where to find it in the `HTPasswdFile` field under the `Web` config. 

On Linux, you can create this file using the following commands (requires `openssl`):

```bash
echo username:$(echo password | openssl passwd -stdin -apr1) > .htpasswd
```

Remember to replace `username` and `password` with your own values

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
Web:
  Address: localhost:8080
```
