---
cache_control: {"type": "ephemeral"}
---
# GoRogue - アーキテクチャ設計書

## 概要

GoRogueは、Go言語のシンプルさと並行処理の強みを活かした、モダンなソフトウェアアーキテクチャの原則に基づいて設計されたローグライクゲームです。責務分離、テスト可能性、拡張性、保守性を重視した設計により、高品質なゲーム体験と継続的な開発を可能にしています。

## アーキテクチャの基本原則

### 1. 責務分離 (Separation of Concerns)
- 各クラスが単一の責任を持つ
- ビジネスロジックとUIの分離
- データとロジックの分離

### 2. テスト可能性 (Testability)
- 依存関係の注入
- モックしやすい設計
- CLI/GUIモードの両方をサポート

### 3. 拡張性 (Extensibility)
- 新機能の追加が容易
- 既存機能の変更が他に影響しない
- プラグイン可能な設計

### 4. 保守性 (Maintainability)
- 明確な型ヒント
- 包括的なドキュメント
- 一貫したコーディング規約

## 全体アーキテクチャ

### レイヤー構成

```
┌─────────────────────────────────────────────────┐
│                   UI Layer                      │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │ MenuScreen  │  │ GameScreen  │  │ Other    │ │
│  │             │  │             │  │ Screens  │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                Business Logic Layer              │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │ GameLogic   │  │ Managers    │  │ Handlers │ │
│  │             │  │             │  │          │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                  Entity Layer                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │ Actors      │  │ Items       │  │ Magic    │ │
│  │             │  │             │  │ Traps    │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                   Data Layer                    │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │ Map/Tiles   │  │ Save Data   │  │ Config   │ │
│  │             │  │             │  │          │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
```

### 主要コンポーネント

#### 1. Core (コア)
**役割**: ゲームエンジン、状態管理、入力処理の中核
**場所**: `internal/game/`

```
game/
├── game.go               # メインゲームループ
├── state.go             # ゲーム状態管理
├── input.go             # 入力処理
├── world.go             # ワールド管理
├── save.go              # セーブ・ロード
├── score/               # スコアシステム
│   ├── score.go         # スコア計算
│   └── ranking.go       # ランキング管理
└── managers/            # 各種マネージャー
    ├── turn.go         # ターン管理
    ├── combat.go       # 戦闘管理
    └── ai.go           # AI管理
```

#### 2. Entities (エンティティ)
**役割**: ゲーム内オブジェクトの定義と管理
**場所**: `internal/entity/`

```
entity/
├── entity.go            # エンティティインターフェース
├── player.go            # プレイヤー実装
├── monster.go           # モンスター実装
├── item.go              # アイテム実装
├── inventory.go         # インベントリ管理
├── effects.go           # 状態効果システム
├── magic.go             # 魔法システム
└── trap.go              # トラップシステム
```

#### 3. Map (マップ)
**役割**: ダンジョン生成、タイル定義、階層管理
**場所**: `internal/dungeon/`

```
dungeon/
├── dungeon.go           # ダンジョン構造体
├── generator.go         # ダンジョン生成ロジック
├── floor.go             # フロア管理
├── tile.go              # タイル定義
├── room.go              # 部屋定義
├── corridor.go          # 通路生成
└── builders/            # 各種ビルダー
    ├── bsp.go          # BSPダンジョン生成
    ├── maze.go         # 迷路生成
    └── cave.go         # 洞窟生成
```

#### 4. UI (ユーザーインターフェース)
**役割**: 画面管理、描画処理、ユーザーインターフェース
**場所**: `internal/ui/`

```
ui/
├── screen.go            # 画面インターフェース
├── menu.go              # メインメニュー
├── game_view.go         # ゲームプレイ画面
├── inventory.go         # インベントリ画面
├── message.go           # メッセージログ
├── status.go            # ステータス表示
├── render.go            # 描画処理
├── input.go             # 入力処理
└── fov.go               # 視界計算
```

## 設計パターンの活用

### 1. Interface-based Design Pattern
**適用場所**: ゲーム全体のアーキテクチャ
**実装**: 各パッケージのインターフェース定義

Go言語の特性を活かし、インターフェースベースの設計により、疎結合で拡張性の高いアーキテクチャを実現します。

#### アーキテクチャ構造
```go
// Entity インターフェース
type Entity interface {
    GetPosition() (int, int)
    SetPosition(x, y int)
    GetChar() rune
    GetColor() color.Color
}

// Actor インターフェース（Entity を埋め込み）
type Actor interface {
    Entity
    GetHP() int
    TakeDamage(amount int)
    IsAlive() bool
}

// Renderer インターフェース
type Renderer interface {
    Draw(screen *ebiten.Image)
    Update() error
}
```

#### 利点
- **疎結合**: インターフェースによる依存関係の抽象化
- **拡張性**: 新しい実装の追加が容易
- **テスト性**: モックの作成が簡単
- **並行処理**: goroutineによる並列処理の活用

### 2. Builder Pattern
**適用場所**: ダンジョン生成システム
**実装**: `internal/dungeon/builders/`

