# ============================
# 1. Build stage
# ============================
FROM golang:1.25-alpine AS builder

RUN apk add --no-cache git

WORKDIR /app

# for caching
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go install github.com/pressly/goose/v3/cmd/goose@v3.25.0
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o server ./cmd

# ============================
# 2. Runtime stage
# ============================
FROM gcr.io/distroless/static:nonroot

WORKDIR /app

COPY --from=builder /app/server .

USER nonroot:nonroot

EXPOSE 8080

ENTRYPOINT ["./server"]

# ============================
# 3. Migration stage
# ============================
FROM debian:12-slim AS migrate

WORKDIR /app

COPY --from=builder /go/bin/goose /usr/local/bin/goose

COPY migrations /app/migrations

# Needed for postgres TLS, DNS
RUN apt-get update && apt-get install -y ca-certificates && update-ca-certificates

ENTRYPOINT []
CMD []