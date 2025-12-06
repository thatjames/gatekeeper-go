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

The image is available on [github packages](https://github.com/thatjames?tab=packages&repo_name=gatekeeper-go)

### Compose File

```yaml
version: '3.8'
services:
  gatekeeper:
    image: smokeycircles/gatekeeper:latest
    container_name: gatekeeper
    ports:
      - 53:53/udp
    volumes:
      - ./data/gatekeeper/data:/var/lib/gatekeeper/
      - ./data/gatekeeper/config.yml:/app/config.yml
      - ./data/gatekeeper/.htpasswd:/app/.htpasswd
```

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
