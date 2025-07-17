package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	MinRoomSize = 4
	MaxRoomSize = 10
	MaxRooms    = 30
	MinRooms    = 15
)

// Room represents a room in the dungeon
type Room struct {
	X, Y          int
	Width, Height int
	IsSpecial     bool
	Connected     bool
}

// Level represents a single dungeon level
type Level struct {
	Width, Height int
	Tiles         [][]*Tile
	Rooms         []*Room
	FloorNumber   int
	Monsters      []*actor.Monster
	Items         []*item.Item
}

// NewLevel creates a new dungeon level using the builder pattern
func NewLevel(width, height, floorNum int) *Level {
	var level *Level

	// 特別な階層（迷路階層）のチェック
	if floorNum == 7 || floorNum == 13 || floorNum == 19 {
		// 迷路階層を生成
		mazeBuilder := NewMazeBuilder(width, height, floorNum)
		level = mazeBuilder.Build()
		logger.Info("Created maze level",
			"width", width,
			"height", height,
			"floor", floorNum,
			"type", "maze",
		)
	} else {
		// 通常の階層を生成
		builder := NewDungeonBuilder(width, height, floorNum)
		level = builder.Build()
		logger.Debug("Created normal level",
			"width", width,
			"height", height,
			"floor", floorNum,
			"type", "normal",
		)
	}

	return level
}

// Generate generates the dungeon layout
func (l *Level) Generate() {
	// 部屋の生成
	numRooms := MinRooms + rand.Intn(MaxRooms-MinRooms+1)
	for i := 0; i < numRooms; i++ {
		l.GenerateRoom()
	}

	// 部屋の接続
	l.ConnectRooms()

	// 特別な部屋の生成
	l.GenerateSpecialRoom()

	// 階段の配置
	l.PlaceStairs()

	// モンスターの配置
	l.SpawnMonsters()

	// アイテムの配置
	l.SpawnItems()

	logger.Info("Generated dungeon level",
		"floor", l.FloorNumber,
		"rooms", len(l.Rooms),
		"monsters", len(l.Monsters),
		"items", len(l.Items),
	)
}

// GenerateRoom generates a single room
func (l *Level) GenerateRoom() {
	for attempts := 0; attempts < 100; attempts++ {
		width := MinRoomSize + rand.Intn(MaxRoomSize-MinRoomSize+1)
		height := MinRoomSize + rand.Intn(MaxRoomSize-MinRoomSize+1)
		x := 1 + rand.Intn(l.Width-width-2)
		y := 1 + rand.Intn(l.Height-height-2)

		if l.CanPlaceRoom(x, y, width, height) {
			room := &Room{
				X:      x,
				Y:      y,
				Width:  width,
				Height: height,
			}
			l.AddRoom(room)
			return
		}
	}
}

// CanPlaceRoom checks if a room can be placed at the given position
func (l *Level) CanPlaceRoom(x, y, width, height int) bool {
	// 部屋の周囲1マスも含めてチェック
	for dy := -1; dy <= height; dy++ {
		for dx := -1; dx <= width; dx++ {
			nx, ny := x+dx, y+dy
			if !l.IsInBounds(nx, ny) {
				return false
			}
			if l.GetTile(nx, ny).Type != TileWall {
				return false
			}
		}
	}
	return true
}

// ConnectRooms connects all rooms with corridors
func (l *Level) ConnectRooms() {
	if len(l.Rooms) < 2 {
		return
	}

	// 最初の部屋を接続済みとしてマーク
	l.Rooms[0].Connected = true

	// 残りの部屋を接続
	for i := 1; i < len(l.Rooms); i++ {
		room := l.Rooms[i]
		// 最も近い接続済みの部屋を探す
		closestRoom := l.FindClosestConnectedRoom(room)
		if closestRoom != nil {
			l.ConnectRoomPair(room, closestRoom)
			room.Connected = true
		}
	}
}

// FindClosestConnectedRoom finds the closest connected room
func (l *Level) FindClosestConnectedRoom(room *Room) *Room {
	var closest *Room
	minDist := l.Width * l.Height

	for _, other := range l.Rooms {
		if other == room || !other.Connected {
			continue
		}

		dist := (room.X-other.X)*(room.X-other.X) + (room.Y-other.Y)*(room.Y-other.Y)
		if dist < minDist {
			minDist = dist
			closest = other
		}
	}

	return closest
}

