#!/bin/sh
set -eu
/tailscaled --tun=userspace-networking --outbound-http-proxy-listen=localhost:1055 &
/tailscale up --authkey=${TAILSCALE_AUTHKEY} --hostname=gmailagg-$1
HTTP_PROXY=http://localhost:1055/ http_proxy=http://localhost:1055/ /gmailagg --config=gs://gmailagg-config/production.json $@
