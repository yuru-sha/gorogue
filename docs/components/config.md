# Config コンポーネント

GoRogueの設定管理システム。環境変数とゲーム設定の統合管理を担当します。

## 概要

`internal/config/`は、現代的な環境変数管理と後方互換性を両立した設定システムです。`.env`ファイルによる外部設定、型安全なアクセスAPI、レガシーシステムとの統合を提供します。

## アーキテクチャ

### ファイル構成

```
config/
├── config.go        # 統合APIの提供
├── env.go           # 現代的な環境変数管理
└── legacy.go        # 後方互換性維持
```

### 設計原則

- **型安全性**: 環境変数の型変換とバリデーション
- **後方互換性**: 既存のCONFIGインターフェースの維持
- **責務分離**: 環境変数管理とゲーム設定の分離
- **拡張性**: 新しい設定項目の容易な追加
- **Handler Pattern連携**: v0.1.0のHandler Patternとの統合設計

## 主要コンポーネント

### EnvConfig 構造体 (env.go)

環境変数の読み込みと型安全なアクセスを提供する中核構造体。

#### 機能

**自動的な.envファイル探索**
```go
func (e *EnvConfig) LoadEnv(envFile string) error {
    // .envファイルを自動探索・読み込み
    // 指定がない場合、現在ディレクトリから親に向かって探索
    return nil
}
```

**型安全な値取得API**
```go
func (e *EnvConfig) GetBool(key string, defaultValue bool) bool {
    // 真偽値として取得（true/1/yes/onを真として認識）
    return defaultValue
}

func (e *EnvConfig) GetInt(key string, defaultValue int) int {
    // 整数値として取得（変換エラー時はdefault返却）
    return defaultValue
}

func (e *EnvConfig) GetFloat(key string, defaultValue float64) float64 {
    // 浮動小数点値として取得（変換エラー時はdefault返却）
    return defaultValue
}
```

#### 実装例

```go
import "github.com/yuru-sha/gorogue/internal/config"

// 環境変数の読み込み
envConfig := config.NewEnvConfig()
if err := envConfig.LoadEnv(""); err != nil {
    log.Fatal(err)
}

// 型安全なアクセス
debugMode := envConfig.GetBool("DEBUG", false)
windowWidth := envConfig.GetInt("WINDOW_WIDTH", 80)
logLevel := envConfig.Get("LOG_LEVEL", "INFO")
```

### アクセサー関数

設定項目ごとの専用アクセサー関数を提供し、タイポ防止と一元管理を実現。

```go
func GetDebugMode() bool {
    // デバッグモードの設定を取得
    return envConfig.GetBool("DEBUG", false)
}

func GetLogLevel() string {
    // ログレベルの設定を取得
    return envConfig.Get("LOG_LEVEL", "INFO")
}

func GetAutoSaveEnabled() bool {
    // オートセーブ機能の設定を取得
    return envConfig.GetBool("AUTO_SAVE_ENABLED", true)
}
```

### レガシー設定 (legacy.go)

後方互換性を維持するため、既存の構造体とグローバル`CONFIG`インスタンスを提供。

```go
type GameConfig struct {
    // メインゲーム設定
    Display DisplayConfig
    Player  PlayerConfig
    Monster MonsterConfig
    Item    ItemConfig
}

// グローバル設定インスタンス
var CONFIG = GameConfig{
    Display: DefaultDisplayConfig(),
    Player:  DefaultPlayerConfig(),
    Monster: DefaultMonsterConfig(),
    Item:    DefaultItemConfig(),
}
```

## 対応設定項目

### 環境変数 (.env)

| 変数名 | 型 | デフォルト値 | 説明 |
|--------|----|-----------:|------|
| `DEBUG` | bool | false | デバッグモード |
| `LOG_LEVEL` | str | INFO | ログレベル |
| `SAVE_DIRECTORY` | str | saves | セーブディレクトリ |
| `AUTO_SAVE_ENABLED` | bool | true | オートセーブ |
| `FONT_PATH` | str | auto | フォントファイルパス |

### ゲーム定数

`legacy.py`を通じて以下の定数群を管理：

- **CombatConstants**: 戦闘関連の定数
- **GameConstants**: ゲーム全般の定数
- **HungerConstants**: 飢餓システムの定数
- **ItemConstants**: アイテム関連の定数
- **ProbabilityConstants**: 確率値の定数

## 使用パターン

### 基本的な使用方法

```go
import (
    "github.com/yuru-sha/gorogue/internal/config"
    "fmt"
)

// 環境設定の初期化
envConfig := config.NewEnvConfig()
if err := envConfig.LoadEnv(""); err != nil {
    log.Fatal(err)
}

// 型安全なアクセス
if config.GetDebugMode() {
    fmt.Printf("Debug mode enabled, auto save: %t\n", config.GetAutoSaveEnabled())
    fmt.Printf("Log level: %s\n", config.GetLogLevel())
}
```

### レガシーAPIの継続利用

```go
import "github.com/yuru-sha/gorogue/internal/config"

// 既存コードとの互換性
displayConfig := config.CONFIG.Display
playerConfig := config.CONFIG.Player
```

### .envファイル設定例

