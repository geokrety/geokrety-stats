# ── Build stage ─────────────────────────────────────────────────────────────
FROM golang:1.24-alpine AS builder
RUN apk add --no-cache git ca-certificates

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /bin/geokrety-stats ./cmd/geokrety-stats

# ── Runtime stage ────────────────────────────────────────────────────────────
FROM scratch
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /bin/geokrety-stats /geokrety-stats

ENTRYPOINT ["/geokrety-stats"]
