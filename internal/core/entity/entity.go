package entity

// Entity represents a positioned object in the game world.
type Entity struct {
	Position *Position
}

// NewEntity creates an entity at a position.
func NewEntity(x, y int) *Entity {
	return &Entity{Position: NewPosition(x, y)}
}

// Move moves the entity by the given delta
func (e *Entity) Move(dx, dy int) {
	e.Position.Move(dx, dy)
}