```go
// DungeonBuilder インターフェース
type DungeonBuilder interface {
    Build(width, height int) *Dungeon
    SetSeed(seed int64)
    SetDifficulty(level int)
}

// BSPBuilder - BSPアルゴリズムによるダンジョン生成
type BSPBuilder struct {
    minRoomSize int
    maxRoomSize int
    splitRatio  float64
}

func (b *BSPBuilder) Build(width, height int) *Dungeon {
    dungeon := NewDungeon(width, height)
    
    // BSPツリーの構築
    root := b.splitSpace(0, 0, width, height)
    
    // 各ノードに部屋を生成
    b.createRooms(root, dungeon)
    
    // 部屋を通路で接続
    b.connectRooms(root, dungeon)
    
    return dungeon
}
```

**利点**:
- 複雑な生成プロセスを段階的に管理
- 生成パラメータの変更が容易
- 新しい生成ルールの追加が容易

### 3. Component Pattern
**適用場所**: エンティティシステム
**実装**: `internal/entity/`

```go
// Component インターフェース
type Component interface {
    Type() string
}

// PositionComponent - 位置情報
type PositionComponent struct {
    X, Y int
}

// CombatComponent - 戦闘能力
type CombatComponent struct {
    HP, MaxHP   int
    Attack      int
    Defense     int
}

// Entity with components
type Entity struct {
    ID         string
    components map[string]Component
}

func (e *Entity) AddComponent(c Component) {
    e.components[c.Type()] = c
}

func (e *Entity) GetComponent(cType string) Component {
    return e.components[cType]
}
```

**利点**:
- 複雑な処理を機能別に分割
- 個別のテストが容易
- 機能の追加・修正が局所的

### 4. State Pattern
**適用場所**: ゲーム状態管理
**実装**: `internal/game/state.go`

```go
// GameState インターフェース
type GameState interface {
    Enter(g *Game)
    Exit(g *Game)
    Update(g *Game) error
    Draw(g *Game, screen *ebiten.Image)
    HandleInput(g *Game, key ebiten.Key)
}

// MenuState - メニュー状態
type MenuState struct{}

func (s *MenuState) Enter(g *Game) {
    // メニュー初期化
}

func (s *MenuState) HandleInput(g *Game, key ebiten.Key) {
    switch key {
    case ebiten.KeyEnter:
        g.ChangeState(&PlayState{})
    case ebiten.KeyEscape:
        g.Exit()
    }
}

// PlayState - ゲームプレイ状態
type PlayState struct {
    turnManager *TurnManager
}
```

**利点**:
- 状態遷移の明確化
- 状態固有の処理を分離
- 新しい状態の追加が容易

### 5. Command Pattern
**適用場所**: アクションシステム
**実装**: `internal/game/action.go`

```go
// Action インターフェース
type Action interface {
    Execute(actor Actor, world *World) error
    CanExecute(actor Actor, world *World) bool
}

// MoveAction - 移動アクション
type MoveAction struct {
    DX, DY int
}

func (a *MoveAction) Execute(actor Actor, world *World) error {
    x, y := actor.GetPosition()
    newX, newY := x + a.DX, y + a.DY
    
    if !world.IsPassable(newX, newY) {
        return errors.New("cannot move to that position")
    }
    
    actor.SetPosition(newX, newY)
    return nil
}

// AttackAction - 攻撃アクション
type AttackAction struct {
    Target Actor
}

func (a *AttackAction) Execute(actor Actor, world *World) error {
    damage := calculateDamage(actor, a.Target)
    a.Target.TakeDamage(damage)
    return nil
}
```

**利点**:
- 効果の実行と定義を分離
- 新しい効果の追加が容易
- 効果の組み合わせが可能
- **GUIとCLIで統一されたコマンド処理**
- **コマンド実行環境の抽象化**

## 新実装システムアーキテクチャ

### 1. BSPダンジョン生成システム

**概要**: Binary Space Partitioningアルゴリズムによる自然なダンジョン生成

```go
type BSPNode struct {
    X, Y          int
    Width, Height int
    Left, Right   *BSPNode
    Room          *Room
}

type BSPBuilder struct {
    MinSize    int
    MaxSize    int
    SplitRatio float64
}

func (b *BSPBuilder) Build(width, height int) *Dungeon {
    dungeon := NewDungeon(width, height)
    
    // ルートノードから再帰的に分割
    root := &BSPNode{
        X: 0, Y: 0,
        Width: width, Height: height,
    }
    
    b.split(root, 0)
    b.createRooms(root, dungeon)
    b.connectRooms(root, dungeon)
    
    return dungeon
}

func (b *BSPBuilder) split(node *BSPNode, depth int) {
    if depth > MaxDepth || !b.shouldSplit(node) {
        return
    }
    
    // 縦横どちらで分割するか決定
    if node.Width > node.Height {
        b.splitVertically(node)
    } else {
        b.splitHorizontally(node)
    }
    
    // 子ノードを再帰的に分割
    b.split(node.Left, depth+1)
    b.split(node.Right, depth+1)
}
```

**主要特徴**:
- 再帰的空間分割による自然な部屋配置
- 部屋中心間のL字型通路接続
- 全部屋の接続保証

### 2. 高度なドア配置システム

**概要**: 戦術的に意味のある位置でのドア配置と重複防止

