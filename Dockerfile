FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY bin/gatekeeper gatekeeper
ENV GIN_MODE=release
ENTRYPOINT ["/app/gatekeeper"]
