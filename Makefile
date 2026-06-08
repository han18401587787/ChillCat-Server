.PHONY: build run test clean docker-build docker-up migrate

# 构建
build:
	go build -o bin/server ./cmd/server

# 运行
run:
	go run ./cmd/server

# 测试
test:
	go test -v -race ./...

# 清理
clean:
	rm -rf bin/
	rm -rf tmp/

# Docker 构建
docker-build:
	docker build -t chillcat-server:latest .

# Docker 启动
docker-up:
	docker-compose up -d

# Docker 停止
docker-down:
	docker-compose down

# 查看日志
docker-logs:
	docker-compose logs -f

# 填充种子数据
seed:
	go run ./cmd/seed

# 数据库迁移（手动）
migrate:
	go run ./cmd/migrate

# 代码格式化
fmt:
	go fmt ./...

# 代码检查
lint:
	golangci-lint run ./...

# 依赖更新
deps:
	go mod tidy
	go mod vendor

# 查看 API 文档
doc:
	@echo "打开浏览器访问 http://localhost:8080/swagger/index.html"
