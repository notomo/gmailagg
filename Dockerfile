FROM alpine:3.18

COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscaled /tailscaled
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscale /tailscale
RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale

WORKDIR /
COPY . /

ENTRYPOINT ["/start.sh"]
