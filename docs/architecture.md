---
cache_control: {"type": "ephemeral"}
---
# GoRogue - アーキテクチャ設計書

## 概要

GoRogueは、Go言語の特徴を活かしたシンプルで効率的なソフトウェアアーキテクチャの原則に基づいて設計されるローグライクゲームです。責務分離、テスト可能性、拡張性、保守性を重視した設計により、高品質なゲーム体験と継続的な開発を可能にします。

## アーキテクチャの基本原則

### 1. シンプルさ (Simplicity)
- Goらしいシンプルで読みやすい設計
- 複雑な抽象化を避け、直感的な構造
- 最小限の依存関係

### 2. 責務分離 (Separation of Concerns)
- 各パッケージが単一の責任を持つ
- ビジネスロジックとUIの分離
- データとロジックの分離

### 3. テスト可能性 (Testability)
- 依存関係の適切な管理
- インターフェースによる抽象化
- モックしやすい設計

### 4. 拡張性 (Extensibility)
- 新機能の追加が容易
- 既存機能の変更が他に影響しない
- プラグイン可能な設計

### 5. 保守性 (Maintainability)
- 明確な型定義
- 包括的なドキュメント
- 一貫したコーディング規約

## 全体アーキテクチャ

### レイヤー構成

```
┌─────────────────────────────────────────────────┐
│                 UI Layer                        │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │   Renderer  │  │    Input    │  │   View   │ │
│  │             │  │   Handler   │  │          │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│               Game Logic Layer                   │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │   Engine    │  │   Player    │  │  World   │ │
│  │             │  │             │  │          │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                 Entity Layer                    │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │  Monsters   │  │    Items    │  │  Traps   │ │
│  │             │  │             │  │          │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
┌─────────────────────────────────────────────────┐
│                  Data Layer                     │
│  ┌─────────────┐  ┌─────────────┐  ┌──────────┐ │
│  │  Map/Tiles  │  │  Save Data  │  │  Config  │ │
│  │             │  │             │  │          │ │
│  └─────────────┘  └─────────────┘  └──────────┘ │
└─────────────────────────────────────────────────┘
```

### 主要コンポーネント

#### 1. Game Engine (ゲームエンジン)
**役割**: ゲームループ、状態管理、イベント処理の中核
**場所**: `internal/game/engine.go`

```go
type Engine struct {
    state     GameState
    player    *Player
    world     *World
    ui        *UIManager
    running   bool
}

func (e *Engine) Run() error {
    for e.running {
        e.handleInput()
        e.update()
        e.render()
    }
    return nil
}
```

#### 2. Player (プレイヤー)
**役割**: プレイヤーの状態管理、移動処理
**場所**: `internal/game/player.go`

```go
type Player struct {
    position   Position
    stats      Stats
    inventory  *Inventory
    level      int
    experience int
}

func (p *Player) Move(dx, dy int) bool {
    // 移動処理
    return true
}
```

#### 3. World (ワールド)
**役割**: マップ、ダンジョン生成、エンティティ管理
**場所**: `internal/game/world.go`

```go
type World struct {
    currentMap *Map
    floors     []*Floor
    generator  *DungeonGenerator
}

func (w *World) GenerateFloor(depth int) *Floor {
    // ダンジョン生成処理
    return nil
}
```

#### 4. UI Manager (UI管理)
**役割**: レンダリング、入力処理、ユーザーインターフェース
**場所**: `internal/ui/`

```go
type UIManager struct {
    renderer *Renderer
    input    *InputHandler
    model    *gruid.Model
}

func (ui *UIManager) Render() {
    // Gruidを使用した描画処理
}
```

## 設計パターンの活用

### 1. MVC Pattern (Model-View-Controller)
**適用場所**: ゲーム全体の構造
**実装**: Gruidライブラリの設計に準拠

