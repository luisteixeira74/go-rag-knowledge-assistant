# --- Stage 1: Build ---
FROM golang:alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

# Permite baixar o toolchain do Go 1.26 automaticamente
ENV GOTOOLCHAIN=auto
ENV GOPROXY=https://proxy.golang.org,direct

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-w -s" -o /app/bin/api ./cmd/api/main.go

# --- Stage 2: Runner ---
FROM alpine:latest

WORKDIR /app

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /app/bin/api /app/api

RUN mkdir -p /app/data/input

EXPOSE 8080

CMD ["/app/api"]