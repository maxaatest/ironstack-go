# Build stage
FROM golang:1.22-alpine AS builder

RUN apk add --no-cache git make

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o ironstack ./cmd/ironstack

# Runtime stage
FROM alpine:3.19

RUN apk add --no-cache \
    ca-certificates \
    curl \
    bash \
    tzdata

WORKDIR /app

COPY --from=builder /app/ironstack /usr/local/bin/ironstack

# Create directories
RUN mkdir -p /etc/ironstack /var/log/ironstack /backups /var/www

# Set timezone
ENV TZ=UTC

ENTRYPOINT ["ironstack"]
