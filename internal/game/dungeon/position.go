package dungeon

// Position represents a position in the dungeon
type Position struct {
	X, Y int
}

// NewPosition creates a new position
func NewPosition(x, y int) Position {
	return Position{X: x, Y: y}
}

// IsEqual checks if two positions are equal
func (p Position) IsEqual(other Position) bool {
	return p.X == other.X && p.Y == other.Y
}

// Distance calculates the Manhattan distance to another position
func (p Position) Distance(other Position) int {
	dx := p.X - other.X
	dy := p.Y - other.Y
	if dx < 0 {
		dx = -dx
	}
	if dy < 0 {
		dy = -dy
	}
	return dx + dy
}