// ConnectRoomPair connects two rooms with a corridor
func (l *Level) ConnectRoomPair(r1, r2 *Room) {
	// 部屋の中心点を計算
	x1 := r1.X + r1.Width/2
	y1 := r1.Y + r1.Height/2
	x2 := r2.X + r2.Width/2
	y2 := r2.Y + r2.Height/2

	// L字型の通路を生成
	if rand.Float64() < 0.5 {
		l.CreateHorizontalCorridor(x1, x2, y1)
		l.CreateVerticalCorridor(y1, y2, x2)
	} else {
		l.CreateVerticalCorridor(y1, y2, x1)
		l.CreateHorizontalCorridor(x1, x2, y2)
	}
}

// CreateHorizontalCorridor creates a horizontal corridor
func (l *Level) CreateHorizontalCorridor(x1, x2, y int) {
	for x := min(x1, x2); x <= max(x1, x2); x++ {
		if l.GetTile(x, y).Type == TileWall {
			l.SetTile(x, y, TileFloor)
		}
	}
}

// CreateVerticalCorridor creates a vertical corridor
func (l *Level) CreateVerticalCorridor(y1, y2, x int) {
	for y := min(y1, y2); y <= max(y1, y2); y++ {
		if l.GetTile(x, y).Type == TileWall {
			l.SetTile(x, y, TileFloor)
		}
	}
}

// PlaceStairs places the stairs in the dungeon
func (l *Level) PlaceStairs() {
	if len(l.Rooms) == 0 {
		return
	}

	// 最初の階層では上り階段を配置しない
	if l.FloorNumber > 1 {
		// 上り階段は最初の部屋に配置
		firstRoom := l.Rooms[0]
		l.SetTile(
			firstRoom.X+firstRoom.Width/2,
			firstRoom.Y+firstRoom.Height/2,
			TileStairsUp,
		)
	}

	// 最終階層では下り階段を配置しない
	if l.FloorNumber < 26 {
		// 下り階段は最後の部屋に配置
		lastRoom := l.Rooms[len(l.Rooms)-1]
		l.SetTile(
			lastRoom.X+lastRoom.Width/2,
			lastRoom.Y+lastRoom.Height/2,
			TileStairsDown,
		)
	}
}

// IsInBounds checks if the given coordinates are within the level bounds
func (l *Level) IsInBounds(x, y int) bool {
	return x >= 0 && x < l.Width && y >= 0 && y < l.Height
}

// GetTile returns the tile at the given coordinates
func (l *Level) GetTile(x, y int) *Tile {
	if !l.IsInBounds(x, y) {
		return nil
	}
	return l.Tiles[y][x]
}

// IsWalkable checks if a position is walkable
func (l *Level) IsWalkable(x, y int) bool {
	tile := l.GetTile(x, y)
	return tile != nil && tile.Walkable()
}

// SetTile sets the tile at the given coordinates
func (l *Level) SetTile(x, y int, tileType TileType) {
	if l.IsInBounds(x, y) {
		l.Tiles[y][x] = NewTile(tileType)
		logger.Debug("Set tile",
			"x", x,
			"y", y,
			"tile_type", tileType,
		)
	}
}

// AddRoom adds a room to the level
func (l *Level) AddRoom(room *Room) {
	// Fill room with floor tiles
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			if l.IsInBounds(x, y) {
				l.SetTile(x, y, TileFloor)
			}
		}
	}

	// Add room to the list
	l.Rooms = append(l.Rooms, room)

	logger.Debug("Added room",
		"x", room.X,
		"y", room.Y,
		"width", room.Width,
		"height", room.Height,
		"is_special", room.IsSpecial,
		"total_rooms", len(l.Rooms),
	)
}

// IsSpecialFloor returns whether this floor should have a special room
func (l *Level) IsSpecialFloor() bool {
	return l.FloorNumber%5 == 0
}

// ShouldGenerateSpecialRoom returns whether a special room should be generated
func (l *Level) ShouldGenerateSpecialRoom() bool {
	shouldGenerate := l.IsSpecialFloor() && rand.Float64() < 0.10 // 10% chance
	if shouldGenerate {
		logger.Info("Special room generation triggered",
			"floor", l.FloorNumber,
		)
	}
	return shouldGenerate
}