```go
// Model: ゲーム状態
type GameModel struct {
    player *Player
    world  *World
    state  GameState
}

// View: 描画処理
func (m *GameModel) Draw(gd gruid.Grid) {
    // 描画ロジック
}

// Controller: 入力処理
func (m *GameModel) Update(msg gruid.Msg) gruid.Effect {
    switch msg := msg.(type) {
    case gruid.MsgKeyDown:
        // 入力処理
    }
    return nil
}
```

### 2. State Pattern
**適用場所**: ゲーム状態管理
**実装**: 状態別の処理分離

```go
type GameState int

const (
    StateMenu GameState = iota
    StateGame
    StateGameOver
    StateInventory
)

type StateManager struct {
    current GameState
    states  map[GameState]StateHandler
}

type StateHandler interface {
    Enter()
    Update(msg gruid.Msg) gruid.Effect
    Exit()
}
```

### 3. Component Pattern
**適用場所**: エンティティシステム
**実装**: 再利用可能なコンポーネント

```go
type Entity struct {
    id         int
    components map[string]Component
}

type Component interface {
    Update(entity *Entity, dt float64)
}

type PositionComponent struct {
    X, Y int
}

type HealthComponent struct {
    Current, Max int
}
```

### 4. Observer Pattern
**適用場所**: イベント処理
**実装**: イベントの発行と購読

```go
type EventManager struct {
    listeners map[string][]func(Event)
}

func (em *EventManager) Subscribe(eventType string, handler func(Event)) {
    em.listeners[eventType] = append(em.listeners[eventType], handler)
}

func (em *EventManager) Publish(event Event) {
    for _, handler := range em.listeners[event.Type] {
        handler(event)
    }
}
```

## Gruidライブラリ統合

### 基本構造

```go
import "github.com/anaseto/gruid"

type Game struct {
    model *GameModel
    app   *gruid.App
}

func NewGame() *Game {
    model := &GameModel{
        // 初期化
    }
    
    app := gruid.NewApp(gruid.AppConfig{
        Model: model,
        Quit:  gruid.QuitOnCtrlC,
    })
    
    return &Game{
        model: model,
        app:   app,
    }
}

func (g *Game) Run() error {
    return g.app.Start(gruid.DefaultConfig())
}
```

### 描画システム

```go
func (m *GameModel) Draw(gd gruid.Grid) {
    // マップの描画
    m.drawMap(gd)
    
    // エンティティの描画
    m.drawEntities(gd)
    
    // UIの描画
    m.drawUI(gd)
}

func (m *GameModel) drawMap(gd gruid.Grid) {
    for y := 0; y < m.world.height; y++ {
        for x := 0; x < m.world.width; x++ {
            tile := m.world.GetTile(x, y)
            cell := gruid.Cell{
                Rune:  tile.Rune,
                Style: tile.Style,
            }
            gd.Set(x, y, cell)
        }
    }
}
```

### 入力処理

```go
func (m *GameModel) Update(msg gruid.Msg) gruid.Effect {
    switch msg := msg.(type) {
    case gruid.MsgKeyDown:
        return m.handleKeyDown(msg.Key)
    case gruid.MsgMouse:
        return m.handleMouse(msg)
    case gruid.MsgFrame:
        return m.handleFrame()
    }
    return nil
}

func (m *GameModel) handleKeyDown(key gruid.Key) gruid.Effect {
    switch key {
    case gruid.KeyArrowUp, 'k':
        m.player.Move(0, -1)
    case gruid.KeyArrowDown, 'j':
        m.player.Move(0, 1)
    case gruid.KeyArrowLeft, 'h':
        m.player.Move(-1, 0)
    case gruid.KeyArrowRight, 'l':
        m.player.Move(1, 0)
    }
    return nil
}
```

## データ構造とアルゴリズム

### 1. マップ表現

```go
type Tile struct {
    Rune      rune
    Style     gruid.Style
    Walkable  bool
    Opaque    bool
    Explored  bool
    Visible   bool
}

type Map struct {
    width  int
    height int
    tiles  [][]Tile
}

func (m *Map) GetTile(x, y int) *Tile {
    if x < 0 || x >= m.width || y < 0 || y >= m.height {
        return nil
    }
    return &m.tiles[y][x]
}
```

