package dungeon

import (
	"github.com/anaseto/gruid"
	"github.com/anaseto/gruid/rl"
)

// UpdateVisibility recalculates the player's current field of view and keeps
// previously visible terrain marked as explored.
func (l *Level) UpdateVisibility(x, y int) {
	for _, row := range l.Tiles {
		for _, tile := range row {
			if tile != nil {
				tile.Visible = false
			}
		}
	}
	if !l.IsInBounds(x, y) {
		return
	}

	fov := rl.NewFOV(gruid.NewRange(0, 0, l.Width, l.Height))
	visible := fov.SSCVisionMap(
		gruid.Point{X: x, Y: y},
		max(l.Width, l.Height),
		func(p gruid.Point) bool {
			tile := l.GetTile(p.X, p.Y)
			return tile != nil && tile.Walkable()
		},
		true,
	)
	for _, position := range visible {
		tile := l.GetTile(position.X, position.Y)
		if tile != nil {
			tile.Visible = true
			tile.Explored = true
		}
	}
}
