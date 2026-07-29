# syntax=docker/dockerfile:1

FROM golang:1.26.1-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
COPY vendor ./vendor
COPY backend ./backend

RUN CGO_ENABLED=0 \
    GOOS=linux \
    go build \
    -mod=vendor \
    -trimpath \
    -ldflags="-s -w" \
    -o /out/todo-server \
    ./backend/cmd

FROM ubuntu:24.04

WORKDIR /app

RUN apt-get update \
    && apt-get install -y --no-install-recommends \
        ca-certificates \
        tzdata \
    && rm -rf /var/lib/apt/lists/* \
    && mkdir -p /app/logs /data

COPY --from=builder /out/todo-server ./todo-server
COPY web ./web

ENV TODO_PORT=7540
ENV TODO_DBFILE=/data/scheduler.db

EXPOSE 7540

VOLUME ["/data"]

ENTRYPOINT ["/app/todo-server"]