### 2. ダンジョン生成

```go
type DungeonGenerator struct {
    width    int
    height   int
    maxRooms int
    roomSize struct{ min, max int }
}

func (dg *DungeonGenerator) Generate() *Map {
    dungeonMap := NewMap(dg.width, dg.height)
    
    // 部屋の生成
    rooms := dg.generateRooms()
    
    // 部屋の配置
    for _, room := range rooms {
        dg.carveRoom(dungeonMap, room)
    }
    
    // 通路の生成
    dg.connectRooms(dungeonMap, rooms)
    
    return dungeonMap
}
```

### 3. 視界システム (FOV)

```go
type FOVCalculator struct {
    viewRange int
}

func (fov *FOVCalculator) Calculate(gameMap *Map, px, py int) {
    // シャドウキャスティングアルゴリズム
    for octant := 0; octant < 8; octant++ {
        fov.castShadow(gameMap, px, py, octant)
    }
}

func (fov *FOVCalculator) castShadow(gameMap *Map, px, py, octant int) {
    // 視界計算の実装
}
```

## 並行処理の活用

### 1. ゲームループの最適化

```go
func (e *Engine) Run() error {
    ticker := time.NewTicker(time.Second / 60) // 60 FPS
    defer ticker.Stop()
    
    for e.running {
        select {
        case <-ticker.C:
            e.update()
            e.render()
        case event := <-e.eventChan:
            e.handleEvent(event)
        }
    }
    return nil
}
```

### 2. 非同期処理

```go
func (e *Engine) processAI() {
    go func() {
        for {
            select {
            case <-e.aiTicker.C:
                e.updateMonsters()
            case <-e.done:
                return
            }
        }
    }()
}
```

## エラーハンドリング

### 1. Goらしいエラー処理

```go
func (p *Player) Move(dx, dy int) error {
    newX, newY := p.position.X+dx, p.position.Y+dy
    
    if !p.world.IsValidPosition(newX, newY) {
        return fmt.Errorf("invalid position: (%d, %d)", newX, newY)
    }
    
    if !p.world.IsWalkable(newX, newY) {
        return fmt.Errorf("position blocked: (%d, %d)", newX, newY)
    }
    
    p.position.X, p.position.Y = newX, newY
    return nil
}
```

### 2. パニック処理

```go
func (e *Engine) Run() (err error) {
    defer func() {
        if r := recover(); r != nil {
            err = fmt.Errorf("game panic: %v", r)
        }
    }()
    
    return e.gameLoop()
}
```

## テストアーキテクチャ

### 1. 単体テスト

```go
func TestPlayerMove(t *testing.T) {
    player := &Player{
        position: Position{X: 5, Y: 5},
        world:    createTestWorld(),
    }
    
    err := player.Move(1, 0)
    assert.NoError(t, err)
    assert.Equal(t, Position{X: 6, Y: 5}, player.position)
}
```

### 2. 統合テスト

```go
func TestGameLoop(t *testing.T) {
    engine := NewEngine()
    
    // テスト用の初期状態設定
    engine.LoadTestState()
    
    // 複数のアクションを実行
    engine.ProcessInput("move_north")
    engine.ProcessInput("attack")
    
    // 期待される状態を検証
    assert.Equal(t, StateGame, engine.state)
}
```

### 3. ベンチマーク

```go
func BenchmarkFOVCalculation(b *testing.B) {
    gameMap := generateTestMap(100, 100)
    fov := &FOVCalculator{viewRange: 8}
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        fov.Calculate(gameMap, 50, 50)
    }
}
```

## 性能に関する考慮

### 1. メモリ最適化

```go
// オブジェクトプールの使用
var cellPool = sync.Pool{
    New: func() interface{} {
        return &gruid.Cell{}
    },
}

func getCell() *gruid.Cell {
    return cellPool.Get().(*gruid.Cell)
}

func putCell(cell *gruid.Cell) {
    cellPool.Put(cell)
}
```

### 2. 描画最適化