// Helper functions
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// SpawnMonsters spawns monsters in the dungeon
func (l *Level) SpawnMonsters() {
	if len(l.Rooms) == 0 {
		return
	}

	// 階層に応じたモンスター数を計算（DungeonManagerの計算を使用）
	numMonsters := l.getMonsterSpawnCount()

	// 各部屋にモンスターを配置
	for i := 0; i < numMonsters; i++ {
		// ランダムな部屋を選択
		room := l.Rooms[rand.Intn(len(l.Rooms))]

		// 部屋内のランダムな位置を選択
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)

		// その位置が床タイルかチェック
		if l.GetTile(x, y).Type != TileFloor {
			i-- // 無効な位置の場合は再試行
			continue
		}

		// 既にモンスターがいないかチェック
		if l.GetMonsterAt(x, y) != nil {
			i-- // 既にモンスターがいる場合は再試行
			continue
		}

		// 階層に応じたモンスターを選択
		monsterType := l.selectMonsterType()
		monster := actor.NewMonster(x, y, monsterType)

		// 階層に応じた難易度スケーリング
		l.scaleMonsterForFloor(monster)

		l.Monsters = append(l.Monsters, monster)

		logger.Debug("Spawned monster",
			"type", monster.Type.Name,
			"x", x,
			"y", y,
			"floor", l.FloorNumber,
		)
	}

	logger.Info("Finished spawning monsters",
		"total_monsters", len(l.Monsters),
		"floor", l.FloorNumber,
	)
}

// selectMonsterType selects a monster type based on the floor level
// Following original Rogue's monster distribution system
func (l *Level) selectMonsterType() rune {
	// 階層に応じたモンスター選択 (A-Z全26種類対応)
	// より詳細な階層分布を実装
	switch {
	case l.FloorNumber <= 2:
		// 最浅階層：超弱いモンスター
		monsters := []rune{'A', 'B', 'F', 'G', 'K'} // Aquator, Bat, Flyting, Griffin, Kobold
		return monsters[rand.Intn(len(monsters))]
	case l.FloorNumber <= 5:
		// 浅い階層：弱いモンスター
		monsters := []rune{'A', 'B', 'E', 'F', 'G', 'I', 'K', 'N'} // + Emu, Ice monster, Nymph
		return monsters[rand.Intn(len(monsters))]
	case l.FloorNumber <= 8:
		// 初期中間階層：基本的なモンスター
		monsters := []rune{'A', 'B', 'E', 'F', 'G', 'I', 'K', 'L', 'N', 'R', 'S'} // + Leprechaun, Rattlesnake, Snake
		return monsters[rand.Intn(len(monsters))]
	case l.FloorNumber <= 12:
		// 中間階層：中程度のモンスター
		monsters := []rune{'B', 'C', 'E', 'G', 'H', 'I', 'J', 'L', 'O', 'R', 'S', 'W'} // + Centaur, Hobgoblin, Jackal, Orc, Wraith
		return monsters[rand.Intn(len(monsters))]
	case l.FloorNumber <= 16:
		// 深い階層：強いモンスター
		monsters := []rune{'C', 'E', 'G', 'H', 'J', 'M', 'O', 'P', 'S', 'T', 'U', 'W', 'Z'} // + Minotaur, Phantom, Troll, Ur-vile, Zombie
		return monsters[rand.Intn(len(monsters))]
	case l.FloorNumber <= 20:
		// 深層：非常に強いモンスター
		monsters := []rune{'C', 'H', 'M', 'O', 'P', 'Q', 'T', 'U', 'V', 'W', 'X', 'Y', 'Z'} // + Quasit, Vampire, Xorn, Yeti
		return monsters[rand.Intn(len(monsters))]
	case l.FloorNumber <= 24:
		// 最深層：最強のモンスター
		monsters := []rune{'D', 'M', 'P', 'Q', 'T', 'U', 'V', 'X', 'Y', 'Z'} // + Dragon
		return monsters[rand.Intn(len(monsters))]
	default:
		// 最終階層：ドラゴンと最強モンスター
		monsters := []rune{'D', 'Q', 'T', 'V', 'X', 'Y', 'Z'} // 最強のみ
		return monsters[rand.Intn(len(monsters))]
	}
}

