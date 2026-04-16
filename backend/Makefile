.PHONY: dev test lint migrate-up migrate-down migrate-create build

# 本地开发启动
dev:
	go run cmd/server/main.go

# 运行所有测试
test:
	go test ./... -v -count=1

# 代码检查
lint:
	go vet ./...

# 数据库迁移 - 向上
migrate-up:
	migrate -path migrations -database "$$DATABASE_URL" up

# 数据库迁移 - 回滚
migrate-down:
	migrate -path migrations -database "$$DATABASE_URL" down 1

# 创建新迁移文件 (make migrate-create NAME=xxx)
migrate-create:
	migrate create -ext sql -dir migrations -seq $(NAME)

# 构建二进制
build:
	CGO_ENABLED=0 go build -o bin/server cmd/server/main.go
