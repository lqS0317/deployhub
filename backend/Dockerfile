# Stage 1: 构建阶段
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o deployhub ./cmd/server

# Stage 2: 运行阶段
FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/deployhub .
COPY --from=builder /app/migrations ./migrations
EXPOSE 8080
CMD ["./deployhub"]
