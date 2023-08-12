FROM gcr.io/distroless/base-debian11

USER nonroot:nonroot

WORKDIR /
COPY ./gmailagg /gmailagg
COPY ./config.yaml /config.yaml

ENTRYPOINT ["/gmailagg", "--token=gs://gmailagg-token/token.json", "--config=/config.yaml", "run"]
