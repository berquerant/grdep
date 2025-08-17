FROM debian:bookworm-slim

RUN apt-get update \
    && apt-get install --no-install-recommends -y \
    curl \
    && apt-get clean \
    && rm -rf /var/lib/apt/lists/*

ENTRYPOINT ["curl"]
