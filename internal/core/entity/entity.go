package entity

// Entity represents any object in the game world
type Entity struct {
	Position *Position
	Symbol   rune
	Color    [3]uint8
}

// NewEntity creates a new Entity
func NewEntity(x, y int, symbol rune, color [3]uint8) *Entity {
	return &Entity{
		Position: NewPosition(x, y),
		Symbol:   symbol,
		Color:    color,
	}
}

// Move moves the entity by the given delta
func (e *Entity) Move(dx, dy int) {
	e.Position.Move(dx, dy)
}