```go
type DoorSystem struct {
    doors map[Point]*Door
}

type Door struct {
    Position Point
    State    DoorState // Open, Closed, Locked, Hidden
}

func (ds *DoorSystem) PlaceDoors(dungeon *Dungeon) {
    for _, room := range dungeon.Rooms {
        ds.placeRoomDoors(room, dungeon)
    }
}

func (ds *DoorSystem) placeRoomDoors(room *Room, dungeon *Dungeon) {
    // 部屋の境界を調査
    boundaries := room.GetBoundaryPoints()
    
    for _, point := range boundaries {
        if ds.isValidDoorPosition(point, dungeon) {
            door := &Door{
                Position: point,
                State:    ds.randomDoorState(),
            }
            ds.doors[point] = door
            dungeon.SetTile(point.X, point.Y, TileDoor)
        }
    }
}

func (ds *DoorSystem) isValidDoorPosition(p Point, dungeon *Dungeon) bool {
    // 隣接8方向にドアがないかチェック
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            if dx == 0 && dy == 0 {
                continue
            }
            neighbor := Point{X: p.X + dx, Y: p.Y + dy}
            if _, exists := ds.doors[neighbor]; exists {
                return false
            }
        }
    }
    return true
}
```

**主要特徴**:
- 部屋境界突破箇所のみでドア配置
- 隣接8方向の重複ドア防止
- 確率的ドア状態（クローズド・オープン・隠し扉）

### 3. トラップシステム

**概要**: 探索・解除可能なトラップシステム

```go
type TrapSystem struct {
    traps map[Point]*Trap
}

type Trap struct {
    Position Point
    Type     TrapType
    Hidden   bool
    Disarmed bool
}

func (ts *TrapSystem) SearchTrap(actor Actor, x, y int) bool {
    // 隣接8方向のトラップを探索
    for dx := -1; dx <= 1; dx++ {
        for dy := -1; dy <= 1; dy++ {
            point := Point{X: x + dx, Y: y + dy}
            if trap, exists := ts.traps[point]; exists && trap.Hidden {
                // レベルに応じた成功率
                successRate := min(90, 40 + actor.GetLevel()*5)
                if rand.Intn(100) < successRate {
                    trap.Hidden = false
                    return true
                }
            }
        }
    }
    return false
}

func (ts *TrapSystem) DisarmTrap(actor Actor, x, y int) error {
    point := Point{X: x, Y: y}
    trap, exists := ts.traps[point]
    if !exists || trap.Hidden {
        return errors.New("no visible trap at this location")
    }
    
    // 70%の成功率
    if rand.Intn(100) < 70 {
        trap.Disarmed = true
        return nil
    }
    
    // 失敗時30%で発動
    if rand.Intn(100) < 30 {
        return trap.Trigger(actor)
    }
    
    return errors.New("failed to disarm trap")
}
```

**主要特徴**:
- 隣接8方向からの安全な探索・解除
- プレイヤーレベル依存の成功率
- 発見→解除の段階的処理

### 4. デバッグシステム

**概要**: 開発・テスト支援システム

```go
type DebugSystem struct {
    enabled     bool
    wizardMode  bool
    showFPS     bool
    showCoords  bool
}

func NewDebugSystem() *DebugSystem {
    return &DebugSystem{
        enabled: os.Getenv("DEBUG") == "true",
    }
}

func (ds *DebugSystem) ToggleWizardMode() {
    ds.wizardMode = !ds.wizardMode
}

func (ds *DebugSystem) Draw(screen *ebiten.Image, game *Game) {
    if !ds.enabled {
        return
    }
    
    if ds.showFPS {
        ebitenutil.DebugPrint(screen, fmt.Sprintf("FPS: %0.2f", ebiten.CurrentFPS()))
    }
    
    if ds.wizardMode {
        // 全マップ表示
        ds.drawFullMap(screen, game)
        // 隠し要素表示
        ds.drawHiddenElements(screen, game)
    }
    
    if ds.showCoords {
        x, y := game.Player.GetPosition()
        ebitenutil.DebugPrintAt(screen, fmt.Sprintf("Pos: (%d, %d)", x, y), 0, 20)
    }
}

func (ds *DebugSystem) IsInvincible() bool {
    return ds.wizardMode
}
```

**主要特徴**:
- 可視化（全マップ・隠し要素・エンティティ表示）
- 無敵機能（ダメージ・トラップ無効化）
- 操作機能（テレポート・レベルアップ・全回復・全探索）
- 環境変数連携（DEBUG=true自動有効化）

## データフロー

### 1. ゲームループのデータフロー

```
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│   Input     │───▶│  GameLogic  │───▶│   Render    │
│  Handler    │    │   Manager   │    │   System    │
└─────────────┘    └─────────────┘    └─────────────┘
       │                   │                   │
       │                   ▼                   │
       │            ┌─────────────┐            │
       │            │   Entities  │            │
       │            │  (Player,   │            │
       │            │ Monsters,   │            │
       │            │  Items)     │            │
       │            └─────────────┘            │
       │                   │                   │
       │                   ▼                   │
       │            ┌─────────────┐            │
       │            │   Dungeon   │            │
       │            │    Map      │            │
       │            └─────────────┘            │
       │                                       │
       └──────────────────────────────────────────┘
```

### 2. 戦闘システムのデータフロー

```
Player Action ─────┐
                   │
                   ▼
            ┌─────────────┐
            │   Combat    │
            │  Manager    │
            └─────────────┘
                   │
                   ▼
            ┌─────────────┐
            │   Damage    │
            │ Calculation │
            └─────────────┘
                   │
                   ▼
            ┌─────────────┐
            │   Status    │
            │  Effects    │
            └─────────────┘
                   │
                   ▼
            ┌─────────────┐
            │   Monster   │
            │    AI       │
            └─────────────┘
```

