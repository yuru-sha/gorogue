# Core コンポーネント

GoRogueのゲームエンジンとコアシステム。ゲームループ、状態管理、入力処理、セーブ・ロード機能を統合管理します。

## 概要

`internal/core/`は、GoRogueの心臓部となるゲームエンジンシステムです。GUI/CLIの両対応、ターンベース制御、責務分離されたマネージャーアーキテクチャにより、堅牢で拡張性の高いゲームループを実現しています。

## アーキテクチャ

### ファイル構成

```
core/
├── engine.go              # メインゲームエンジン (GUI)
├── cli_engine.go          # CLIテスト用エンジン
├── game_logic.go          # ゲームロジック統合管理
├── game_states.go         # ゲーム状態定義
├── input_handlers.go      # 入力処理システム
├── save_manager.go        # セーブ・ロード機能
├── score_manager.go       # スコアランキング
├── command_handler.go     # 共通コマンドハンドラー (v0.1.0)
├── auto_explore_handler.go # 自動探索ハンドラー (v0.1.0)
├── debug_command_handler.go # デバッグコマンドハンドラー (v0.1.0)
├── save_load_handler.go   # セーブ・ロードハンドラー (v0.1.0)
├── info_command_handler.go # 情報表示ハンドラー (v0.1.0)
└── managers/              # 専門マネージャー群
    ├── game_context.go    # 共有コンテキスト
    ├── turn_manager.go    # ターン制御
    ├── combat_manager.go  # 戦闘システム
    ├── monster_ai_manager.go # モンスターAI
    ├── movement_manager.go   # 移動処理
    ├── item_manager.go      # アイテム管理
    └── floor_manager.go     # フロア管理
```

### 設計原則

- **Handler Pattern**: 機能別専用ハンドラーによる責務分離（v0.1.0で導入）
- **責務分離**: 各マネージャーが単一責任を持つ
- **依存関係注入**: GameContextによる統一的な依存管理
- **状態管理**: 明確なゲーム状態とその遷移
- **テスタビリティ**: CLI対応による自動テスト可能性

## 主要コンポーネント

### エンジンシステム

#### Engine (GUI版エンジン)

GoRogueのメインゲームエンジン。フルグラフィカルUIとリアルタイム操作を提供。

**主要機能:**
- **レンダリングエンジン**: gruid使用のフルグラフィカルUI
- **イベント処理**: リアルタイムなキー入力とウィンドウイベント処理
- **画面管理**: 複数画面状態の統合管理
- **リサイズ対応**: 動的ウィンドウサイズ変更サポート

**ゲームループ実装:**
```go
func (e *Engine) MainLoop() {
    // メインゲームループ
    for e.running {
        e.console.Clear()
        e.renderCurrentState()
        e.console.Flush()
        e.processEvents()
        e.updateGameState()
    }
}
```

#### CLIEngine (テスト・自動化用エンジン)

テストと自動化に特化したコマンドライン版エンジン。

**特徴:**
- **テキストベース**: コマンド入力による操作
- **自動化対応**: スクリプト実行可能な設計
- **簡素化UI**: 標準出力によるステータス表示
- **デバッグ機能**: 開発用デバッグコマンド豊富

**使用例:**
```go
cliEngine := NewCLIEngine()
cliEngine.StartNewGame()
cliEngine.ExecuteCommand("move south")
cliEngine.ExecuteCommand("get item")
```

### ゲーム状態管理

#### GameStates (game_states.go)

ゲーム全体の状態を型安全に定義。

```go
type GameState int

const (
    MENU GameState = iota           // メニュー画面
    PLAYERS_TURN                    // プレイヤーターン
    ENEMY_TURN                      // 敵ターン
    PLAYER_DEAD                     // プレイヤー死亡
    GAME_OVER                       // ゲームオーバー
    VICTORY                         // 勝利
    SHOW_INVENTORY                  // インベントリ表示
    DROP_INVENTORY                  // アイテムドロップ
    SHOW_MAGIC                      // 魔法画面
    TARGETING                       // ターゲット選択
    DIALOGUE                        // 対話
    LEVEL_UP                        // レベルアップ
    CHARACTER_SCREEN                // キャラクター画面
    EXIT                            // 終了
)
```

### 統合管理システム

#### GameLogic (game_logic.go)

ゲーム全体のビジネスロジックを統合管理する調整役（Coordinator）。

**設計転換:**
- **リファクタリング前**: モノリシックな巨大構造体
- **リファクタリング後**: 責務分離された専門マネージャーの統合

