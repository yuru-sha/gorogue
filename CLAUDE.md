# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## プロジェクト概要

GoRogueは、Go言語とGruidライブラリを使用した**本格的なローグライクゲーム**です。オリジナルRogueの26階層構造を忠実に再現し、手続き生成ダンジョン、ターンベース戦闘、パーマデス、探索重視のゲームプレイを提供することを目指しています。

### 現在の状態
GoRogueは現在、**初期開発段階**にあります：
- ✅ **基本的なゲームフレームワーク**（Go + Gruid）
- ✅ **基本的なプレイヤー移動システム**
- ✅ **セルベースのレンダリング**
- ✅ **Gruidライブラリ統合**
- 🚧 **開発中**: プレイヤーステータス、モンスターシステム、アイテムシステム

### 技術スタック
- **Go 1.21+**: モダンなGo言語機能を活用
- **Gruid**: ローグライクゲーム開発に特化したフレームワーク
- **GitHub**: バージョン管理とプロジェクト管理

### 開発方針
1. **段階的な実装**: 基本機能から順次実装
2. **Gruidライブラリの活用**: ローグライクゲームに特化したライブラリの効果的な活用
3. **シンプルな設計**: 複雑さを避けた実装
4. **Go言語の特徴を活かした実装**: 並行処理、型安全性、シンプルさを重視

## ディレクトリ構造

```
gorogue/
├── CLAUDE.md              # Claude Code guidance (this file)
├── README.md              # Project documentation
├── LICENSE                # License information
├── go.mod                 # Go module definition
├── go.sum                 # Go module checksums
├── main.go                # Application entry point
│
├── internal/              # Internal packages
│   ├── game/              # Game core
│   │   ├── engine.go      # Game engine
│   │   ├── player.go      # Player management
│   │   └── world.go       # World/Map management
│   │
│   ├── ui/                # User interface
│   │   ├── renderer.go    # Rendering system
│   │   └── input.go       # Input handling
│   │
│   └── util/              # Utility functions
│       └── logging.go     # Logging utilities
│
├── assets/                # Game assets
├── docs/                  # Documentation
│   ├── overview.md        # Project overview
│   ├── architecture.md    # Architecture documentation
│   ├── features.md        # Feature documentation
│   ├── development.md     # Development guide
│   └── task.md           # Task management
│
└── scripts/               # Build and development scripts
    └── build.sh           # Build script
```

## 開発コマンド

### 基本的な開発フロー
```bash
# 依存関係の更新
go mod tidy

# 開発サーバー実行
go run main.go

# ビルド
go build -o gorogue main.go

# テスト実行
go test ./...

# フォーマット
go fmt ./...

# 静的解析
go vet ./...
```

### 推奨開発ツール
```bash
# 静的解析ツール
go install honnef.co/go/tools/cmd/staticcheck@latest

# インポート整理
go install golang.org/x/tools/cmd/goimports@latest

# より高度なリンター
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest

# コード品質チェック（全体）
make lint

# golangci-lintのみ実行
golangci-lint run
```

## アーキテクチャ概要

### 設計原則
- **責務分離**: 各パッケージが明確な役割を持つ
- **シンプルさ**: Goらしいシンプルな設計
- **テスト可能性**: 依存関係の適切な管理
- **拡張性**: 新機能の追加が容易

### 主要コンポーネント
- **Game Engine**: ゲームループ、状態管理
- **Player**: プレイヤーの状態管理、移動処理
- **World**: マップ、ダンジョン生成
- **UI**: レンダリング、入力処理
- **Util**: ログ、ヘルパー関数

### 設計パターン
- **MVC Pattern**: モデル、ビュー、コントローラーの分離
- **State Pattern**: ゲーム状態の管理
- **Component Pattern**: ゲームオブジェクトの構成

## 開発ガイドライン

### コーディング規約
- **Go標準**: `go fmt`、`go vet`準拠
- **命名規約**: Goの慣習に従う
- **コメント**: 公開API、複雑なロジックに必須
- **エラーハンドリング**: Goらしいエラー処理

