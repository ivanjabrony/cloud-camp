FROM golang:1.24.3 AS builder

WORKDIR /app

COPY . .
RUN go mod download
RUN CGO_ENABLED=0 GOOS=linux go build -o ratelimiter ./cmd/ratelimiter/main.go \
    && go clean -cache -modcache

FROM alpine:latest

WORKDIR /root

COPY --from=builder /app/ratelimiter .
COPY --from=builder /app/internal/ratelimit/migrations ./migrations
COPY --from=builder /app/cmd/ratelimiter/config.json .

EXPOSE 3000
CMD ["./ratelimiter", "./config.json"]