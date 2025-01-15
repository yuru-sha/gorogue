package dungeon

import (
	"math/rand"

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
			// 外周は壁、それ以外は床
			if x == 0 || x == width-1 || y == 0 || y == height-1 {
				level.Tiles[y][x] = NewTile(TileWall)
			} else {
				level.Tiles[y][x] = NewTile(TileFloor)
			}
		}
	}

	logger.Debug("Created level",
		"width", width,
		"height", height,
		"floor", floorNum,
		"total_tiles", width*height,
	)
	return level
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
	l.Rooms = append(l.Rooms, room)
	logger.Debug("Added room",
		"x", room.X,
		"y", room.Y,
		"width", room.Width,
		"height", room.Height,
		"is_special", room.IsSpecial,
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

// GenerateSpecialRoom generates a special room if conditions are met
func (l *Level) GenerateSpecialRoom() {
	if !l.ShouldGenerateSpecialRoom() {
		return
	}
	// TODO: Implement special room generation
	logger.Debug("Special room generation requested",
		"floor", l.FloorNumber,
	)
}