### Gruidライブラリの活用
```go
// 基本的なGruidの使用例
import "github.com/anaseto/gruid"

// ゲームモデル
type Model struct {
    // ゲーム状態
}

// ゲームループ
func (m *Model) Update(msg gruid.Msg) gruid.Effect {
    switch msg := msg.(type) {
    case gruid.MsgKeyDown:
        // キー入力処理
    case gruid.MsgQuit:
        // 終了処理
    }
    return nil
}

// 描画処理
func (m *Model) Draw(gd gruid.Grid) {
    // 描画ロジック
}
```

### テスト方針
- **単体テスト**: 各パッケージの機能テスト
- **統合テスト**: コンポーネント間の連携テスト
- **テストカバレッジ**: 主要機能の80%以上を目標

## 現在の実装状況

### 完成済み機能
- ✅ 基本的なゲームループ
- ✅ Gruidライブラリ統合
- ✅ プレイヤー移動システム
- ✅ セルベースのレンダリング
- ✅ 基本的なマップ描画

### 実装中の機能
- 🚧 プレイヤーステータス管理
- 🚧 基本的なダンジョン生成
- 🚧 ゲーム状態管理の改良

### 未実装の機能
- ❌ モンスターシステム
- ❌ アイテムシステム
- ❌ 戦闘システム
- ❌ セーブ/ロード機能
- ❌ 階層システム

## 次の実装ステップ

### 優先度: 高
1. **プレイヤーステータス管理の完全実装**
   - HP、攻撃力、防御力の管理
   - レベルシステム
   - 経験値システム

2. **ダンジョン生成システムの改良**
   - 部屋と通路の生成
   - 扉システム
   - 階段の配置

3. **基本的なモンスターシステム**
   - モンスターの配置
   - 基本的なAI
   - 視界システム

### 優先度: 中
1. **アイテムシステムの基礎**
   - アイテムの配置
   - 拾い上げ機能
   - 基本的なインベントリ

2. **戦闘システムの実装**
   - プレイヤーvsモンスター
   - ダメージ計算
   - ターンベース処理

### 優先度: 低
1. **UI/UX改善**
   - メッセージログ
   - ステータス表示
   - インベントリ画面

2. **セーブ/ロード機能**
   - ゲーム状態の保存
   - データの永続化

## トラブルシューティング

### よくある問題

#### 1. Gruidライブラリの問題
```bash
# モジュールが見つからない場合
go mod tidy
go mod download

# 古いバージョンの場合
go get -u github.com/anaseto/gruid
```

#### 2. ビルドエラー
```bash
# 依存関係の問題
go mod tidy

# キャッシュのクリア
go clean -modcache

# 再ビルド
go build -a main.go
```

#### 3. 実行時エラー
```bash
# デバッグモードで実行
go run -race main.go

# 詳細なエラー情報
GODEBUG=gctrace=1 go run main.go
```

### 開発支援

#### デバッグ方法
```go
// ログベースのデバッグ
import "log"

log.Printf("Debug: %+v", variable)

// Gruidのデバッグ機能
// 適切なデバッグ情報の出力
```

#### プロファイリング
```go
// パフォーマンス測定
import _ "net/http/pprof"

// http://localhost:6060/debug/pprof/
```

## 参考資料

### 公式ドキュメント
- [Go言語仕様](https://golang.org/ref/spec)
- [Gruidライブラリ](https://github.com/anaseto/gruid)
- [Gruidサンプルコード](https://github.com/anaseto/gruid/tree/master/examples)

### 開発リソース
- [Effective Go](https://golang.org/doc/effective_go.html)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)
- [Roguelike開発リソース](http://www.roguebasin.com/)

### 学習リソース
- [Go by Example](https://gobyexample.com/)
- [Go Tour](https://tour.golang.org/)
- [ローグライクゲーム開発入門](http://www.roguebasin.com/index.php?title=How_to_Write_a_Roguelike_in_15_Steps)

## 協力開発について

### 貢献方法
1. **Issues**: バグ報告、機能要求
2. **Pull Requests**: コード貢献
3. **Documentation**: ドキュメント改善
4. **Testing**: テストケース追加

### 開発環境
- **Go Version**: 1.21以上
- **IDE**: VS Code、GoLand推奨
- **OS**: クロスプラットフォーム対応

このプロジェクトは、Go言語とGruidライブラリを活用したローグライクゲーム開発の学習・実践を目的としています。段階的な実装により、基本的な機能から高度な機能まで順次追加していく予定です。