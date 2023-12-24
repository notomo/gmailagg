FROM alpine:3.19

RUN adduser nonroot -D

COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscaled /tailscaled
COPY --from=docker.io/tailscale/tailscale:stable /usr/local/bin/tailscale /tailscale
RUN mkdir -p /var/run/tailscale /var/cache/tailscale /var/lib/tailscale \
  && chown nonroot /var/run/tailscale /var/cache/tailscale /var/lib/tailscale /tailscaled /tailscale

USER nonroot

WORKDIR /
COPY . /

ENTRYPOINT ["/start.sh"]
