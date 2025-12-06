FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY bin/gatekeeper gatekeeper
EXPOSE 8085
EXPOSE 53
ENTRYPOINT ["/app/gatekeeper"]