// getMonsterSpawnCount returns the number of monsters to spawn on this floor
func (l *Level) getMonsterSpawnCount() int {
	switch {
	case l.FloorNumber <= 3:
		return 3 + (l.FloorNumber - 1) // 3-5体
	case l.FloorNumber <= 8:
		return 5 + (l.FloorNumber - 4) // 5-9体
	case l.FloorNumber <= 15:
		return 8 + (l.FloorNumber-9)/2 // 8-11体
	case l.FloorNumber <= 22:
		return 12 + (l.FloorNumber-16)/3 // 12-14体
	default:
		return 15 + (l.FloorNumber - 23) // 15-18体
	}
}

// getDifficultyScaling returns the difficulty scaling for this floor
func (l *Level) getDifficultyScaling() float64 {
	switch {
	case l.FloorNumber <= 5:
		return 1.0 + (float64(l.FloorNumber-1) * 0.1) // 1.0 - 1.4
	case l.FloorNumber <= 10:
		return 1.5 + (float64(l.FloorNumber-6) * 0.1) // 1.5 - 1.9
	case l.FloorNumber <= 15:
		return 2.0 + (float64(l.FloorNumber-11) * 0.1) // 2.0 - 2.4
	case l.FloorNumber <= 20:
		return 2.5 + (float64(l.FloorNumber-16) * 0.1) // 2.5 - 2.9
	case l.FloorNumber <= 26:
		return 3.0 + (float64(l.FloorNumber-21) * 0.2) // 3.0 - 4.0
	default:
		return 4.0 // 最大難易度
	}
}

// scaleMonsterForFloor scales monster stats based on the floor level
func (l *Level) scaleMonsterForFloor(monster *actor.Monster) {
	// 新しい階層スケーリング係数を使用
	scaleFactor := l.getDifficultyScaling()

	// HP, 攻撃力, 防御力を階層に応じて強化
	monster.MaxHP = int(float64(monster.MaxHP) * scaleFactor)
	monster.HP = monster.MaxHP
	monster.Attack = int(float64(monster.Attack) * scaleFactor)
	monster.Defense = int(float64(monster.Defense) * scaleFactor)

	// 深い階層では追加のボーナス（21階以上）
	if l.FloorNumber > 20 {
		extraBonus := float64(l.FloorNumber-20) * 0.1
		monster.MaxHP = int(float64(monster.MaxHP) * (1.0 + extraBonus))
		monster.HP = monster.MaxHP
		monster.Attack = int(float64(monster.Attack) * (1.0 + extraBonus))
		monster.Defense = int(float64(monster.Defense) * (1.0 + extraBonus))
	}

	logger.Debug("Scaled monster for floor",
		"floor", l.FloorNumber,
		"type", monster.Type.Name,
		"hp", monster.HP,
		"attack", monster.Attack,
		"defense", monster.Defense,
		"scale_factor", scaleFactor,
	)
}

// GetMonsterAt returns the monster at the given coordinates
func (l *Level) GetMonsterAt(x, y int) *actor.Monster {
	for _, monster := range l.Monsters {
		if monster.Position.X == x && monster.Position.Y == y && monster.IsAlive() {
			return monster
		}
	}
	return nil
}

// RemoveMonster removes a monster from the level
func (l *Level) RemoveMonster(monster *actor.Monster) {
	for i, m := range l.Monsters {
		if m == monster {
			l.Monsters = append(l.Monsters[:i], l.Monsters[i+1:]...)
			logger.Debug("Removed monster",
				"type", monster.Type.Name,
				"x", monster.Position.X,
				"y", monster.Position.Y,
			)
			break
		}
	}
}

// UpdateMonsters updates all monsters in the level
func (l *Level) UpdateMonsters(player *actor.Player) {
	for _, monster := range l.Monsters {
		if monster.IsAlive() {
			monster.Update(player, l)
		}
	}

	// 死んだモンスターを削除
	l.RemoveDeadMonsters()
}

// RemoveDeadMonsters removes all dead monsters from the level
func (l *Level) RemoveDeadMonsters() {
	aliveMonsters := make([]*actor.Monster, 0)
	for _, monster := range l.Monsters {
		if monster.IsAlive() {
			aliveMonsters = append(aliveMonsters, monster)
		}
	}
	l.Monsters = aliveMonsters
}

