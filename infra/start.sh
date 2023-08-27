#!/bin/sh
set -eu
/tailscaled --tun=userspace-networking --outbound-http-proxy-listen=localhost:1055 &
/tailscale up --authkey=${TAILSCALE_AUTHKEY} --hostname=gmailagg-job
HTTP_PROXY=http://localhost:1055/ http_proxy=http://localhost:1055/ /gmailagg --token=gs://gmailagg-oauth/token.json run
