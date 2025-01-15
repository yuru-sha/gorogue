.PHONY: build run clean test lint

# ビルドターゲット
build:
	go build -o bin/gorogue cmd/gorogue/main.go

# 実行
run: build
	./bin/gorogue

# クリーンアップ
clean:
	rm -rf bin/
	go clean

# テスト実行
test:
	go test -v ./...

# リント実行
lint:
	go vet ./...
	golangci-lint run

# 依存関係の更新
deps:
	go mod tidy

# 開発用ビルド（デバッグ情報付き）
dev: clean
	go build -gcflags="all=-N -l" -o bin/gorogue cmd/gorogue/main.go
	./bin/gorogue

# ヘルプ
help:
	@echo "Available commands:"
	@echo "  make build    - Build the game"
	@echo "  make run     - Build and run the game"
	@echo "  make clean   - Clean build artifacts"
	@echo "  make test    - Run tests"
	@echo "  make lint    - Run linters"
	@echo "  make deps    - Update dependencies"
	@echo "  make dev     - Build with debug info and run"
	@echo "  make help    - Show this help"

# デフォルトターゲット
.DEFAULT_GOAL := help 