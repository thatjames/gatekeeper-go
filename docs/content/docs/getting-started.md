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

The image is available on Docker Hub as [`smokeycircles/gatekeeper`](https://hub.docker.com/r/smokeycircles/gatekeeper).

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
