# syntax=docker/dockerfile:1

FROM golang:1.24-alpine AS builder
WORKDIR /app

ENV USER=appuser
ENV UID=10001
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

RUN apk add --no-cache git ca-certificates


COPY go.mod go.sum ./
RUN go mod download

COPY . .


RUN go build -ldflags="-w -s" -o /main .

FROM alpine:3.18
WORKDIR /app


RUN apk add --no-cache ca-certificates

COPY --from=builder /main /app/main
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group
COPY --from=builder /app/users.yaml /app/users.yaml


USER appuser:appuser

EXPOSE 9091

ENTRYPOINT ["/app/main"]

LABEL org.opencontainers.image.description="Promo Bot"
LABEL org.opencontainers.image.licenses=MIT
LABEL org.opencontainers.image.source="https://github.com/MikebangSfilya/promoBot"