## 依存関係の管理

### 1. 依存関係の方向

```
UI Layer ──────────────────▶ Business Logic Layer
                                       │
                                       ▼
Business Logic Layer ──────────▶ Entity Layer
                                       │
                                       ▼
Entity Layer ──────────────────▶ Data Layer
```

### 2. 依存関係注入の例

```python
class GameScreen:
    def __init__(self, game_logic: GameLogic):
        self.game_logic = game_logic  # 依存関係注入

    def process_input(self, action: Action):
        # UIがビジネスロジックを呼び出す
        self.game_logic.process_action(action)
```

## テストアーキテクチャ

### 1. テストの層構造

```
┌─────────────────────────────────────────────────┐
│                Integration Tests                │
│  ゲーム全体のワークフローテスト                      │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                  Unit Tests                     │
│  個別クラス・メソッドのテスト                       │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                Property Tests                   │
│  確率的生成の検証テスト                           │
└─────────────────────────────────────────────────┘
```

### 2. テストダブルの活用

```python
class MockDungeon:
    """テスト用のダンジョンモック"""
    def __init__(self):
        self.width = 80
        self.height = 50
        self.tiles = self.create_test_tiles()

class TestCombatManager:
    def test_combat_calculation(self):
        # モックを使用したテスト
        mock_player = MockPlayer(attack=10, defense=5)
        mock_monster = MockMonster(attack=8, defense=3)

        combat_manager = CombatManager()
        result = combat_manager.calculate_damage(mock_player, mock_monster)

        assert result == 5  # 10 - 5 = 5
```

## 性能に関する考慮

### 1. 描画最適化

```go
type Renderer struct {
    tileCache   map[Point]*ebiten.Image
    dirtyRegion []Point
}

func (r *Renderer) Draw(screen *ebiten.Image) {
    // ビューポート内のタイルのみ描画
    viewport := r.calculateViewport()
    
    for y := viewport.MinY; y <= viewport.MaxY; y++ {
        for x := viewport.MinX; x <= viewport.MaxX; x++ {
            if r.isDirty(x, y) {
                r.drawTile(screen, x, y)
            }
        }
    }
    
    r.clearDirtyRegion()
}
```

### 2. 並行処理の活用

```go
type Game struct {
    updateCh chan UpdateMsg
    renderCh chan RenderMsg
}

func (g *Game) Run() error {
    // 更新と描画を別goroutineで処理
    go g.updateLoop()
    go g.renderLoop()
    
    return ebiten.RunGame(g)
}

func (g *Game) updateLoop() {
    ticker := time.NewTicker(time.Second / 60)
    defer ticker.Stop()
    
    for {
        select {
        case <-ticker.C:
            g.update()
        case msg := <-g.updateCh:
            g.handleUpdate(msg)
        }
    }
}
```

## イベント駆動アーキテクチャ

### 概要
GoRogueは、イベント駆動アーキテクチャを採用し、ゲーム内の各種イベントを非同期で処理します。これにより、レスポンシブなゲーム体験と拡張性の高いシステムを実現します。

### イベントシステム構成

```go
// Event インターフェース
type Event interface {
    Type() EventType
    Timestamp() time.Time
}

// EventBus - イベントの配信システム
type EventBus struct {
    subscribers map[EventType][]EventHandler
    eventQueue  chan Event
    mu          sync.RWMutex
}

func (eb *EventBus) Subscribe(eventType EventType, handler EventHandler) {
    eb.mu.Lock()
    defer eb.mu.Unlock()
    
    eb.subscribers[eventType] = append(eb.subscribers[eventType], handler)
}

func (eb *EventBus) Publish(event Event) {
    select {
    case eb.eventQueue <- event:
    default:
        // Queue full, drop event or handle overflow
    }
}

func (eb *EventBus) processEvents() {
    for event := range eb.eventQueue {
        eb.mu.RLock()
        handlers := eb.subscribers[event.Type()]
        eb.mu.RUnlock()
        
        for _, handler := range handlers {
            go handler.Handle(event)
        }
    }
}
```

### 主要イベントタイプ

```go
type EventType int

const (
    EventMove EventType = iota
    EventAttack
    EventItemPickup
    EventItemUse
    EventDoorOpen
    EventTrapTriggered
    EventLevelUp
    EventDeath
    EventFloorChange
)

// MoveEvent - 移動イベント
type MoveEvent struct {
    Actor    Actor
    FromX, FromY int
    ToX, ToY     int
    timestamp    time.Time
}

// CombatEvent - 戦闘イベント
type CombatEvent struct {
    Attacker Actor
    Defender Actor
    Damage   int
    Critical bool
    timestamp time.Time
}
```

#### 2. CommonCommandHandler (共通処理レイヤー)
```python
class CommonCommandHandler:
    """GUIとCLIで共通のコマンド処理"""
    def handle_command(self, command: str, args: list[str]) -> CommandResult:
        # 統一されたコマンド処理ロジック
        if command in ["move", "north", "south", "east", "west"]:
            return self._handle_move_command(command, args)
        elif command in ["get", "pickup", "g"]:
            return self._handle_get_item()
        # ... 他のコマンド処理
```

