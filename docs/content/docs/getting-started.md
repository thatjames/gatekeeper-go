---
title: 'Getting Started'
weight: 1
---

# Getting Started

There are a number of options available for running GateKeeper.

Currently, GateKeeper is not available for Windows.

## Docker

GateKeeper is built as a Docker image against a variety of architectures. The supported architectures are:

- amd64
- arm64
- armv7
- armv8
- i386

The image is available on github packages as [ghcr.io/thatjames/gatekeeper-go](https://github.com/thatjames/gatekeeper-go/pkgs/container/gatekeeper-go)

### HTPasswd

GateKeeper uses an `.htpasswd` file to authenticate users. This can be generated using openssl.

```bash
printf "username:$(openssl passwd -apr1 password)\n" > .htpasswd
```

Natrually, you need to replace `username` and `password` with the appropriate values.

If you do not configure a password, then default username/password combination is admin/admin.

```bash
docker run -d -p 53:53/udp -p 8085:8085 \
    -v /path/to/data:/var/lib/gatekeeper \
    -v /path/to/config.yml:/app/config.yml \
    -v /path/to/.htpasswd:/app/.htpasswd \
    --cap-add=NET_BIND_SERVICE \
    ghcr.io/thatjames/gatekeeper-go:latest
```

### Compose File

```yaml
services:
  gatekeeper:
    image: ghcr.io/thatjames/gatekeeper-go:latest
    container_name: gatekeeper
    ports:
      - 53:53/udp # DNS
      - 8085:8085 # Web
    volumes:
      - ./data/gatekeeper/data:/var/lib/gatekeeper/
      - ./data/gatekeeper/config.yml:/app/config.yml
      - ./data/gatekeeper/.htpasswd:/app/.htpasswd
    cap_add:
      - NET_BIND_SERVICE
```

**Note**: The keen eyed will notice the lack of a DHCP port in the initial examples.

This is because running a DHCP server inside of a container is a tricky proposition due to how Docker's default bridge networking works (specifically, bridge networks do not support broadcast propagation between the host and containers). If you need DHCP functionality, you must use host networking mode as shown in the examples below. This configuration is doable, but it is not something I will officially support, as network configurations vary widely and debugging network issues in containerized environments is complex.

That said, the below examples _might_ work for your use case, but I will not accept support requests if it does not.

```bash
docker run -d \
    -v /path/to/data:/var/lib/gatekeeper \
    -v /path/to/config.yml:/app/config.yml \
    -v /path/to/.htpasswd:/app/.htpasswd \
    --network=host \
    ghcr.io/thatjames/gatekeeper-go:latest
```

```yaml
services:
  gatekeeper:
    image: ghcr.io/thatjames/gatekeeper-go:latest
    container_name: gatekeeper
    volumes:
      - ./data/gatekeeper/data:/var/lib/gatekeeper/
      - ./data/gatekeeper/config.yml:/app/config.yml
      - ./data/gatekeeper/.htpasswd:/app/.htpasswd
    network_mode: host
```

If you are interested in running GateKeeper with the [DHCP module]({{< relref "/docs/modules#dhcp" >}}) enabled, you will have a much easier time running it as a native binary instead. True of many DHCP servers out there.

## Binaries

GateKeeper is also available as a binary that can be run on the supported architectures on linux. The binary is available on [GitHub Releases](https://github.com/thatjames/gatekeeper-go/releases).

### Installing from Source

To install from source, you will need to have [Go](https://golang.org/doc/install) 1.24.6 or higher installed, as well as the necessary dependencies for make.

```bash
git clone https://github.com/thatjames/gatekeeper-go.git
cd gatekeeper-go
make
```

### Running

To run GateKeeper, you will need to create a configuration file (see [Configuration](#configuration)). Once you have created the configuration file, you can run the binary with the following command:

```bash
bin/gatekeeper-go -config /path/to/config.yml
```

You can also include the `-debug` and `-trace` flags to enable debug and trace logging respectively.
