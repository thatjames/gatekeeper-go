FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY bin/gatekeeper gatekeeper
EXPOSE 8085
EXPOSE 53
EXPOSE 67
ENTRYPOINT ["/app/gatekeeper"]
