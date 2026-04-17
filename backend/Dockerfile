# Stage 1: 构建阶段
FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o deployhub ./cmd/server

# Stage 2: 运行阶段
FROM amazon/aws-cli:latest AS awscli

FROM alpine:3.19
RUN apk --no-cache add ca-certificates tzdata gcompat
COPY --from=awscli /usr/local/aws-cli/ /usr/local/aws-cli/
RUN ln -s /usr/local/aws-cli/v2/current/bin/aws /usr/local/bin/aws
WORKDIR /app
COPY --from=builder /app/deployhub .
EXPOSE 8080
CMD ["./deployhub"]