**統合管理機能:**
```go
type GameLogic struct {
    movementManager   *MovementManager
    combatManager     *CombatManager
    itemManager       *ItemManager
    floorManager      *FloorManager
    // ... 他のマネージャー
}

func NewGameLogic(gameContext *GameContext) *GameLogic {
    return &GameLogic{
        movementManager: NewMovementManager(gameContext),
        combatManager:   NewCombatManager(gameContext),
        itemManager:     NewItemManager(gameContext),
        floorManager:    NewFloorManager(gameContext),
        // ... 他のマネージャー
    }
}
```

### コマンド処理システム（Handler Pattern - v0.1.0）

#### CommonCommandHandler (command_handler.go)

機能別専用ハンドラーによる責務分離型コマンド処理システム。

**Handler Pattern アーキテクチャ:**
```
CommonCommandHandler (コア)
├── AutoExploreHandler     # 自動探索機能
├── DebugCommandHandler    # デバッグコマンド
├── SaveLoadHandler        # セーブ・ロード処理
└── InfoCommandHandler     # 情報表示機能
```

**実装例:**
```go
type CommonCommandHandler struct {
    context           *CommandContext
    // 遅延初期化によるメモリ効率化
    autoExploreHandler *AutoExploreHandler
    debugHandler      *DebugCommandHandler
    saveLoadHandler   *SaveLoadHandler
    infoHandler       *InfoCommandHandler
}

func NewCommonCommandHandler(context *CommandContext) *CommonCommandHandler {
    return &CommonCommandHandler{
        context: context,
    }
}

func (h *CommonCommandHandler) HandleCommand(command string, args []string) CommandResult {
    switch command {
    case "auto_explore", "O":
        return h.getAutoExploreHandler().HandleAutoExplore()
    case "debug":
        return h.getDebugHandler().HandleDebugCommand(args)
    // ...
    }
}

func (h *CommonCommandHandler) getAutoExploreHandler() *AutoExploreHandler {
    // 自動探索ハンドラーを取得（遅延初期化）
    if h.autoExploreHandler == nil {
        h.autoExploreHandler = NewAutoExploreHandler(h.context)
    }
    return h.autoExploreHandler
}
```

**利点:**
- **責務分離**: 各機能を専用ハンドラーに分離
- **拡張性**: 新機能は新ハンドラー追加で対応
- **保守性**: 修正範囲が明確に限定される
- **テスト性**: 各ハンドラーを独立してテスト可能
- **再利用性**: CLI/GUIで同一ハンドラーを共有

#### 専用ハンドラー群

**AutoExploreHandler**: 自動探索機能
- 未探索エリアの検出
- 安全ルート計算
- 敵発見時の自動停止

**DebugCommandHandler**: デバッグコマンド
- イェンダーのアミュレット付与
- 階層テレポート
- HP・ダメージ調整

**SaveLoadHandler**: セーブ・ロード処理
- IDベースセーブシステム
- 後方互換性維持
- エラーハンドリング

**InfoCommandHandler**: 情報表示機能
- シンボル説明
- アイテム識別状況
- キャラクター詳細

### 入力処理システム

#### InputHandlers (input_handlers.go)

GUI/CLI両対応の統一入力処理システム。

**アーキテクチャ:**
- **StateManager**: ゲーム状態別入力処理
- **CommonCommandHandler**: 共通コマンド処理
- **Strategy Pattern**: 状態別ハンドリング

**実装例:**
```go
func (h *InputHandler) HandleInput(key gruid.Key) *Action {
    // 状態に応じた入力処理
    currentState := h.gameLogic.GameState

    switch currentState {
    case PLAYERS_TURN:
        return h.handlePlayerTurn(key)
    case SHOW_INVENTORY:
        return h.handleInventory(key)
    // ... 他の状態処理
    }
    return nil
}
```

### 永続化システム

#### SaveManager (save_manager.go)

Permadeathシステムに対応したセーブ・ロード機能。

**重要機能:**
```go
type SaveManager struct {
    // セーブマネージャーの実装
}

func (s *SaveManager) SaveGame(gameContext *GameContext) error {
    // ゲーム状態の保存
    // SHA256チェックサムによる整合性保証
    // メイン/バックアップの二重保存
    return nil
}

func (s *SaveManager) DeleteSaveOnDeath() error {
    // Permadeath: 死亡時のセーブデータ削除
    // 真のローグライクゲーム体験の実現
    return nil
}
```

**Permadeath実装:**
- 死亡時の自動セーブデータ削除
- 改ざん検出によるチート防止
- バックアップ機能による安全性

#### ScoreManager (score_manager.go)

ランキングシステムとゲーム記録管理。