#### 3. 実装クラス
- **CLICommandContext**: CLI環境での実装
- **GUICommandContext**: GUI環境での実装

### スコアシステム

```go
type ScoreSystem struct {
    currentScore int
    multiplier   float64
    events       []ScoreEvent
}

type ScoreEvent struct {
    Description string
    Points      int
    Timestamp   time.Time
}

func (s *ScoreSystem) AddScore(event ScoreEvent) {
    points := int(float64(event.Points) * s.multiplier)
    s.currentScore += points
    s.events = append(s.events, event)
}

// スコア計算ルール
func (s *ScoreSystem) CalculateScore(player *Player) int {
    score := 0
    
    // 基本スコア
    score += player.Gold * 10
    score += player.Level * 1000
    score += player.ExploredRooms * 50
    
    // ボーナススコア
    if player.HasAmulet {
        score *= 2
    }
    
    // 時間ボーナス
    if player.TurnCount < 10000 {
        score += (10000 - player.TurnCount) * 5
    }
    
    return score
}
```

### キー入力の統一化

#### GUI環境でのキー→コマンド変換
```python
def _key_to_command(self, event: tcod.event.KeyDown) -> str | None:
    """キーイベントをコマンド文字列に変換"""
    key = event.sym

    # viキー + 矢印キー対応
    if key in (ord('h'), tcod.event.KeySym.LEFT):
        return "west"
    elif key in (ord('j'), tcod.event.KeySym.DOWN):
        return "south"
    # ... 他のキーマッピング
```

### 利点

1. **一貫性**: GUIとCLIで同じコマンドセット
2. **保守性**: コマンド処理の共通化によりバグ修正が一箇所で完了
3. **拡張性**: 新しいコマンドの追加が両環境で自動適用
4. **テスト性**: 共通ロジックの単一テストで両環境をカバー

## 拡張性の設計

### 1. プラグイン可能な設計

```go
// Plugin インターフェース
type Plugin interface {
    Name() string
    Version() string
    Initialize(game *Game) error
    Shutdown() error
}

// ItemEffect インターフェース
type ItemEffect interface {
    Apply(target Actor, world *World) error
    GetDescription() string
}

// 新しい効果の追加例
type TeleportEffect struct {
    Range int
}

func (e *TeleportEffect) Apply(target Actor, world *World) error {
    // ランダムな位置にテレポート
    x, y := world.GetRandomPassablePosition()
    target.SetPosition(x, y)
    return nil
}
```

### 2. 設定システム

```go
type Config struct {
    Game    GameConfig    `json:"game"`
    Display DisplayConfig `json:"display"`
    Debug   DebugConfig   `json:"debug"`
}

type GameConfig struct {
    DungeonDepth      int     `json:"dungeon_depth"`
    PlayerStartingHP  int     `json:"player_starting_hp"`
    MonsterSpawnRate  float64 `json:"monster_spawn_rate"`
    DifficultyLevel   int     `json:"difficulty_level"`
}

func LoadConfig(path string) (*Config, error) {
    data, err := os.ReadFile(path)
    if err != nil {
        return nil, err
    }
    
    var config Config
    if err := json.Unmarshal(data, &config); err != nil {
        return nil, err
    }
    
    return &config, nil
}
```

## セキュリティ・セーフティの考慮

### 1. 型安全性

Go言語の静的型付けにより、コンパイル時に型エラーを検出できます。

```go
// 型安全なエンティティシステム
type EntityID uint64

type EntityManager struct {
    entities map[EntityID]Entity
    nextID   EntityID
    mu       sync.RWMutex
}

func (em *EntityManager) CreateEntity() EntityID {
    em.mu.Lock()
    defer em.mu.Unlock()
    
    id := em.nextID
    em.nextID++
    return id
}

func (em *EntityManager) GetEntity(id EntityID) (Entity, bool) {
    em.mu.RLock()
    defer em.mu.RUnlock()
    
    entity, exists := em.entities[id]
    return entity, exists
}
```

### 2. エラーハンドリング

```go
// カスタムエラー型
type GameError struct {
    Code    string
    Message string
    Err     error
}

func (e *GameError) Error() string {
    if e.Err != nil {
        return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Err)
    }
    return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// エラー定義
var (
    ErrInvalidMove   = &GameError{Code: "INVALID_MOVE", Message: "cannot move to that position"}
    ErrItemNotFound  = &GameError{Code: "ITEM_NOT_FOUND", Message: "item not found"}
    ErrInsufficientMP = &GameError{Code: "INSUFFICIENT_MP", Message: "not enough MP"}
)

// エラーハンドリング例
func (g *Game) MovePlayer(dx, dy int) error {
    newX, newY := g.Player.X + dx, g.Player.Y + dy
    
    if !g.World.IsPassable(newX, newY) {
        return ErrInvalidMove
    }
    
    g.Player.SetPosition(newX, newY)
    return nil
}
```

## 開発・運用の支援

### 1. ログ・デバッグ

