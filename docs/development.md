---
cache_control: {"type": "ephemeral"}
---
# GoRogue 開発者ガイド

このドキュメントでは、GoRogueの開発に必要な情報を提供します。

## 目次

1. [開発環境のセットアップ](#開発環境のセットアップ)
2. [プロジェクト構造](#プロジェクト構造)
3. [設定管理](#設定管理)
4. [開発ワークフロー](#開発ワークフロー)
5. [テスト](#テスト)
6. [コーディング規約](#コーディング規約)

## 開発環境のセットアップ

### 必要条件

- Go 1.22以上
- make（ビルドツール）
- Git（バージョン管理）
- golangci-lint v1.64.8（静的解析ツール。`make setup-dev`でバージョンを検査）

### セットアップ手順

```bash
# リポジトリのクローン
git clone https://github.com/yuru-sha/gorogue.git
cd gorogue

# 依存関係のインストール
go mod tidy

# ビルドの確認
go build ./cmd/gorogue

# 開発ツールのインストール
make setup-dev
```

`make setup-dev`および`make ci-checks`は、golangci-lint v1.64.8を使用します。GoのbinディレクトリがPATHに含まれていない場合も、セットアップとlintはインストール先の実行ファイルを直接使用します。PATH上に別のバージョンがある場合は、そのバージョンを検査してエラーで停止します。

クリーンセットアップからツールの導入とバージョン検査を検証する場合は、セットアップマーカーを削除してから次を実行します。

```bash
rm -f .setup-check .setup-dev-check
make setup-dev
make check-setup-dev
```

`make ci-checks`は全体の品質ゲートです。既存のlint指摘は[#26](https://github.com/yuru-sha/gorogue/issues/26)で解消済みで、以後の新しい指摘もこのコマンドで検出されます。

```bash
make ci-checks
```

## プロジェクト構造

```
gorogue/
├── docs/                # ドキュメント
├── assets/             # アセット（フォント、画像など）
├── cmd/gorogue/        # メインアプリケーション
├── internal/           # 内部パッケージ
│   ├── game/         # ゲームロジック
│   ├── entity/       # エンティティ管理
│   ├── dungeon/      # ダンジョン生成
│   └── ui/           # ユーザーインターフェース
├── pkg/                # 外部公開パッケージ
└── test/               # テストコード
```

## 設定管理

GoRogueでは環境変数と設定ファイルによる設定管理を採用しています。開発・運用設定の分離が可能です。

### 環境変数の設定

1. **開発環境での設定**：
```bash
# 環境変数テンプレートをコピー
cp .env.example .env

# 環境変数ファイルを編集
vim .env  # または好みのエディタで編集
```

2. **実行時の環境変数指定**：
```bash
# 一時的な設定変更
GOROGUE_DEBUG=true go run ./cmd/gorogue

# 永続的な設定変更は.envファイルで行う
```

### 主要な環境変数

| 環境変数 | デフォルト値 | 説明 | 用途 |
|----------|--------------|------|------|
| `GOROGUE_DEBUG` | false | デバッグモード | 開発時の詳細ログ出力 |
| `GOROGUE_LOG_LEVEL` | INFO | ログレベル | DEBUG/INFO/WARNING/ERROR |
| `GOROGUE_PROFILE` | false | プロファイリング | パフォーマンス解析 |
| `GOROGUE_SAVE_DIR` | ./saves | セーブディレクトリ | セーブファイルの保存場所 |

### 開発・運用環境の分離

**開発環境設定例**（.env.dev）：
```bash
GOROGUE_DEBUG=true
GOROGUE_LOG_LEVEL=DEBUG
GOROGUE_PROFILE=true
GOROGUE_SAVE_DIR=./saves/dev
```

**運用環境設定例**（.env.prod）：
```bash
GOROGUE_DEBUG=false
GOROGUE_LOG_LEVEL=INFO
GOROGUE_PROFILE=false
GOROGUE_SAVE_DIR=./saves
```

### 再現可能なゲーム

GUIは起動時にシードを自動生成します。CLIでは`-seed`を指定すると同じダンジョン生成を再現できます。

```bash
go run ./cmd/gorogue-cli -seed 12345
```

コードからは`core.NewEngineWithSeed`または`dungeon.NewDungeonManagerWithSeed`を使用します。

### 実装場所
- `internal/config/` - 設定管理パッケージ
- `internal/game/settings/` - ゲーム設定構造体

## 開発ワークフロー

1. 新機能の開発：
```bash
# 新しいブランチを作成
git checkout -b feature/new-feature

# コードの変更
make fmt      # コードフォーマット
make build    # ビルド確認
make test     # テスト実行

# 変更のコミット
git add .
git commit -m "Add new feature"
```

2. コードの検証：
```bash
make lint     # 静的解析（golangci-lint）
make test     # テストの実行
make check    # 全チェック（fmt, lint, test）
```

## テスト

テストは`go test`を使用して実行します：

```bash
# 全てのテストを実行
make test

# 特定のパッケージのテストを実行
go test ./internal/game/...

# カバレッジレポートの生成
go test -cover ./...

# 詳細なカバレッジレポート
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# ベンチマークテストの実行
go test -bench=. ./...
```

## コーディング規約

- コードフォーマット：`gofmt`、`goimports`
- 静的解析：`golangci-lint`（複数のlinterを統合実行）
- 命名規約：Go標準の命名規則に従う
- コメント：godoc形式でのドキュメント作成

これらのツールは`make lint`で一括実行できます。

### 主要な規約

1. **命名規約**：
   - パッケージ名：小文字、短く、意味のある名前
   - 関数・変数：キャメルケース、exportする場合は大文字で開始
   - 定数：大文字、アンダースコア区切り

2. **コメント**：
   - 公開関数・型にはgodoc形式のコメントを必須
   - 複雑なロジックには説明コメントを追加

3. **エラーハンドリング**：
   - エラーは明示的に処理する
   - 適切なエラーメッセージを提供

## コマンド統一化システムの開発

### 概要

GUIとCLIのゲームプレイ操作は、`internal/core/command.Execute` を共通の実行本体として使用します。GUIは `Parser` でキー入力を `Command` に変換し、CLIは `internal/core/cli/game_commands.go` でテキスト入力を `Command` に変換します。新しいゲームプレイ操作は共通実行本体に追加し、両方の入口から呼び出してください。

### 新しいコマンドの追加手順

#### 1. コマンド種別と入口を追加

```go
// internal/core/command/types.go
// Type に CmdNewAction を追加する

// internal/core/command/parser.go
p.keyMap["x"] = Command{Type: CmdNewAction}

// internal/core/cli/game_commands.go
return c.executeGameplay(Command{Type: CmdNewAction}, args...)
```

#### 2. 共通実行本体を実装

```go
// internal/core/command/executor.go
func executeNewAction(ctx *Context, args []string) Result {
    // ゲーム状態の検証と状態変更をここに実装する
    return turnResult(ctx, "Action performed successfully.")
}
```

`Execute` の switch に `CmdNewAction` を追加し、ターン消費やエラーは `Result` に設定します。GUI側は `GameScreen.executeCommand`、CLI側は `CLIMode.executeGameplay` を経由させ、入口ごとにゲームルールを複製しないでください。

#### 3. テストの作成

```go
// internal/ui/screen/gameplay_parity_test.go
func TestGameplayCommandsHaveGUICLIParity(t *testing.T) {
    // GUIキー入力とCLIコマンドで、状態・メッセージ・ターン消費を比較する
}
```

### 開発のベストプラクティス

#### コマンド設計の原則

1. **一貫性**: 既存コマンドとの命名規則を統一
2. **直感性**: ユーザーが理解しやすいコマンド名
3. **エイリアス**: 短縮形と完全形の両方を提供
4. **エラーハンドリング**: 適切なエラーメッセージの提供

#### 実装の注意点

1. **両環境対応**: CLIとGUIの両方で動作することを確認
2. **テスト**: 新機能のテストケースを必ず作成
3. **ドキュメント**: ヘルプテキストとドキュメントの更新
4. **型安全性**: Goの型システムの適切な使用

### デバッグとテスト

#### CLIモードでのテスト

```bash
# CLIモードで新しいコマンドをテスト
go run ./cmd/gorogue-cli
> help           # ヘルプの確認
> move east      # ゲームプレイコマンドのテスト
```

#### 単体テストの実行

```bash
# 共通実行本体とGUI/CLIのパリティテストを実行
go test ./internal/core/command ./internal/core/cli ./internal/ui/screen
```

### トラブルシューティング

#### よくある問題

1. **コマンドが認識されない**
   - `command.Execute` の `switch` とCLIのコマンド登録を確認
   - コマンド名のスペルチェック

2. **キー入力が反応しない**
   - `command.Parser` のキーマッピングを確認
   - キーコードの正確性をチェック

3. **テストが失敗する**
   - モックオブジェクトの設定を確認
   - 期待される戻り値の検証

## 既知の課題 (Known Issues)

### UI関連の課題

#### 1. 入力処理の修正中問題
- **場所**: `internal/ui/input_handler.go`
- **問題**: キーボード入力処理の一部で不具合が発生
- **詳細**: 特定のキー組み合わせで期待通りの動作がしない場合がある
- **影響範囲**: 特定の操作シナリオでのユーザー体験
- **対応状況**: 修正作業中
- **回避策**: 代替キーバインドの使用

#### 2. 大規模マップでのレンダリングパフォーマンス
- **問題**: マップサイズ拡大時の描画処理負荷増加
- **影響**: フレームレートの低下、操作レスポンス悪化
- **原因**: 全タイル描画によるCPU負荷
- **対策候補**:
  - タイル描画の最適化
  - 差分描画システムの強化
- **優先度**: 中

#### 3. 複雑なゲーム状態のシリアライゼーション
- **場所**: `internal/game/save/save_manager.go`
- **問題**: セーブデータの一貫性保証が困難
- **詳細**:
  - フロアデータの完全復元
  - エンティティ状態の複雑な依存関係
  - メモリ効率とデータ整合性のトレードオフ
- **影響**: セーブ・ロード機能の信頼性
- **対応状況**: 継続的改善中
- **関連**: Permadeathシステムとの連携

### 対応予定とロードマップ

#### 短期対応（次リリース）
1. **入力処理問題の修正**
   - 優先度: 高
   - 予定: バグ修正リリース
   - 担当: UI開発チーム

#### 中期対応（2-3リリース後）
2. **レンダリング最適化**
   - 優先度: 中
   - 予定: パフォーマンス改善リリース
   - 実装方針: 段階的最適化

#### 長期対応（メジャーバージョン）
3. **セーブシステム改善**
   - 優先度: 中
   - 予定: アーキテクチャ改善時
   - 実装方針: 設計レベルでの見直し

### 課題報告・修正への貢献

#### バグ報告時の情報
1. **再現手順**: 具体的な操作順序
2. **環境情報**: OS、Python版本、依存関係版本
3. **ログ情報**: エラーメッセージ、デバッグログ
4. **期待動作**: 本来あるべき動作の説明

#### 修正への貢献
1. **Issue作成**: GitHubでの課題報告
2. **Pull Request**: 修正案の提案
3. **テスト**: 修正内容の検証
4. **ドキュメント**: 変更内容の文書化

## 変更履歴

このセクションでは、主要なバグ修正と機能改善の履歴を記録します。

### 2025-07-13: スタック可能アイテムの数量管理修正

#### 問題の概要
- **問題1**: インベントリでスタック可能アイテム（ヒーリングポーション等）を使用（u）すると、1個使用のつもりが全スタック（3個等）がインベントリから消失
- **問題2**: スタック可能アイテムをドロップ（d）すると、1個だけドロップされ、残りがインベントリに残存

#### 根本原因
- `Inventory.RemoveItem()`メソッドがスタック数量を考慮せず、アイテム全体を削除
- ドロップ処理が`RemoveItem()`をデフォルト引数（count=1）で呼び出し

#### 修正内容
1. **`internal/game/actor/inventory.go`**
   - `RemoveItem(item, count int)`に数量パラメータを追加
   - スタック可能アイテムの場合、count分だけStackCountを減算
   - StackCountが0以下になった場合のみアイテム削除

2. **`internal/ui/inventory_screen.go`**
   - ドロップ処理で全スタック削除するよう修正：`RemoveItem(item, item.StackCount)`
   - ドロップメッセージにスタック数を反映

#### テスト結果
- 全単体テスト: 成功
- 統合テスト: 全成功
- スタック機能の動作確認: 完全に正常化

#### 修正ファイル
```
internal/game/actor/inventory.go
internal/ui/inventory_screen.go
```