```go
type Renderer struct {
    dirtyRegions []Region
    lastFrame    [][]gruid.Cell
}

func (r *Renderer) Render(gd gruid.Grid) {
    // 変更された領域のみを再描画
    for _, region := range r.dirtyRegions {
        r.renderRegion(gd, region)
    }
    r.dirtyRegions = r.dirtyRegions[:0]
}
```

## 設定管理

### 1. 設定構造

```go
type Config struct {
    Display struct {
        Width      int    `json:"width"`
        Height     int    `json:"height"`
        Fullscreen bool   `json:"fullscreen"`
        Font       string `json:"font"`
    } `json:"display"`
    
    Game struct {
        Difficulty  int  `json:"difficulty"`
        AutoSave    bool `json:"auto_save"`
        ShowTutorial bool `json:"show_tutorial"`
    } `json:"game"`
}
```

### 2. 設定読み込み

```go
func LoadConfig(filename string) (*Config, error) {
    data, err := os.ReadFile(filename)
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

## セーブ・ロードシステム

### 1. ゲーム状態のシリアライゼーション

```go
type SaveData struct {
    Version  int            `json:"version"`
    Player   *PlayerSave    `json:"player"`
    World    *WorldSave     `json:"world"`
    GameTime int64          `json:"game_time"`
}

func (g *Game) Save(filename string) error {
    saveData := &SaveData{
        Version:  1,
        Player:   g.player.ToSave(),
        World:    g.world.ToSave(),
        GameTime: g.gameTime,
    }
    
    data, err := json.Marshal(saveData)
    if err != nil {
        return err
    }
    
    return os.WriteFile(filename, data, 0644)
}
```

## 拡張性の設計

### 1. プラグインシステム

```go
type Plugin interface {
    Name() string
    Initialize(engine *Engine) error
    Update(dt float64) error
    Shutdown() error
}

type PluginManager struct {
    plugins []Plugin
}

func (pm *PluginManager) LoadPlugin(plugin Plugin) error {
    if err := plugin.Initialize(pm.engine); err != nil {
        return err
    }
    pm.plugins = append(pm.plugins, plugin)
    return nil
}
```

### 2. イベントシステム

```go
type Event struct {
    Type string
    Data interface{}
}

type EventSystem struct {
    handlers map[string][]func(Event)
}

func (es *EventSystem) On(eventType string, handler func(Event)) {
    es.handlers[eventType] = append(es.handlers[eventType], handler)
}

func (es *EventSystem) Emit(event Event) {
    for _, handler := range es.handlers[event.Type] {
        handler(event)
    }
}
```

## 今後の拡張予定

### 1. 高度な機能
- **AI システム**: より複雑なモンスターAI
- **アイテム生成**: 手続き的なアイテム生成
- **クエストシステム**: 動的なクエスト生成
- **マルチプレイヤー**: ネットワーク対応

### 2. 技術的改善
- **パフォーマンス最適化**: プロファイリングベースの最適化
- **メモリ効率**: より効率的なメモリ使用
- **並行処理**: ゲームロジックの並列化
- **キャッシュシステム**: 効率的なデータキャッシュ

## まとめ

GoRogueのアーキテクチャは、Go言語の特徴を活かしつつ、以下の要素を統合することで、高品質なゲーム体験と継続的な開発を可能にしています：

1. **シンプルな設計**: Goらしいシンプルで理解しやすい構造
2. **実証済みの設計パターン**: MVC、State、Component、Observerパターンの適切な活用
3. **Gruidライブラリの効果的な活用**: ローグライクゲームに特化したライブラリの特性を活かした設計
4. **テスト可能な設計**: 単体テスト、統合テスト、ベンチマークの包括的な実装
5. **型安全性**: Goの静的型付けによる安全な開発
6. **性能最適化**: メモリ効率、描画最適化、並行処理の活用

この設計により、GoRogueは単なるゲームプロジェクトではなく、Go言語を使用したゲーム開発のベストプラクティスを示す包括的な教材としても機能することを目指しています。