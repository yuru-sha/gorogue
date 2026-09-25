package dungeon

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
	TilePassage
	TileSecretPassage
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
	case TilePassage:
		return "passage"
	case TileSecretPassage:
		return "secret_passage"
	case TileLava:
		return "lava"
	default:
		return "unknown"
	}
}

// Tile stores logical terrain and exploration state.
type Tile struct {
	Type       TileType
	Visible    bool
	Explored   bool
	IsWalkable bool
}

// Walkable returns whether the tile can be walked on.
func (t *Tile) Walkable() bool {
	return t.IsWalkable
}

// NewTile creates a logical tile of the given terrain type.
func NewTile(tileType TileType) *Tile {
	return &Tile{
		Type:       tileType,
		IsWalkable: IsWalkable(tileType),
	}
}

// IsWalkable returns whether the tile can be walked on
func IsWalkable(t TileType) bool {
	switch t {
	case TileFloor, TilePassage, TileDoorOpen, TileOpenDoor, TileStairsUp, TileStairsDown:
	default:
		return false
	}
	return true
}