**記録項目:**
- プレイヤー名、レベル、階層到達度
- ゴールド獲得量、モンスター撃破数
- ゲーム結果（勝利/死亡）とその詳細

**機能:**
```go
type ScoreEntry struct {
    PlayerName     string
    Level          int
    Floor          int
    Gold           int
    MonstersKilled int
    Result         string
    Timestamp      time.Time
}

func (s *ScoreManager) RecordScore(gameContext *GameContext, deathReason string, victory bool) error {
    // スコア記録
    scoreEntry := ScoreEntry{
        PlayerName:     gameContext.Player.Name,
        Level:          gameContext.Player.Level,
        Floor:          gameContext.DungeonManager.CurrentFloor,
        Gold:           gameContext.Player.Gold,
        MonstersKilled: gameContext.Player.MonstersKilled,
        Timestamp:      time.Now(),
    }
    
    if victory {
        scoreEntry.Result = "Victory"
    } else {
        scoreEntry.Result = fmt.Sprintf("Died: %s", deathReason)
    }
    
    return s.saveScore(scoreEntry)
}
```

## Manager アーキテクチャ

### GameContext (共有コンテキスト)

全マネージャー間の共有データハブ。依存関係注入のコンテナ役割。

```go
type GameContext struct {
    // ゲーム全体の共有コンテキスト
    Player         *Player
    Inventory      *Inventory
    DungeonManager *DungeonManager
    MessageLog     *MessageLog
    GameState      GameState
    TurnCount      int
}

func NewGameContext(player *Player, inventory *Inventory, dungeonManager *DungeonManager, messageLog *MessageLog) *GameContext {
    return &GameContext{
        Player:         player,
        Inventory:      inventory,
        DungeonManager: dungeonManager,
        MessageLog:     messageLog,
        GameState:      PLAYERS_TURN,
        TurnCount:      0,
    }
}
```

### 専門マネージャー群

#### TurnManager (ターン制御)

**責務:**
- ターン進行制御
- ステータス異常処理
- 満腹度システム更新
- MP自然回復処理

**ターン処理フロー:**
```go
func (t *TurnManager) ExecuteTurn() error {
    // 1ターンの実行
    t.processStatusEffects()    // 状態異常処理
    t.updateHungerSystem()      // 満腹度更新
    t.processMPRegeneration()   // MP回復
    t.checkWinCondition()       // 勝利条件確認
    t.incrementTurnCounter()    // ターン数増加
    return nil
}
```

#### CombatManager (戦闘システム)

**責務:**
- 戦闘処理とダメージ計算
- レベルアップシステム
- モンスタードロップ処理
- 死亡判定

#### MonsterAIManager (モンスターAI)

**責務:**
- AI行動決定
- 視界判定
- 特殊攻撃実行
- 分裂処理

**AI行動パターン:**
```go
func (m *MonsterAIManager) ExecuteMonsterAI(monster *Monster) error {
    // モンスターAI実行
    if m.canSeePlayer(monster) {
        if m.isAdjacentToPlayer(monster) {
            return m.attackPlayer(monster)
        } else {
            return m.moveTowardsPlayer(monster)
        }
    } else {
        return m.randomMovement(monster)
    }
}
```

#### MovementManager (移動処理)

**責務:**
- 移動処理と衝突判定
- 移動後イベント（アイテム自動取得等）
- 階段移動トリガー

#### ItemManager (アイテム管理)

**責務:**
- アイテム取得・使用・装備
- ドロップ処理
- 呪われたアイテム制御

#### FloorManager (フロア管理)

**責務:**
- 階層移動処理
- 扉操作（開閉）
- 隠し扉探索
- トラップ処理

## ゲームループ詳細

### メインゲームループ (Engine)

```go
func (e *Engine) MainLoop() {
    // 統合ゲームループ
    for e.running {
        // 1. 画面クリア
        e.rootConsole.Clear()

        // 2. 現在状態のレンダリング
        e.renderCurrentState()

        // 3. 画面更新
        e.rootConsole.Flush()

        // 4. イベント処理
        e.processEvents()

        // 5. ゲーム状態更新
        e.updateGameState()
    }
}
```

### ターン制御フロー

```
Player Action Input
        ↓
Action Validation
        ↓
Execute Player Action
        ↓
Process Status Effects
        ↓
Execute Monster Turns
        ↓
Update Systems (Hunger, MP)
        ↓
Check Win/Lose Conditions
        ↓
Increment Turn Counter
        ↓
[Loop continues]
```

## 依存関係注入とテスタビリティ

### 依存関係注入パターン