```bash
# .env
DEBUG=true
LOG_LEVEL=DEBUG
AUTO_SAVE_ENABLED=false
```

## エラー処理

### 型変換エラーの安全な処理

```go
// 不正な値が設定されていてもクラッシュしない
windowWidth := envConfig.GetInt("WINDOW_WIDTH", 80)  // 変換エラー時は80を返却
debugMode := envConfig.GetBool("DEBUG", false)       // 不正値時はfalseを返却
```

### ファイル不在時の処理

```go
// .envファイルが存在しなくてもエラーにならない
if err := envConfig.LoadEnv(""); err != nil {
    // エラーログを出力するがデフォルト値で動作継続
    log.Printf("Warning: %v, using default values", err)
}
```

## テスト戦略

### 単体テストでの活用

```go
func TestEnvConfig(t *testing.T) {
    // 環境設定のテスト
    config := NewEnvConfig()

    // モック環境変数での検証
    os.Setenv("DEBUG", "true")
    os.Setenv("WINDOW_WIDTH", "100")
    defer os.Unsetenv("DEBUG")
    defer os.Unsetenv("WINDOW_WIDTH")

    err := config.LoadEnv("")
    assert.NoError(t, err)
    assert.True(t, config.GetBool("DEBUG", false))
    assert.Equal(t, 100, config.GetInt("WINDOW_WIDTH", 80))
}
```

## 拡張ガイド

### 新しい環境変数の追加

1. **アクセサー関数の定義**
```go
func GetNewSetting() string {
    // 新しい設定項目の取得
    return envConfig.Get("NEW_SETTING", "default_value")
}
```

2. **.env.exampleの更新**
```bash
NEW_SETTING=default_value
```

3. **ドキュメントの更新**
上記の対応設定項目テーブルに追加

### レガシー設定の拡張

```go
type NewConfig struct {
    // 新しい設定カテゴリ
    NewOption string
}

func DefaultNewConfig() NewConfig {
    return NewConfig{
        NewOption: "default",
    }
}

type GameConfig struct {
    // 既存設定...
    NewCategory NewConfig
}
```

## 技術的特徴

### 現代的なGo機能

- **ゼロ値**: 構造体の安全な初期化
- **構造体**: 設定構造の簡潔な定義
- **filepath**: ファイルパス操作の標準的な手法

### 依存関係

- **github.com/joho/godotenv**: .envファイル読み込み
- **filepath**: ファイルパス操作
- **strconv**: 型変換機能

### パフォーマンス特性

- **遅延読み込み**: 必要時にのみ.envファイルを読み込み
- **キャッシュ**: 一度読み込んだ設定値はos.environに保存
- **軽量**: 最小限の依存関係とメモリ使用量

## Handler Pattern連携（v0.1.0）

### Handler Patternでの設定活用

各Handlerは環境設定を適切に参照し、機能の可用性を制御します：

```go
type DebugCommandHandler struct {
    context      *CommandContext
    debugEnabled bool
}

func NewDebugCommandHandler(context *CommandContext) *DebugCommandHandler {
    return &DebugCommandHandler{
        context:      context,
        debugEnabled: GetDebugMode(),
    }
}

func (h *DebugCommandHandler) HandleDebugCommand(args []string) CommandResult {
    // デバッグコマンド処理（設定依存）
    if !h.debugEnabled {
        return CommandResult{Success: false, Message: "Debug mode is disabled"}
    }

    // デバッグ機能の実行
    return h.executeDebugAction(args)
}
```

### 設定ベースの機能制御

```go
type SaveLoadHandler struct {
    context *CommandContext
}

func (h *SaveLoadHandler) HandleAutoSave() CommandResult {
    // オートセーブ処理（設定依存）
    if !GetAutoSaveEnabled() {
        return CommandResult{Success: true, Message: "Auto-save disabled"}
    }

    // オートセーブの実行
    return h.performAutoSave()
}
```

### ハンドラー初期化時の設定注入

```go
type CommonCommandHandler struct {
    context      *CommandContext
    debugHandler *DebugCommandHandler
}

func NewCommonCommandHandler(context *CommandContext) *CommonCommandHandler {
    h := &CommonCommandHandler{
        context: context,
    }
    // 設定値に基づくハンドラー初期化制御
    h.initHandlersBasedOnConfig()
    return h
}

func (h *CommonCommandHandler) initHandlersBasedOnConfig() {
    // 設定に基づくハンドラー初期化
    if GetDebugMode() {
        h.debugHandler = NewDebugCommandHandler(h.context)
    } else {
        h.debugHandler = nil
    }
}
```

## まとめ

Config コンポーネントは、GoRogueプロジェクトの設定管理において以下の価値を提供します：

- **開発効率**: .envファイルによる環境依存設定の外部化
- **型安全性**: 実行時エラーを防ぐ型安全なAPI
- **保守性**: 後方互換性を保ちながらの段階的移行
- **拡張性**: 新しい設定項目の容易な追加
- **Handler Pattern統合**: v0.1.0のHandler Patternとの完全な連携

この設計により、開発・テスト・本番環境での設定管理が統一され、プロジェクトの成長に対応できる柔軟なシステムを実現しています。
