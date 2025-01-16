package dungeon

import "github.com/anaseto/gruid"

// TileType represents different types of tiles in the dungeon
type TileType int

const (
	TileWall TileType = iota
	TileFloor
	TileDoorClosed
	TileDoorOpen
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
	Type     TileType
	Cell     gruid.Cell
	Visible  bool
	Walkable bool
}

// GetTileCell returns the cell configuration for a given tile type
func GetTileCell(t TileType) gruid.Cell {
	switch t {
	case TileWall:
		return gruid.Cell{
			Rune:  '#',
			Style: gruid.Style{Fg: 8}, // Gray
		}
	case TileFloor:
		return gruid.Cell{
			Rune:  '.',
			Style: gruid.Style{Fg: 8}, // Gray
		}
	case TileDoorClosed:
		return gruid.Cell{
			Rune:  '+',
			Style: gruid.Style{Fg: 3}, // Brown
		}
	case TileDoorOpen:
		return gruid.Cell{
			Rune:  '/',
			Style: gruid.Style{Fg: 3}, // Brown
		}
	case TileStairsUp:
		return gruid.Cell{
			Rune:  '<',
			Style: gruid.Style{Fg: 15}, // White
		}
	case TileStairsDown:
		return gruid.Cell{
			Rune:  '>',
			Style: gruid.Style{Fg: 15}, // White
		}
	case TileWater:
		return gruid.Cell{
			Rune:  '~',
			Style: gruid.Style{Fg: 4}, // Blue
		}
	case TileLava:
		return gruid.Cell{
			Rune:  '^',
			Style: gruid.Style{Fg: 1}, // Red
		}
	case TileSecretDoor:
		return gruid.Cell{
			Rune:  '#',
			Style: gruid.Style{Fg: 8}, // Gray (同じく壁と同じ)
		}
	default:
		return gruid.Cell{
			Rune:  ' ',
			Style: gruid.Style{},
		}
	}
}

// IsWalkable returns whether the tile can be walked on
func IsWalkable(t TileType) bool {
	switch t {
	case TileFloor, TileDoorOpen, TileStairsUp, TileStairsDown:
		return true
	default:
		return false
	}
}

// NewTile creates a new tile of the given type
func NewTile(tileType TileType) *Tile {
	return &Tile{
		Type:     tileType,
		Cell:     GetTileCell(tileType),
		Visible:  false,
		Walkable: IsWalkable(tileType),
	}
}