```go
type Logger struct {
    debugMode bool
    logger    *log.Logger
}

func NewLogger(debugMode bool) *Logger {
    return &Logger{
        debugMode: debugMode,
        logger:    log.New(os.Stdout, "[GOROGUE] ", log.LstdFlags),
    }
}

func (l *Logger) Debug(format string, args ...interface{}) {
    if l.debugMode {
        l.logger.Printf("[DEBUG] "+format, args...)
    }
}

func (l *Logger) LogCombat(attacker, target string, damage int) {
    l.Debug("Combat: %s → %s (%d damage)", attacker, target, damage)
}
```

### 2. プロファイリング

```go
type Profiler struct {
    enabled bool
    timers  map[string]*Timer
    mu      sync.Mutex
}

type Timer struct {
    start   time.Time
    total   time.Duration
    count   int
}

func (p *Profiler) StartTimer(name string) func() {
    if !p.enabled {
        return func() {}
    }
    
    start := time.Now()
    return func() {
        p.mu.Lock()
        defer p.mu.Unlock()
        
        timer, exists := p.timers[name]
        if !exists {
            timer = &Timer{}
            p.timers[name] = timer
        }
        
        timer.total += time.Since(start)
        timer.count++
    }
}

func (p *Profiler) Report() string {
    p.mu.Lock()
    defer p.mu.Unlock()
    
    var report strings.Builder
    for name, timer := range p.timers {
        avg := timer.total / time.Duration(timer.count)
        report.WriteString(fmt.Sprintf("%s: %v (avg: %v, count: %d)\n", 
            name, timer.total, avg, timer.count))
    }
    return report.String()
}
```

## UIシステム詳細設計

### アーキテクチャ概要

GoRogueのUIシステムは、**Ebitenゲームエンジン**を基盤とした2Dグラフィックスシステムを採用しています。ピクセルパーフェクトな描画と高パフォーマンスを実現し、レトロな雰囲気を保ちながらモダンなゲーム体験を提供します。

### 主要設計原則

1. **状態ベースの画面管理**: GameStates列挙型による明確な状態遷移
2. **コンポーネント化**: 単一責務の原則に基づく機能分離
3. **統一された入力処理**: Vi-keys、矢印キー、テンキーの包括的サポート
4. **レスポンシブ描画**: ウィンドウサイズに適応する動的レイアウト

### 画面システム

#### Screen インターフェース設計

```go
// Screen インターフェース
type Screen interface {
    Update() error
    Draw(screen *ebiten.Image)
    HandleInput(key ebiten.Key)
    IsTransparent() bool  // 背景を透過するか
}

// BaseScreen - 画面の基本実装
type BaseScreen struct {
    game       *Game
    width      int
    height     int
    tileSize   int
}

func (s *BaseScreen) calculateScreenSize() (int, int) {
    return s.width * s.tileSize, s.height * s.tileSize
}
```

**設計思想**:
- 最小限のインターフェース定義
- 各画面の独立性確保
- エンジンへの参照による状態アクセス

#### 画面構成

| 画面 | 責務 | 主要機能 |
|------|------|----------|
| **MenuScreen** | メインメニュー | タイトル表示、メニュー選択 |
| **GameScreen** | ゲームプレイ | マップ描画、ステータス表示、メッセージログ |
| **InventoryScreen** | インベントリ | アイテム一覧、装備管理 |
| **GameOverScreen** | ゲームオーバー | スコア表示、統計情報 |
| **VictoryScreen** | 勝利画面 | 最終スコア、クリア時間 |

### 状態管理システム

#### GameStates列挙型

```python
class GameStates(Enum):
    MENU = auto()                # メインメニュー表示中
    PLAYERS_TURN = auto()        # プレイヤーの入力待ち
    ENEMY_TURN = auto()          # 敵の行動処理中
    PLAYER_DEAD = auto()         # プレイヤー死亡時の処理
    GAME_OVER = auto()           # ゲームオーバー画面表示
    VICTORY = auto()             # ゲーム勝利画面表示
    SHOW_INVENTORY = auto()      # インベントリ一覧表示
    DROP_INVENTORY = auto()      # アイテム破棄モード
    SHOW_MAGIC = auto()          # 魔法一覧表示
    TARGETING = auto()           # ターゲット選択モード
    DIALOGUE = auto()            # NPC対話状態
    LEVEL_UP = auto()           # レベルアップ時の選択
    CHARACTER_SCREEN = auto()    # キャラクター情報表示
    EXIT = auto()               # ゲーム終了シグナル
```

#### ScreenManager

```go
type ScreenManager struct {
    screens      map[string]Screen
    activeScreen Screen
    screenStack  []Screen  // 画面のスタック（オーバーレイ用）
}

func (sm *ScreenManager) Push(screen Screen) {
    sm.screenStack = append(sm.screenStack, screen)
    sm.activeScreen = screen
}

func (sm *ScreenManager) Pop() {
    if len(sm.screenStack) > 1 {
        sm.screenStack = sm.screenStack[:len(sm.screenStack)-1]
        sm.activeScreen = sm.screenStack[len(sm.screenStack)-1]
    }
}

func (sm *ScreenManager) Update() error {
    // 透過スクリーンの場合、下層も更新
    for _, screen := range sm.screenStack {
        if err := screen.Update(); err != nil {
            return err
        }
    }
    return nil
}

func (sm *ScreenManager) Draw(screen *ebiten.Image) {
    // 透過スクリーンの場合、下層から描画
    for i, s := range sm.screenStack {
        if i == 0 || s.IsTransparent() {
            s.Draw(screen)
        }
    }
}
```

