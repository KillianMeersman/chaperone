FROM --platform=$BUILDPLATFORM golang:1-alpine AS builder

# Default BuildKit arguments, need to be defined as ARG to be useable in the Dockerfile.
# See https://www.docker.com/blog/faster-multi-platform-builds-dockerfile-cross-compilation-guide/
ARG TARGETOS
ARG TARGETARCH

LABEL org.opencontainers.image.source=https://github.com/KillianMeersman/chaperone
LABEL org.opencontainers.image.description="A rate-limiting & caching forward HTTP proxy."
LABEL org.opencontainers.image.licenses=MIT

WORKDIR /app
ENV GOCACHE=/app/.gocache

COPY . .

RUN --mount=type=cache,id=gocache,target=/app/.gocache,sharing=private \
    GOOS="$TARGETOS" GOARCH="$TARGETARCH" go build -o chaperone ./cmd/chaperone/main.go
RUN chmod +x chaperone

FROM alpine:3 AS main

WORKDIR /app
COPY --from=builder /app/chaperone chaperone

RUN adduser --disabled-password chaperone
USER chaperone:chaperone

ENTRYPOINT [ "/app/chaperone" ]