// GenerateSpecialRoom generates a special room
func (l *Level) GenerateSpecialRoom() {
	// 1階では特別な部屋を生成しない
	if l.FloorNumber <= 1 {
		return
	}

	// 5階ごとに1つの特別な部屋を生成
	if l.FloorNumber%5 != 0 {
		return
	}

	// 10%の確率で特別な部屋を生成
	if rand.Float64() > 0.1 {
		return
	}

	// 既に特別な部屋が存在する場合は生成しない
	for _, room := range l.Rooms {
		if room.IsSpecial {
			return
		}
	}

	// 5x5の特別な部屋を生成
	for attempts := 0; attempts < 100; attempts++ {
		x := 1 + rand.Intn(l.Width-7)  // 5x5の部屋 + 周囲1マス
		y := 1 + rand.Intn(l.Height-7) // 5x5の部屋 + 周囲1マス

		if l.CanPlaceRoom(x, y, 5, 5) {
			room := &Room{
				X:         x,
				Y:         y,
				Width:     5,
				Height:    5,
				IsSpecial: true,
			}
			l.AddRoom(room)

			// 隠し扉を配置
			l.PlaceSecretDoor(room)

			// 部屋の内容を生成
			l.PopulateSpecialRoom(room)

			logger.Info("Generated special room",
				"floor", l.FloorNumber,
				"x", x,
				"y", y,
			)
			return
		}
	}
}

// PlaceSecretDoor places a secret door for a special room
func (l *Level) PlaceSecretDoor(room *Room) {
	// 部屋の4辺のいずれかにランダムに隠し扉を配置
	side := rand.Intn(4)
	var x, y int

	switch side {
	case 0: // 上辺
		x = room.X + rand.Intn(room.Width)
		y = room.Y - 1
	case 1: // 右辺
		x = room.X + room.Width
		y = room.Y + rand.Intn(room.Height)
	case 2: // 下辺
		x = room.X + rand.Intn(room.Width)
		y = room.Y + room.Height
	case 3: // 左辺
		x = room.X - 1
		y = room.Y + rand.Intn(room.Height)
	}

	if l.IsInBounds(x, y) {
		l.SetTile(x, y, TileSecretDoor)
		logger.Debug("Placed secret door",
			"x", x,
			"y", y,
		)
	}
}

// PopulateSpecialRoom populates a special room with content
func (l *Level) PopulateSpecialRoom(room *Room) {
	// 部屋の種類をランダムに決定
	roomType := rand.Intn(6)

	switch roomType {
	case 0: // 宝物庫
		logger.Info("Generating treasure vault")
		// TODO: 宝物を配置
	case 1: // 武器庫
		logger.Info("Generating armory")
		// TODO: 武器を配置
	case 2: // 食料庫
		logger.Info("Generating food storage")
		// TODO: 食料を配置
	case 3: // 魔物のねぐら
		logger.Info("Generating monster lair")
		// TODO: モンスターを配置
	case 4: // 実験室
		logger.Info("Generating laboratory")
		// TODO: 薬を配置
	case 5: // 図書室
		logger.Info("Generating library")
		// TODO: 巻物を配置
	}
}

// SpawnItems spawns items in the level
func (l *Level) SpawnItems() {
	logger.Debug("Starting item spawning", "floor", l.FloorNumber)

	// 階層に応じたアイテムスポーン確率を取得
	itemSpawnChance := l.getItemSpawnChance()

	// 各部屋にアイテムを配置
	for _, room := range l.Rooms {
		// 通常の部屋: 階層に応じた確率でアイテムを配置
		if rand.Float64() < itemSpawnChance {
			l.spawnItemInRoom(room)
		}

		// 特別な部屋: 必ずアイテムを配置
		if room.IsSpecial {
			l.spawnItemInRoom(room)
		}
	}

	logger.Info("Finished spawning items",
		"total_items", len(l.Items),
		"floor", l.FloorNumber,
	)
}

