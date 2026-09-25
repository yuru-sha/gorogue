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
		maxInt(l.Width, l.Height),
		func(p gruid.Point) bool {
			tile := l.GetTile(p.X, p.Y)
			return tile != nil && tile.Walkable()
		},
		true,
	)
	room := l.roomAtPosition(x, y)
	for _, position := range visible {
		if room != nil && room.IsDark && (absInt(position.X-x) > 1 || absInt(position.Y-y) > 1) {
			continue
		}
		tile := l.GetTile(position.X, position.Y)
		if tile != nil {
			tile.Visible = true
			tile.Explored = true
		}
	}
}

// LightRoomAt permanently lights the room containing the source location.
// Corridors only glow temporarily, so they have no stored light state.
func (l *Level) LightRoomAt(x, y int) bool {
	room := l.roomAtPosition(x, y)
	if room == nil {
		return false
	}
	room.IsDark = false
	l.UpdateVisibility(x, y)
	return true
}

func (l *Level) roomAtPosition(x, y int) *Room {
	for _, room := range l.Rooms {
		if x >= room.X && x < room.X+room.Width && y >= room.Y && y < room.Y+room.Height {
			return room
		}
	}
	return nil
}
