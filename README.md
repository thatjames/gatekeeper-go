# GateKeeper

GateKeeper is a configurable hobbyist DHCP server written in Go.

It is **not** intended for production use, but it powers my home network just fine!

**UPDATE** Version 2.0.0 brings a new UI and a docker image!

## Installation

Install gatekeeper from source with `go install gitlab.com/thatjames-go/gatekeeper-go@latest` or download one of the binaries under the [releases](https://gitlab.com/thatjames-go/gatekeeper-go/-/releases) page.

## Building

### Native Binary

```bash
make build
```

### Docker Image

```bash
make docker
```

## Running

### Native

The binary will look for a `config.yml` file in the PWD, or it can be passed with the `-config <file>` flag.

### Docker

You can run the docker image with the following command:

```bash
docker run --name gatekeeper -dp 8080:8080 -v/path/to/config.yml:/app/config.yml -v /path/to/leases:/var/lib/gatekeeper/leases thatjames/gatekeeper:latest
```

#### Docker Compose

You can also use docker compose to run the docker image:

````yaml
version: "3.9"
services:
  gatekeeper:
    image: thatjames/gatekeeper:latest
    container_name: gatekeeper
    ports:
      - "8080:8080"
    volumes:
      - ./config.yml:/app/config.yml
      - ./leases:/var/lib/gatekeeper/leases

### Webserver

If the webserver is enabled, but no `htpasswd` file has been specified, then the default login credentials are admin/admin.

In order to secure the webserver, you can create an `htpasswd` file and specify where to find it in the `HTPasswdFile` field under the `Web` config.

On Linux, you can create this file using the following commands (requires `openssl`):

```bash
echo username:$(echo password | openssl passwd -stdin -apr1) > .htpasswd
````

Remember to replace `username` and `password` with your own values

#### Screenshots

[WebUI Home Page](images/web_home_page.png)

[WebUI DHCP View](images/web_dhcp_active_leases.png)

[WebUI DHCP Active Lease Options](images/web_dhcp_active_lease_options.png)

[WebUI DHCP Reserved Leases](images/web_dhcp_reserved_leases.png)

[WebUI DHCP Reserved Lease Options](images/web_dhcp_reserved_lease_options.png)

[WebUI DHCP Options](images/web_dhcp_options_display.png)

[WebUI DHCP Options Config](images/web_dhcp_options_form.png)

[WebUI DHCP View]

## Configuration

Example config:

```yaml
DHCP:
  Interface: eth0
  StartAddr: 10.0.0.2
  EndAddr: 10.0.0.99
  DomainName: international-space-station
  NameServers:
    - 8.8.8.8
    - 1.1.1.1
  LeaseTTL: 300
  SubnetMask: 255.255.255.0
  Gateway: 10.0.0.1
  ReservedAddresses:
    00:d8:61:39:b5:6a: 10.0.0.101
    bc:5f:f4:ac:ea:c4: 10.0.0.100
  LeaseFile: /var/lib/gatekeeper/leases
Web:
  Address: :8085
  HTPasswdFile: .htpasswd
  Prometheus: true
```

### DHCP

The DHCP configuration section.

#### Interface

The interface name gatekeeper will bind to

#### DomainName

The domain name option supplied in DHCP responses

#### StartAddr

The first assignable DHCP address gatekeeper will hand out

#### EndAddr

The last assignable DHCP address gatekeeper will hand out. This makes the effective pool equal to (StartAddr - EndAddr) + 1 (inclusive). Our example has 98 possible addresses.

#### Lease TTL

Time To Live for given leases, in seconds. Our example has a TLL of 300 seconds, or 5 minutes.

#### SubnetMask

The subnet mask option returned in the DHCP response

#### Gateway

The gateway option returned in the DHCP response

#### NameServers:

An array of values for the NameServer option(s) returned in the DHCP response

#### ReservedAddresses

A mapping of mac_address:desired_static_ip to provide static IP addresses to clients. Make sure this is outside of the assignable range of start and end addresses

### Web

The webserver configuration section. If this section is not defined, gatekeeper will not run a web server.

#### Address

Listen address the webserver should bind to

#### HTPasswdFile

Relative or Fully Qualified path to the htaccess file

#### Prometheus

If true, the webserver will expose a prometheus metrics endpoint at `/metrics`

**note:** if no `HTPasswdFile` is provided, then the default username/password is admin/admin