// spawnItemInRoom spawns an item in a specific room
func (l *Level) spawnItemInRoom(room *Room) {
	maxAttempts := 20
	for attempts := 0; attempts < maxAttempts; attempts++ {
		// 部屋内のランダムな位置を選択
		x := room.X + rand.Intn(room.Width)
		y := room.Y + rand.Intn(room.Height)

		// その位置が有効かチェック
		if !l.IsValidItemPosition(x, y) {
			continue
		}

		// アイテムタイプを選択
		itemType := l.selectItemType()

		// アイテムを生成
		var newItem *item.Item
		switch itemType {
		case item.ItemGold:
			newItem = item.NewGold(x, y, room.IsSpecial)
			// 階層に応じてゴールドの価値を調整
			if newItem != nil {
				newItem.Value = int(float64(newItem.Value) * (1.0 + float64(l.FloorNumber-1)*0.1))
			}
		case item.ItemAmulet:
			if l.FloorNumber >= 20 { // 深い階層でのみ魔除けを生成
				newItem = item.NewAmulet(x, y)
			} else {
				continue // 深い階層でない場合は別のアイテムを試す
			}
		default:
			newItem = l.createRandomItem(x, y, itemType)
			// 階層に応じてアイテムの価値を調整
			if newItem != nil {
				newItem.Value = int(float64(newItem.Value) * (1.0 + float64(l.FloorNumber-1)*0.05))
			}
		}

		if newItem != nil {
			l.Items = append(l.Items, newItem)
			logger.Debug("Spawned item",
				"type", newItem.Name,
				"x", x,
				"y", y,
				"floor", l.FloorNumber,
			)
		}
		break
	}
}

// getItemSpawnChance returns the item spawn chance for this floor
func (l *Level) getItemSpawnChance() float64 {
	switch {
	case l.FloorNumber <= 5:
		return 0.2 + (float64(l.FloorNumber-1) * 0.02) // 20%-28%
	case l.FloorNumber <= 10:
		return 0.3 + (float64(l.FloorNumber-6) * 0.02) // 30%-38%
	case l.FloorNumber <= 15:
		return 0.4 + (float64(l.FloorNumber-11) * 0.02) // 40%-48%
	case l.FloorNumber <= 20:
		return 0.5 + (float64(l.FloorNumber-16) * 0.02) // 50%-58%
	default:
		return 0.6 + (float64(l.FloorNumber-21) * 0.02) // 60%-70%
	}
}

// selectItemType selects an item type based on the floor level
func (l *Level) selectItemType() item.ItemType {
	// 階層に応じたアイテム選択（より詳細な分布）
	switch {
	case l.FloorNumber <= 3:
		// 最浅階層: 基本的なアイテムのみ
		items := []item.ItemType{item.ItemGold, item.ItemFood, item.ItemPotion}
		weights := []float64{0.5, 0.3, 0.2} // ゴールド50%、食料30%、薬20%
		return l.selectWeightedItem(items, weights)
	case l.FloorNumber <= 7:
		// 浅い階層: 基本的なアイテム
		items := []item.ItemType{item.ItemGold, item.ItemFood, item.ItemPotion, item.ItemScroll}
		weights := []float64{0.4, 0.25, 0.2, 0.15} // ゴールド40%、食料25%、薬20%、巻物15%
		return l.selectWeightedItem(items, weights)
	case l.FloorNumber <= 12:
		// 中間階層: より多様なアイテム
		items := []item.ItemType{item.ItemGold, item.ItemFood, item.ItemPotion, item.ItemScroll, item.ItemWeapon}
		weights := []float64{0.3, 0.2, 0.2, 0.15, 0.15} // ゴールド30%、食料20%、薬20%、巻物15%、武器15%
		return l.selectWeightedItem(items, weights)
	case l.FloorNumber <= 18:
		// 深い階層: 高価なアイテム
		items := []item.ItemType{item.ItemGold, item.ItemWeapon, item.ItemArmor, item.ItemRing, item.ItemScroll, item.ItemPotion}
		weights := []float64{0.25, 0.2, 0.2, 0.15, 0.1, 0.1} // ゴールド25%、武器20%、鎧20%、指輪15%、巻物10%、薬10%
		return l.selectWeightedItem(items, weights)
	case l.FloorNumber <= 25:
		// 最深階層: 最高のアイテム
		items := []item.ItemType{item.ItemGold, item.ItemWeapon, item.ItemArmor, item.ItemRing, item.ItemScroll}
		weights := []float64{0.2, 0.25, 0.25, 0.2, 0.1} // ゴールド20%、武器25%、鎧25%、指輪20%、巻物10%
		return l.selectWeightedItem(items, weights)
	default:
		// 最終階層: 最高のアイテム + 魔除け
		items := []item.ItemType{item.ItemGold, item.ItemWeapon, item.ItemArmor, item.ItemRing, item.ItemAmulet}
		weights := []float64{0.15, 0.25, 0.25, 0.25, 0.1} // ゴールド15%、武器25%、鎧25%、指輪25%、魔除け10%
		return l.selectWeightedItem(items, weights)
	}
}