**利点**:
- 状態遷移の一元管理
- 入力処理の責務分離
- エスケープキーのフォールバック処理

### UIコンポーネントシステム

#### 1. GameRenderer（描画システム）

```python
class GameRenderer:
    """ゲーム画面の描画処理を担当するクラス"""

    def render(self, console: tcod.Console) -> None:
        """レイヤー化描画"""
        console.clear()
        self._render_map(console)      # マップ層
        self._render_status(console)   # ステータス層
        self._render_messages(console) # メッセージ層
```

**主要機能**:
- **レイヤー化描画**: マップ→ステータス→メッセージの順序描画
- **FOV統合**: 可視/探索済み状態による動的描画制御
- **マップオフセット**: ステータス行を考慮した座標調整
- **エンティティ描画**: アイテム、モンスター、NPCの統合描画

**色彩システム**:
```python
# 視界状態による色彩変化
color = (130, 110, 50) if visible else (0, 0, 100)  # 壁
color = (192, 192, 192) if visible else (64, 64, 64)  # 床
```

#### 2. InputHandler（入力処理システム）

```python
class InputHandler:
    """入力処理システムの管理クラス"""

    def handle_key(self, event: tcod.event.KeyDown) -> None:
        if self.targeting_mode:
            self._handle_targeting_key(event)
        else:
            self._handle_normal_key(event)
```

**入力マッピング**:
```go
type InputMapper struct {
    keyBindings map[ebiten.Key]Action
}

func NewInputMapper() *InputMapper {
    return &InputMapper{
        keyBindings: map[ebiten.Key]Action{
            // Vi-keys
            ebiten.KeyH: &MoveAction{DX: -1, DY: 0},
            ebiten.KeyJ: &MoveAction{DX: 0, DY: 1},
            ebiten.KeyK: &MoveAction{DX: 0, DY: -1},
            ebiten.KeyL: &MoveAction{DX: 1, DY: 0},
            ebiten.KeyY: &MoveAction{DX: -1, DY: -1},
            ebiten.KeyU: &MoveAction{DX: 1, DY: -1},
            ebiten.KeyB: &MoveAction{DX: -1, DY: 1},
            ebiten.KeyN: &MoveAction{DX: 1, DY: 1},
            // 矢印キー
            ebiten.KeyArrowLeft:  &MoveAction{DX: -1, DY: 0},
            ebiten.KeyArrowRight: &MoveAction{DX: 1, DY: 0},
            ebiten.KeyArrowUp:    &MoveAction{DX: 0, DY: -1},
            ebiten.KeyArrowDown:  &MoveAction{DX: 0, DY: 1},
            // アクションキー
            ebiten.KeyG: &PickupAction{},
            ebiten.KeyI: &OpenInventoryAction{},
            ebiten.KeyO: &OpenDoorAction{},
            ebiten.KeyS: &SearchAction{},
        },
    }
}
```

**特殊機能**:
- **ターゲット選択モード**: 魔法詠唱時のターゲット指定
- **修飾キー対応**: Ctrl+S/L（セーブ・ロード）、Shift+./,（階段）
- **周囲検索**: ドア開閉、隠し扉探索、トラップ解除の8方向検索

#### 3. FOVシステム（視界計算）

```go
type FOVSystem struct {
    visible     [][]bool
    explored    [][]bool
    lightRadius int
}

func (fov *FOVSystem) Calculate(world *World, x, y int) {
    // すべてのタイルを非表示に
    fov.clearVisible()
    
    // Shadowcasting アルゴリズム
    for octant := 0; octant < 8; octant++ {
        fov.castLight(world, x, y, 1, 1.0, 0.0, octant)
    }
    
    // プレイヤー位置は常に可視
    fov.visible[y][x] = true
    fov.explored[y][x] = true
}

func (fov *FOVSystem) castLight(world *World, cx, cy, row int, 
                                start, end float64, octant int) {
    if start < end {
        return
    }
    
    for j := row; j <= fov.lightRadius; j++ {
        dx := -j - 1
        dy := -j
        blocked := false
        newStart := 0.0
        
        for dx <= 0 {
            dx++
            // 座標変換（octantに応じて）
            x, y := fov.transformOctant(cx, cy, dx, dy, octant)
            
            if !world.InBounds(x, y) {
                continue
            }
            
            // 視線の角度を計算
            leftSlope := (float64(dx) - 0.5) / (float64(dy) + 0.5)
            rightSlope := (float64(dx) + 0.5) / (float64(dy) - 0.5)
            
            if start < rightSlope {
                continue
            } else if end > leftSlope {
                break
            }
            
            // タイルを可視に設定
            fov.visible[y][x] = true
            fov.explored[y][x] = true
            
            if blocked {
                if world.BlocksSight(x, y) {
                    newStart = rightSlope
                    continue
                } else {
                    blocked = false
                    start = newStart
                }
            } else {
                if world.BlocksSight(x, y) && j < fov.lightRadius {
                    blocked = true
                    fov.castLight(world, cx, cy, j+1, start, leftSlope, octant)
                    newStart = rightSlope
                }
            }
        }
        
        if blocked {
            break
        }
    }
}
```

