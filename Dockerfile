# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Install dependencies
RUN apk add --no-cache git ca-certificates tzdata

# Copy go mod files
COPY go.mod go.sum ./

# 🔥 关键修复：国内 Go 代理，解决下载超时
ENV GOPROXY=https://goproxy.cn,direct
ENV GO111MODULE=on

RUN go mod download

# Copy source code
COPY . .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /steellink-backend ./cmd/api

# Final stage
FROM alpine:3.19

WORKDIR /app

# Install ca-certificates and timezone data
RUN apk --no-cache add ca-certificates tzdata && \
    cp /usr/share/zoneinfo/Asia/Shanghai /etc/localtime && \
    echo "Asia/Shanghai" > /etc/timezone && \
    apk del tzdata

# Copy binary from builder
COPY --from=builder /steellink-backend /app/steellink-backend
COPY --from=builder /app/config/config.example.yaml /app/config/config.yaml

# Create non-root user
RUN adduser -D -g '' appuser
USER appuser

EXPOSE 8080

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

ENTRYPOINT ["/app/steellink-backend"]
CMD ["-config", "/app/config/config.yaml"]