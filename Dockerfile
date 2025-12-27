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

COPY . ./

RUN go build -ldflags="-w -s" -o /promobot .


FROM alpine:3.18

LABEL org.opencontainers.image.description="Promo Bot"
LABEL org.opencontainers.image.licenses=MIT
LABEL org.opencontainers.image.source="https://github.com/MikebangSfilya/promoBot"
LABEL com.centurylinklabs.watchtower.enable="true"

WORKDIR /app

RUN apk add --no-cache ca-certificates

COPY --from=builder /promobot /app/promobot
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

RUN mkdir -p /app/audit-logs && \
    chown -R 10001:10001 /app/audit-logs && \
    chmod 755 /app/audit-logs

USER appuser:appuser

ENTRYPOINT ["/app/promobot"]
