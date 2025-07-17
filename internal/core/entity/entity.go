package entity

import "github.com/anaseto/gruid"

// Entity represents any object in the game world
type Entity struct {
	Position *Position
	Symbol   rune
	Color    gruid.Color
}

// NewEntity creates a new Entity
func NewEntity(x, y int, symbol rune, color gruid.Color) *Entity {
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
