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
