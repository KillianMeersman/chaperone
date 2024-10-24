FROM golang:1-alpine AS builder

WORKDIR /app
ENV GOCACHE=/app/.gocache

COPY . .
RUN --mount=type=cache,id=gocache,target=/app/.gocache,sharing=locked go build -o chaperone ./cmd/chaperone/main.go
RUN chmod +x chaperone

FROM alpine:3 AS main

WORKDIR /app
COPY --from=builder /app/chaperone chaperone

RUN adduser --disabled-password user
USER user:user

ENTRYPOINT [ "/app/chaperone" ]