**効果的FOV半径計算**:
```python
def _calculate_effective_fov_radius(self, x: int, y: int) -> int:
    # 基本半径: 8
    # 暗い部屋での制限: 2-3
    # 光源アイテム使用時: 基本半径復帰
    return dark_room_builder.get_visibility_range_at(
        x, y, rooms, has_light, light_radius
    )
```

#### 4. SaveLoadManager（状態永続化システム）

```python
class SaveLoadManager:
    """セーブ・ロード処理の管理クラス"""

    def save_game(self) -> bool:
        save_data = self._create_save_data()
        return self.save_manager.save_game(save_data)

    def _create_save_data(self) -> dict[str, Any]:
        return {
            "player": self._serialize_player(player),
            "inventory": self._serialize_inventory(inventory),
            "current_floor": dungeon_manager.current_floor,
            "floor_data": dungeon_manager.floor_data,
            "message_log": game_logic.message_log,
        }
```

**シリアライゼーション機能**:
- プレイヤー状態の完全保存
- インベントリ・装備情報の詳細保存
- フロアデータの遅延読み込み対応
- メッセージログの継続性確保

### TCODライブラリ統合

#### Ebitenの初期化と設定

```go
type Game struct {
    screenWidth  int
    screenHeight int
    tileSize     int
    font         *Font
    renderer     *Renderer
}

func NewGame() *Game {
    g := &Game{
        screenWidth:  80,
        screenHeight: 50,
        tileSize:     16,
    }
    
    // フォント読み込み
    g.font = LoadBitmapFont("assets/fonts/terminal16x16.png")
    
    // レンダラー初期化
    g.renderer = NewRenderer(g.font, g.tileSize)
    
    return g
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
    return g.screenWidth * g.tileSize, g.screenHeight * g.tileSize
}

func main() {
    game := NewGame()
    
    ebiten.SetWindowSize(1280, 800)
    ebiten.SetWindowTitle("GoRogue")
    ebiten.SetWindowResizable(true)
    
    if err := ebiten.RunGame(game); err != nil {
        log.Fatal(err)
    }
}
```

#### ウィンドウリサイズ対応

```python
def handle_resize(self, event: tcod.event.WindowEvent) -> None:
    pixel_width = getattr(event, "width", 800)
    pixel_height = getattr(event, "height", 600)

    # 文字数計算
    self.screen_width = max(MIN_SCREEN_WIDTH, pixel_width // font_width)
    self.screen_height = max(MIN_SCREEN_HEIGHT, pixel_height // font_height)

    # コンソール再作成
    self.console = tcod.console.Console(self.screen_width, self.screen_height)

    # 各画面のコンソール参照更新
    self.menu_screen.update_console(self.console)
    self.game_screen.update_console(self.console)
```

### パフォーマンス最適化

#### 1. 描画最適化

- **差分描画**: 変更されたタイルのみ更新
- **FOVベース描画**: 視界外のエンティティ描画を省略
- **レイヤー分離**: マップ、エンティティ、UIの独立レンダリング

#### 2. メモリ効率

- **遅延生成**: フロアデータの必要時生成
- **状態キャッシュ**: 探索済みエリアの効率的格納
- **リソース管理**: 不要なコンソール参照の適切な解放

### ユーザビリティ設計

#### 1. アクセシビリティ

- **多様な入力方式**: Vi-keys、矢印キー、テンキーの包括サポート
- **色彩コントラスト**: 視認性を考慮した色彩選択
- **レスポンシブレイアウト**: 画面サイズに適応するUI要素配置

#### 2. 操作の一貫性

- **共通ナビゲーション**: 全画面での矢印キー+Enterパターン
- **ESCキーの統一**: 一段階戻る動作の一貫性
- **ヘルプ表示**: ?キーによるコンテキストヘルプ

#### 3. 視覚的フィードバック

- **選択状態の明示**: ハイライト表示による現在選択位置の明確化
- **装備状態表示**: 装備中アイテムの視覚的区別
- **ステータス色分け**: HP/MP残量による色彩変化

### 既知の技術的課題

#### 1. 入力処理の修正中問題
- **場所**: `src/pyrogue/ui/components/input_handler.py`
- **問題**: キーボード入力処理の一部で不具合
- **影響**: 特定のキー組み合わせで期待通りの動作がしない

#### 2. 大規模マップでのレンダリングパフォーマンス
- **問題**: マップサイズ拡大時の描画処理負荷
- **対策候補**: 視界ベースのカリング、タイル描画最適化

#### 3. 複雑なゲーム状態のシリアライゼーション
- **問題**: セーブデータの一貫性保証
- **課題**: フロアデータ、エンティティ状態の完全復元

## まとめ

GoRogueのアーキテクチャは、以下の要素を統合することで、高品質なゲーム体験と継続的な開発を可能にしています：

1. **Go言語の特性活用**: 静的型付け、並行処理、高速実行
2. **インターフェースベース設計**: 疎結合で拡張性の高いアーキテクチャ
3. **実証済みの設計パターン**: Builder、Component、State、Commandパターンの適切な活用
4. **イベント駆動アーキテクチャ**: 非同期処理とレスポンシブな操作性
5. **テスト可能な設計**: インターフェースによるモックの容易さ
6. **性能最適化**: 効率的な描画、並行処理、メモリ管理

この設計により、GoRogueは単なるゲームプロジェクトではなく、Go言語でのゲーム開発のベストプラクティスを示す包括的な教材としても機能することを目指しています。
