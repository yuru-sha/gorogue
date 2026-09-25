package dungeon

import (
	"math/rand"
	"time"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

const (
	MinRoomSize = 4
	MaxRoomSize = 10
)

// Room represents a room in the dungeon
type Room struct {
	X, Y          int
	Width, Height int
	IsSpecial     bool
	IsDark        bool
	IsMaze        bool
	Connected     bool
}

// Level represents a single dungeon level
type Level struct {
	Width, Height int
	Tiles         [][]*Tile
	Rooms         []*Room
	FloorNumber   int
	Seed          int64
	NoFood        int
	Monsters      []*actor.Monster
	Items         []*item.Item
	Traps         []*Trap
	rng           *rand.Rand
	rngSource     *trackedRandomSource
}

// NewLevel creates a new dungeon level using the builder pattern
func NewLevel(width, height, floorNum int) *Level {
	return newLevelWithRandomSource(width, height, floorNum, newTrackedRandomSource(time.Now().UnixNano()))
}

// NewLevelWithSeed creates a reproducible dungeon level.
func NewLevelWithSeed(width, height, floorNum int, seed int64) *Level {
	return NewLevelWithSeedAndNoFood(width, height, floorNum, seed, 0)
}

// NewLevelWithSeedAndNoFood resumes Rogue's cross-level food generation state.
func NewLevelWithSeedAndNoFood(width, height, floorNum int, seed int64, noFood int) *Level {
	level := newLevelWithRandomSourceAndNoFood(width, height, floorNum, newTrackedRandomSource(seed), noFood)
	level.Seed = seed
	return level
}

func newRandom() *rand.Rand {
	return newTrackedRandomSource(time.Now().UnixNano()).rand()
}

func newLevelWithRandomSource(width, height, floorNum int, source *trackedRandomSource) *Level {
	return newLevelWithRandomSourceAndNoFood(width, height, floorNum, source, 0)
}

func newLevelWithRandomSourceAndNoFood(width, height, floorNum int, source *trackedRandomSource, noFood int) *Level {
	builder := NewDungeonBuilderWithRand(width, height, floorNum, source.rand())
	builder.level.NoFood = noFood
	level := builder.Build()
	logger.Debug("Created level",
		"width", width,
		"height", height,
		"floor", floorNum,
		"rooms", len(level.Rooms),
	)

	level.rngSource = source
	return level
}

func (l *Level) random() *rand.Rand {
	if l.rng == nil {
		l.rngSource = newTrackedRandomSource(time.Now().UnixNano())
		l.rng = l.rngSource.rand()
	}
	return l.rng
}

// RandomDraws returns the number of values consumed by the level random source.
func (l *Level) RandomDraws() uint64 {
	if l.rngSource == nil {
		return 0
	}
	return l.rngSource.draws
}

// SetRandomDraws restores a level's deterministic random cursor.
func (l *Level) SetRandomDraws(draws uint64) error {
	source, err := newTrackedRandomSourceAt(l.Seed, draws)
	if err != nil {
		return err
	}
	l.rngSource = source
	l.rng = source.rand()
	return nil
}

// Generate runs the source level-generation flow on an initialized level.
func (l *Level) Generate() {
	slots := l.generateRogueRooms()
	l.connectRogueRooms(&slots)
	l.NoFood++
	l.SpawnItems()
	l.placeRogueTraps()
	NewStairsManager(l).PlaceStairs()
}

// GenerateRoom generates a single room
func (l *Level) GenerateRoom() {
	for range 100 {
		width := MinRoomSize + l.random().Intn(MaxRoomSize-MinRoomSize+1)
		height := MinRoomSize + l.random().Intn(MaxRoomSize-MinRoomSize+1)
		x := 1 + l.random().Intn(l.Width-width-2)
		y := 1 + l.random().Intn(l.Height-height-2)

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
		tile := NewTile(tileType)
		if previous := l.Tiles[y][x]; previous != nil {
			tile.Visible = previous.Visible
			tile.Explored = previous.Explored
		}
		l.Tiles[y][x] = tile
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

// Helper functions

func maxInt(a, b int) int {
	if a > b {
		return a
	}
	return b
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
func (l *Level) RemoveItem(gameItem *item.Item) {
	for i, it := range l.Items {
		if it == gameItem {
			l.Items = append(l.Items[:i], l.Items[i+1:]...)
			logger.Debug("Removed item",
				"type", gameItem.Name,
				"x", gameItem.Position.X,
				"y", gameItem.Position.Y,
			)
			break
		}
	}
}

// AddItem アイテムを指定位置に追加
func (l *Level) AddItem(gameItem *item.Item, x, y int) {
	gameItem.Position.X = x
	gameItem.Position.Y = y
	l.Items = append(l.Items, gameItem)
	logger.Debug("Item added to level",
		"type", gameItem.Name,
		"x", x,
		"y", y,
	)
}
