# GoRogue - Go言語で実装されたローグライクゲーム
# PyRogue (https://github.com/yuru-sha/pyrogue) を参考に作成

.PHONY: setup setup-dev build run clean test lint deps dev help check-setup check-setup-dev ci-checks qa-all

# 設定変数
GO_VERSION := 1.24.5
BINARY_NAME := gorogue
BUILD_DIR := bin
LOG_DIR := logs
GOLANGCI_LINT_VERSION := v1.61.0

# SDL2専用設定
RENDER_MODE := sdl2

# セットアップチェック用のマーカーファイル
SETUP_MARKER := .setup-check
SETUP_DEV_MARKER := .setup-dev-check

# 基本環境セットアップ
setup: $(SETUP_MARKER)

$(SETUP_MARKER):
	@echo "🔧 Setting up basic environment..."
	@echo "Checking Go version..."
	@go version
	@echo "Creating necessary directories..."
	@mkdir -p $(BUILD_DIR)
	@mkdir -p $(LOG_DIR)
	@echo "Installing base dependencies..."
	@go mod download
	@go mod tidy
	@echo "✅ Basic setup complete!"
	@touch $(SETUP_MARKER)

# 開発環境セットアップ（SDL2統合）
setup-dev: setup $(SETUP_DEV_MARKER)

$(SETUP_DEV_MARKER):
	@echo "🚀 Setting up development environment..."
	@echo "Installing development tools..."
	@command -v golangci-lint >/dev/null 2>&1 || { \
		echo "Installing golangci-lint..."; \
		curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(shell go env GOPATH)/bin $(GOLANGCI_LINT_VERSION); \
	}
	@command -v staticcheck >/dev/null 2>&1 || { \
		echo "Installing staticcheck..."; \
		go install honnef.co/go/tools/cmd/staticcheck@latest; \
	}
	@command -v goimports >/dev/null 2>&1 || { \
		echo "Installing goimports..."; \
		go install golang.org/x/tools/cmd/goimports@latest; \
	}
	@echo "🎮 Setting up SDL2 environment..."
	@echo "Checking SDL2 installation..."
	@if command -v pkg-config >/dev/null 2>&1; then \
		if pkg-config --exists sdl2; then \
			echo "✅ SDL2 found: $$(pkg-config --modversion sdl2)"; \
		else \
			echo "❌ SDL2 not found. Please install SDL2:"; \
			echo "  macOS: brew install sdl2"; \
			echo "  Ubuntu: sudo apt install libsdl2-dev"; \
			echo "  Fedora: sudo dnf install SDL2-devel"; \
			exit 1; \
		fi; \
	else \
		echo "⚠️  pkg-config not found. Please ensure SDL2 is installed."; \
	fi
	@echo "Installing SDL2 Go bindings..."
	@go get github.com/anaseto/gruid-sdl
	@echo "✅ Development environment setup complete!"
	@touch $(SETUP_DEV_MARKER)

# セットアップ確認
check-setup:
	@if [ ! -f $(SETUP_MARKER) ]; then \
		echo "❌ Basic setup not found. Run 'make setup' first."; \
		exit 1; \
	fi
	@echo "✅ Basic setup verified"

check-setup-dev:
	@if [ ! -f $(SETUP_DEV_MARKER) ]; then \
		echo "❌ Development setup not found. Run 'make setup-dev' first."; \
		exit 1; \
	fi
	@echo "✅ Development setup verified"

# ビルドターゲット
build: check-setup
	@echo "🔨 Building $(BINARY_NAME)..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gorogue
	@echo "✅ Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# CLIモードビルド
build-cli: check-setup
	@echo "🔨 Building $(BINARY_NAME)-cli..."
	@go build -o $(BUILD_DIR)/$(BINARY_NAME)-cli ./cmd/gorogue-cli
	@echo "✅ CLI build complete: $(BUILD_DIR)/$(BINARY_NAME)-cli"

# 両方ビルド
build-all: build build-cli

# 実行（GUI版）
run: build
	@echo "🎮 Starting $(BINARY_NAME)..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# CLI版実行
run-cli: build-cli
	@echo "💻 Starting $(BINARY_NAME) CLI mode..."
	@./$(BUILD_DIR)/$(BINARY_NAME)-cli


# 開発用実行（デバッグ情報付き）
dev: check-setup-dev clean
	@echo "🐛 Building with debug info..."
	@go build -gcflags="all=-N -l" -o $(BUILD_DIR)/$(BINARY_NAME) ./cmd/gorogue
	@echo "🎮 Starting $(BINARY_NAME) in debug mode..."
	@./$(BUILD_DIR)/$(BINARY_NAME)