```go
type GameLogic struct {
    gameContext     *GameContext
    turnManager     *TurnManager
    combatManager   *CombatManager
    movementManager *MovementManager
}

func NewGameLogic(gameContext *GameContext) *GameLogic {
    // 依存関係の注入
    return &GameLogic{
        gameContext:     gameContext,
        // 各マネージャーに共通コンテキストを注入
        turnManager:     NewTurnManager(gameContext),
        combatManager:   NewCombatManager(gameContext),
        movementManager: NewMovementManager(gameContext),
    }
}
```

### テスタビリティ向上策

**CLIEngine活用:**
```go
func TestCombatSystem(t *testing.T) {
    // 戦闘システムのテスト
    cliEngine := NewCLIEngine()
    cliEngine.StartNewGame()

    // モンスターとの戦闘をシミュレート
    cliEngine.ExecuteCommand("move north")  // モンスターに近づく
    cliEngine.ExecuteCommand("attack")      // 攻撃実行

    // 結果の検証
    assert.Greater(t, cliEngine.GameLogic.GameContext.Player.Health, 0)
}
```

**モック対応:**
```go
// インターフェース定義による抽象化
type GameContextInterface interface {
    GetPlayer() *Player
    GetDungeonManager() *DungeonManager
}

// テストでのモック利用
func TestMovementManager(t *testing.T) {
    mockContext := createMockGameContext()
    movementManager := NewMovementManager(mockContext)
    // テスト実行...
}
```

## 使用パターン

### 基本的なゲーム開始

```go
import "github.com/yuru-sha/gorogue/internal/core"

// ゲームエンジンの初期化
engine := core.NewEngine()

// ゲームループ開始
engine.Run()
```

### CLIモードでのテスト

```go
import "github.com/yuru-sha/gorogue/internal/core"

// CLIエンジンでの自動テスト
cli := core.NewCLIEngine()
cli.StartNewGame()

// 自動的なゲームプレイ
commands := []string{"move north", "get gold", "move east", "attack"}
for _, command := range commands {
    cli.ExecuteCommand(command)
}
```

### マネージャーの個別利用

```go
// 戦闘システムの直接利用
combatManager := NewCombatManager(gameContext)
damage := combatManager.CalculateDamage(attacker, defender)
combatManager.ApplyDamage(defender, damage)
```

## 拡張ガイド

### 新しいハンドラーの追加（Handler Pattern）

1. **ハンドラークラスの作成**
```python
class NewFeatureHandler:
    def __init__(self, context: CommandContext):
        self.context = context

    def handle_new_feature(self, args: list[str]) -> CommandResult:
        """新機能の処理"""
        # 機能の実装
        self.context.add_message("New feature executed!")
        return CommandResult(True)
```

2. **CommonCommandHandlerへの統合**
```python
class CommonCommandHandler:
    def __init__(self, context: CommandContext) -> None:
        # 既存ハンドラー...
        self._new_feature_handler = None

    def handle_command(self, command: str, args: list[str] | None = None) -> CommandResult:
        # 既存コマンド...
        if command == "new_feature":
            return self._get_new_feature_handler().handle_new_feature(args)

    def _get_new_feature_handler(self):
        """新機能ハンドラーを取得（遅延初期化）"""
        if self._new_feature_handler is None:
            self._new_feature_handler = NewFeatureHandler(self.context)
        return self._new_feature_handler
```

### 新しいマネージャーの追加

1. **マネージャークラスの作成**
```python
class NewFeatureManager:
    def __init__(self, game_context: GameContext):
        self.game_context = game_context

    def execute_new_feature(self) -> None:
        """新機能の実行"""
        pass
```

2. **GameLogicへの統合**
```python
class GameLogic:
    def __init__(self, game_context: GameContext):
        # 既存マネージャー...
        self.new_feature_manager = NewFeatureManager(game_context)
```

### 新しいゲーム状態の追加

```python
class GameStates(Enum):
    # 既存状態...
    NEW_STATE = auto()
```

### カスタムコマンドの追加

```python
def handle_custom_command(self, command: str) -> bool:
    """カスタムコマンドの処理"""
    if command == "custom_action":
        self._execute_custom_action()
        return True
    return False
```

## Handler Pattern 実装詳細

### CommandContext設計

Handler Pattern実装において、`CommandContext`はハンドラー間の共通インターフェースとして重要な役割を果たします。

```python
@dataclass
class CommandContext:
    """ハンドラー間で共有されるコマンド実行コンテキスト"""
    game_context: GameContext
    engine: Optional['Engine'] = None
    cli_engine: Optional['CLIEngine'] = None

    def add_message(self, message: str) -> None:
        """メッセージログに追加"""
        self.game_context.message_log.add_message(message)

    def is_gui_mode(self) -> bool:
        """GUI モードかどうかの判定"""
        return self.engine is not None

    def is_cli_mode(self) -> bool:
        """CLI モードかどうかの判定"""
        return self.cli_engine is not None
```

