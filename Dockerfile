# Start a new stage from scratch
FROM alpine:3.20.0

# Set the working directory inside the container
WORKDIR /app

# Copy the binary from the builder stage
COPY bin/shortcut /app/shortcut
COPY start.sh /app/start.sh

RUN apk update && apk add ca-certificates && rm -rf /var/cache/apk/*

# Copy Tailscale binaries from the tailscale image on Docker Hub.
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscaled /app/tailscaled
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscale /app/tailscale
RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale


# Expose the port the app runs on
EXPOSE 8080

# Command to run the executable
CMD ["/app/start.sh"]