# テスト実行
test: check-setup
	@echo "🧪 Running tests..."
	@go test -v ./...

# テスト（カバレッジ付き）
test-coverage: check-setup
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "📊 Coverage report generated: coverage.html"

# リント実行
lint: check-setup-dev
	@echo "🔍 Running linters..."
	@go fmt ./...
	@go vet ./...
	@golangci-lint run
	@staticcheck ./...
	@echo "✅ Linting complete"

# 依存関係の更新
deps: check-setup
	@echo "📦 Updating dependencies..."
	@go get -u ./...
	@go mod tidy
	@echo "✅ Dependencies updated"

# クリーンアップ
clean:
	@echo "🧹 Cleaning up..."
	@rm -rf $(BUILD_DIR)/
	@rm -f coverage.out coverage.html
	@go clean
	@echo "✅ Cleanup complete"

# 完全クリーンアップ（セットアップも削除）
clean-all: clean
	@echo "🧹 Complete cleanup..."
	@rm -f $(SETUP_MARKER) $(SETUP_DEV_MARKER)
	@rm -rf $(LOG_DIR)/*
	@echo "✅ Complete cleanup finished"

# CI用チェック
ci-checks: check-setup-dev
	@echo "🔄 Running CI checks..."
	@make lint
	@make test
	@echo "✅ CI checks passed"

# QA用統合チェック
qa-all: check-setup-dev
	@echo "🔍 Running comprehensive QA checks..."
	@make clean
	@make build
	@make lint
	@make test-coverage
	@echo "✅ QA checks completed"

# 開発用の便利コマンド
fmt: check-setup-dev
	@echo "📝 Formatting code..."
	@go fmt ./...
	@goimports -w .
	@echo "✅ Code formatted"

# ログの確認
logs:
	@echo "📋 Recent game logs:"
	@if [ -f $(LOG_DIR)/game.log ]; then \
		tail -20 $(LOG_DIR)/game.log; \
	else \
		echo "No game logs found. Run the game first."; \
	fi

# バージョン情報
version:
	@echo "GoRogue Version Information:"
	@echo "Go Version: $(GO_VERSION)"
	@echo "Binary Name: $(BINARY_NAME)"
	@echo "Build Directory: $(BUILD_DIR)"
	@echo "Log Directory: $(LOG_DIR)"
	@go version

# ヘルプ
help:
	@echo "🎮 GoRogue - Go言語で実装されたローグライクゲーム"
	@echo ""
	@echo "📋 Available commands:"
	@echo ""
	@echo "🔧 Setup Commands:"
	@echo "  make setup       - Basic environment setup"
	@echo "  make setup-dev   - Development environment setup"
	@echo "  make setup-sdl   - SDL2 environment setup"
	@echo "  make check-setup - Verify basic setup"
	@echo "  make check-setup-dev - Verify development setup"
	@echo ""
	@echo "🎮 Game Commands:"
	@echo "  make build       - Build the game (SDL2 graphics)"
	@echo "  make build-cli   - Build CLI version"
	@echo "  make build-all   - Build both GUI and CLI versions"
	@echo "  make run         - Build and run the game (SDL2 graphics)"
	@echo "  make run-cli     - Build and run CLI version"
	@echo "  make dev         - Build with debug info and run"
	@echo ""
	@echo "🧪 Development Commands:"
	@echo "  make test        - Run tests"
	@echo "  make test-coverage - Run tests with coverage"
	@echo "  make lint        - Run linters"
	@echo "  make fmt         - Format code"
	@echo "  make deps        - Update dependencies"
	@echo ""
	@echo "🔍 Quality Assurance:"
	@echo "  make ci-checks   - Run CI checks"
	@echo "  make qa-all      - Run comprehensive QA checks"
	@echo ""
	@echo "🛠️ Utility Commands:"
	@echo "  make clean       - Clean build artifacts"
	@echo "  make clean-all   - Complete cleanup"
	@echo "  make logs        - Show recent game logs"
	@echo "  make version     - Show version information"
	@echo "  make help        - Show this help"
	@echo ""
	@echo "🚀 Quick Start:"
	@echo "  make setup-dev   # First time setup"
	@echo "  make run         # Start playing!"

# デフォルトターゲット
.DEFAULT_GOAL := help 