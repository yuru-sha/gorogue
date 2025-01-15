package dungeon

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
	Symbol   rune
	Color    [3]uint8
	Visible  bool
	Walkable bool
}

// GetTileSymbol returns the symbol for a given tile type
func GetTileSymbol(t TileType) rune {
	switch t {
	case TileWall:
		return '#'
	case TileFloor:
		return '.'
	case TileDoorClosed:
		return '+'
	case TileDoorOpen:
		return '/'
	case TileStairsUp:
		return '<'
	case TileStairsDown:
		return '>'
	case TileWater:
		return '~'
	case TileLava:
		return '^'
	case TileSecretDoor:
		return '#' // 未発見時は壁と同じ
	default:
		return ' '
	}
}

// GetTileColor returns the color for a given tile type
func GetTileColor(t TileType) [3]uint8 {
	switch t {
	case TileWall:
		return [3]uint8{128, 128, 128} // Gray
	case TileFloor:
		return [3]uint8{128, 128, 128} // Gray
	case TileDoorClosed, TileDoorOpen:
		return [3]uint8{139, 69, 19} // Brown
	case TileStairsUp, TileStairsDown:
		return [3]uint8{255, 255, 255} // White
	case TileWater:
		return [3]uint8{0, 0, 255} // Blue
	case TileLava:
		return [3]uint8{255, 0, 0} // Red
	case TileSecretDoor:
		return [3]uint8{128, 128, 128} // Gray (同じく壁と同じ)
	default:
		return [3]uint8{0, 0, 0} // Black
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
		Symbol:   GetTileSymbol(tileType),
		Color:    GetTileColor(tileType),
		Visible:  false,
		Walkable: IsWalkable(tileType),
	}
}
