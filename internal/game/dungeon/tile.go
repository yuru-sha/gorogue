package dungeon

import "github.com/anaseto/gruid"

// TileType represents different types of tiles in the dungeon
type TileType int

const (
	TileWall TileType = iota
	TileFloor
	TileDoor
	TileDoorClosed
	TileDoorOpen
	TileOpenDoor
	TileStairsUp
	TileStairsDown
	TileWater
	TileLava
	TileSecretDoor
)

// String returns the string representation of a TileType
func (t TileType) String() string {
	switch t {
	case TileFloor:
		return "floor"
	case TileWall:
		return "wall"
	case TileWater:
		return "water"
	case TileLava:
		return "lava"
	default:
		return "unknown"
	}
}

// Tile represents a single tile in the dungeon
type Tile struct {
	Type       TileType
	Rune       rune
	Color      gruid.Color
	Visible    bool
	IsWalkable bool
}

// Walkable returns whether the tile can be walked on
func (t *Tile) Walkable() bool {
	return t.IsWalkable
}

// NewTile creates a new tile of the given type
func NewTile(tileType TileType) *Tile {
	t := &Tile{
		Type:       tileType,
		Visible:    true, // すべてのタイルを可視化（簡素化のため）
		IsWalkable: IsWalkable(tileType),
	}
	switch tileType {
	case TileWall:
		t.Rune = '#'
		t.Color = 0x826E32 // RGB(130, 110, 50) - PyRogue仕様
	case TileFloor:
		t.Rune = '.'
		t.Color = 0x808080 // Gray - PyRogue風
	case TileDoor, TileDoorClosed:
		t.Rune = '+'
		t.Color = 0x8B4513 // Brown - PyRogue風
	case TileDoorOpen, TileOpenDoor:
		t.Rune = '/'
		t.Color = 0x8B4513 // Brown - PyRogue風
	case TileStairsUp:
		t.Rune = '<'
		t.Color = 0xFFFFFF // White - PyRogue風
	case TileStairsDown:
		t.Rune = '>'
		t.Color = 0xFFFFFF // White - PyRogue風
	case TileWater:
		t.Rune = '~'
		t.Color = 0x00FFFF // Cyan - PyRogue風
	case TileLava:
		t.Rune = '^'
		t.Color = 0xFF0000 // Red - PyRogue風
	case TileSecretDoor:
		t.Rune = '#'
		t.Color = 0x826E32 // RGB(130, 110, 50) - PyRogue仕様
	default:
		t.Rune = ' '
	}
	return t
}

// IsWalkable returns whether the tile can be walked on
func IsWalkable(t TileType) bool {
	switch t {
	case TileFloor, TileDoorOpen, TileOpenDoor, TileStairsUp, TileStairsDown:
		return true
	default:
		return false
	}
}