// selectWeightedItem selects an item based on weighted probabilities
func (l *Level) selectWeightedItem(items []item.ItemType, weights []float64) item.ItemType {
	if len(items) != len(weights) {
		// フォールバック: 最初のアイテムを返す
		return items[0]
	}

	totalWeight := 0.0
	for _, weight := range weights {
		totalWeight += weight
	}

	r := rand.Float64() * totalWeight
	currentWeight := 0.0

	for i, weight := range weights {
		currentWeight += weight
		if r <= currentWeight {
			return items[i]
		}
	}

	// フォールバック: 最後のアイテムを返す
	return items[len(items)-1]
}

// createRandomItem creates a random item of the specified type
func (l *Level) createRandomItem(x, y int, itemType item.ItemType) *item.Item {
	switch itemType {
	case item.ItemWeapon:
		weapons := []string{"短剣", "剣", "メイス", "斧", "弓"}
		name := weapons[rand.Intn(len(weapons))]
		value := 10 + rand.Intn(50)
		return item.NewItem(x, y, itemType, name, value)
	case item.ItemArmor:
		armors := []string{"革鎧", "鎖帷子", "板金鎧", "ローブ", "盾"}
		name := armors[rand.Intn(len(armors))]
		value := 20 + rand.Intn(80)
		return item.NewItem(x, y, itemType, name, value)
	case item.ItemRing:
		rings := []string{"力の指輪", "知恵の指輪", "体力の指輪", "敏捷の指輪"}
		name := rings[rand.Intn(len(rings))]
		value := 50 + rand.Intn(100)
		return item.NewItem(x, y, itemType, name, value)
	case item.ItemScroll:
		scrolls := []string{"テレポートの巻物", "識別の巻物", "治療の巻物", "魔法の巻物"}
		name := scrolls[rand.Intn(len(scrolls))]
		value := 15 + rand.Intn(35)
		return item.NewItem(x, y, itemType, name, value)
	case item.ItemPotion:
		potions := []string{"体力回復薬", "魔力回復薬", "力強化薬", "敏捷強化薬"}
		name := potions[rand.Intn(len(potions))]
		value := 10 + rand.Intn(30)
		return item.NewItem(x, y, itemType, name, value)
	case item.ItemFood:
		foods := []string{"パン", "肉", "果物", "チーズ", "干し肉"}
		name := foods[rand.Intn(len(foods))]
		value := 5 + rand.Intn(15)
		return item.NewItem(x, y, itemType, name, value)
	default:
		return nil
	}
}

// IsValidItemPosition checks if an item can be placed at the given position
func (l *Level) IsValidItemPosition(x, y int) bool {
	// 境界チェック
	if !l.IsInBounds(x, y) {
		return false
	}

	// 歩行可能タイルかチェック
	tile := l.GetTile(x, y)
	if tile == nil || !tile.Walkable() {
		return false
	}

	// 既にアイテムがある位置かチェック
	for _, existingItem := range l.Items {
		if existingItem.Position.X == x && existingItem.Position.Y == y {
			return false
		}
	}

	return true
}

// GetItemAt returns the item at the given coordinates
func (l *Level) GetItemAt(x, y int) *item.Item {
	for _, item := range l.Items {
		if item.Position.X == x && item.Position.Y == y {
			return item
		}
	}
	return nil
}

// RemoveItem removes an item from the level
func (l *Level) RemoveItem(item *item.Item) {
	for i, it := range l.Items {
		if it == item {
			l.Items = append(l.Items[:i], l.Items[i+1:]...)
			logger.Debug("Removed item",
				"type", item.Name,
				"x", item.Position.X,
				"y", item.Position.Y,
			)
			break
		}
	}
}

// AddItem アイテムを指定位置に追加
func (l *Level) AddItem(item *item.Item, x, y int) {
	item.Position.X = x
	item.Position.Y = y
	l.Items = append(l.Items, item)
	logger.Debug("Item added to level",
		"type", item.Name,
		"x", x,
		"y", y,
	)
}