### CommandResult統一

すべてのハンドラーは一貫したレスポンス形式を返します：

```python
@dataclass
class CommandResult:
    """コマンド実行結果"""
    success: bool
    message: Optional[str] = None
    turn_consumed: bool = False
    game_state_change: Optional[GameStates] = None

    @classmethod
    def success_with_turn(cls, message: str = None) -> 'CommandResult':
        """ターン消費する成功結果"""
        return cls(success=True, message=message, turn_consumed=True)

    @classmethod
    def failure(cls, message: str) -> 'CommandResult':
        """失敗結果"""
        return cls(success=False, message=message)
```

### ハンドラー間通信

ハンドラー間でのデータ共有は、`CommandContext`を通じて行います：

```python
class SaveLoadHandler:
    def handle_save(self) -> CommandResult:
        """セーブ処理"""
        if self._save_game():
            # 他のハンドラーが参照可能な状態を更新
            self.context.game_context.last_save_time = datetime.now()
            return CommandResult.success("Game saved successfully")
        return CommandResult.failure("Failed to save game")

    def _save_game(self) -> bool:
        """実際のセーブ処理"""
        save_manager = SaveManager()
        return save_manager.save_game(self.context.game_context)
```

### エラーハンドリング統一

```python
class BaseHandler:
    """ハンドラーの基底クラス"""

    def __init__(self, context: CommandContext):
        self.context = context
        self.logger = get_logger(self.__class__.__name__)

    def safe_execute(self, func: Callable, *args, **kwargs) -> CommandResult:
        """安全な実行ラッパー"""
        try:
            return func(*args, **kwargs)
        except Exception as e:
            self.logger.error(f"Handler error: {e}")
            return CommandResult.failure(f"Internal error: {str(e)}")
```

### テスト戦略

Handler Patternのテストは、各ハンドラーを独立してテスト可能にします：

```python
class TestAutoExploreHandler:
    def test_auto_explore_basic(self):
        """自動探索の基本動作テスト"""
        # Arrange
        mock_context = create_mock_context()
        handler = AutoExploreHandler(mock_context)

        # Act
        result = handler.handle_auto_explore()

        # Assert
        assert result.success
        assert result.turn_consumed
        assert "Exploring" in result.message

    def test_auto_explore_enemy_detection(self):
        """敵発見時の自動停止テスト"""
        # Arrange
        mock_context = create_mock_context_with_enemy()
        handler = AutoExploreHandler(mock_context)

        # Act
        result = handler.handle_auto_explore()

        # Assert
        assert result.success
        assert not result.turn_consumed  # 敵発見時はターン消費しない
        assert "Enemy detected" in result.message
```

## パフォーマンス最適化

### 効率的なレンダリング

- 差分更新による描画最適化
- 視界外オブジェクトの描画スキップ
- コンソールバッファリングの活用

### メモリ管理

- オブジェクトプールパターンの活用
- 不要なオブジェクトの早期解放
- 循環参照の回避

## トラブルシューティング

### よくある問題

**状態遷移の問題:**
```python
# 状態が正しく設定されているか確認
if game_logic.game_state != expected_state:
    logger.warning(f"Unexpected state: {game_logic.game_state}")
```

**マネージャー間の依存関係:**
```python
# GameContextの正しい共有を確認
assert all_managers_share_same_context(game_context)
```

### デバッグ支援

```python
# ゲーム状態のダンプ
def dump_game_state(game_context: GameContext) -> dict:
    """デバッグ用のゲーム状態出力"""
    return {
        "player_pos": (game_context.player.x, game_context.player.y),
        "health": game_context.player.health,
        "level": game_context.player.level,
        "current_floor": game_context.dungeon_manager.current_floor,
        "game_state": game_context.game_state.name
    }
```

## まとめ

Core コンポーネントは、GoRogueプロジェクトの中核として以下の価値を提供します：

- **統合管理**: 複雑なゲームロジックの整理された統合
- **責務分離**: 保守性と拡張性を高める明確な役割分担
- **テスタビリティ**: CLI対応による包括的な自動テスト
- **状態管理**: 型安全で明確なゲーム状態制御
- **永続化**: Permadeathとランキングによる本格ローグライク体験

この設計により、オリジナルRogueの複雑なゲームメカニクスを現代的なソフトウェア設計パターンで実現し、高い品質と保守性を両立したゲームエンジンを提供しています。
