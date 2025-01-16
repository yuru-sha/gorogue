package dungeon

import (
	"math/rand"

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
}

// NewLevel creates a new dungeon level
func NewLevel(width, height, floorNum int) *Level {
	level := &Level{
		Width:       width,
		Height:      height,
		FloorNumber: floorNum,
		Rooms:       make([]*Room, 0),
	}

	// Initialize tiles with walls
	level.Tiles = make([][]*Tile, height)
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	// ダンジョンの生成
	level.Generate()

	logger.Debug("Created level",
		"width", width,
		"height", height,
		"floor", floorNum,
		"total_tiles", width*height,
	)
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

	logger.Info("Generated dungeon level",
		"floor", l.FloorNumber,
		"rooms", len(l.Rooms),
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
	// 下り階段は最後の部屋に配置
	lastRoom := l.Rooms[len(l.Rooms)-1]
	l.SetTile(
		lastRoom.X+lastRoom.Width/2,
		lastRoom.Y+lastRoom.Height/2,
		TileStairsDown,
	)

	// 上り階段は最初の部屋に配置
	firstRoom := l.Rooms[0]
	l.SetTile(
		firstRoom.X+firstRoom.Width/2,
		firstRoom.Y+firstRoom.Height/2,
		TileStairsUp,
	)
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

// GenerateSpecialRoom generates a special room
func (l *Level) GenerateSpecialRoom() {
	// 5階ごとに1つの特別な部屋を生成
	if l.FloorNumber%5 != 0 {
		return
	}

	// 10%の確率で特別な部屋を生成
	if rand.Float64() > 0.1 {
		return
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
