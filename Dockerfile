FROM golang:1-alpine AS builder
ARG EXECUTABLE

WORKDIR /app
ENV GOCACHE=/app/.gocache

COPY . .
RUN --mount=type=cache,id=gocache,target=/app/.gocache,sharing=locked go build -o main ./cmd/${EXECUTABLE}/main.go
RUN chmod +x main

FROM alpine:3 AS main

WORKDIR /app
COPY --from=builder /app/main main

RUN adduser --disabled-password user
USER user:user

ENTRYPOINT [ "/app/main" ]
