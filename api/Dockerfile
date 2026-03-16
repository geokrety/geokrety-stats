FROM golang:1.24-alpine AS builder
WORKDIR /app
COPY go.mod go.sum* ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/geokrety-stats-api ./cmd/server

FROM gcr.io/distroless/static-debian12:nonroot
WORKDIR /app
COPY --from=builder /out/geokrety-stats-api /app/geokrety-stats-api
COPY docs/openapi.yaml /app/docs/openapi.yaml
EXPOSE 7415
ENTRYPOINT ["/app/geokrety-stats-api"]
