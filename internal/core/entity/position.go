package entity

// Position represents the position of an entity in the game world
type Position struct {
	X, Y int
}

// NewPosition creates a new Position
func NewPosition(x, y int) *Position {
	return &Position{X: x, Y: y}
}

// Move updates the position by the given delta
func (p *Position) Move(dx, dy int) {
	p.X += dx
	p.Y += dy
}
