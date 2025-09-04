FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY bin/gatekeeper gatekeeper
ENTRYPOINT ["/app/gatekeeper"]
