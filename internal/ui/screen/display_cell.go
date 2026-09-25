package screen

import (
	"strings"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
)

type displayEntityKind uint8

const (
	displayEntityNone displayEntityKind = iota
	displayEntityItem
	displayEntityTrap
	displayEntityMonster
	displayEntityPlayer
)

// displayCell carries game-facing display facts without terminal glyphs or colors.
type displayCell struct {
	X, Y        int
	Terrain     dungeon.TileType
	Visible     bool
	Explored    bool
	Entity      displayEntityKind
	Priority    uint8
	ItemType    gameitem.ItemType
	MonsterCode rune
}

func convertDisplayCells(cells []displayCell, level *dungeon.Level, player *actor.Player) []displayCell {
	count := level.Width * level.Height
	if cap(cells) < count {
		cells = make([]displayCell, count)
	} else {
		cells = cells[:count]
	}
	for y := 0; y < level.Height; y++ {
		for x := 0; x < level.Width; x++ {
			index := y*level.Width + x
			cell := displayCell{X: x, Y: y}
			if tile := level.GetTile(x, y); tile != nil {
				cell.Terrain = tile.Type
				cell.Visible = tile.Visible
				cell.Explored = tile.Explored
			}
			cells[index] = cell
		}
	}
	overlayItems(cells, level)
	overlayTraps(cells, level)
	overlayMonsters(cells, level, player)
	if player != nil && player.Position != nil {
		overlayDisplayEntity(cells, level.Width, level.Height, player.Position.X, player.Position.Y, displayEntityPlayer, 4)
	}
	return cells
}

func overlayItems(cells []displayCell, level *dungeon.Level) {
	for _, item := range level.Items {
		if item == nil {
			continue
		}
		x, y := item.Position.X, item.Position.Y
		if !inDisplayBounds(x, y, level.Width, level.Height) || !cells[y*level.Width+x].Visible {
			continue
		}
		if cell := overlayDisplayEntity(cells, level.Width, level.Height, x, y, displayEntityItem, 1); cell != nil {
			cell.ItemType = item.Type
		}
	}
}

func overlayTraps(cells []displayCell, level *dungeon.Level) {
	for _, trap := range level.Traps {
		if trap == nil || !trap.Discovered {
			continue
		}
		x, y := trap.Position.X, trap.Position.Y
		if !inDisplayBounds(x, y, level.Width, level.Height) {
			continue
		}
		current := &cells[y*level.Width+x]
		if current.Visible || current.Explored {
			overlayDisplayEntity(cells, level.Width, level.Height, x, y, displayEntityTrap, 2)
		}
	}
}

func overlayMonsters(cells []displayCell, level *dungeon.Level, player *actor.Player) {
	for _, monster := range level.Monsters {
		if monster == nil || !monster.IsAlive() || !canSeeInvisible(player, monster) {
			continue
		}
		x, y := monster.Position.X, monster.Position.Y
		if !inDisplayBounds(x, y, level.Width, level.Height) || !cells[y*level.Width+x].Visible {
			continue
		}
		if cell := overlayDisplayEntity(cells, level.Width, level.Height, x, y, displayEntityMonster, 3); cell != nil {
			cell.MonsterCode = monster.Type.Code
		}
	}
}

func canSeeInvisible(player *actor.Player, monster *actor.Monster) bool {
	if player == nil {
		return false
	}
	if !monster.IsInvisible || player.SeeInvisibleTurns > 0 {
		return true
	}
	if player.Equipment == nil {
		return false
	}
	for _, ring := range []*gameitem.Item{player.Equipment.RingLeft, player.Equipment.RingRight} {
		if ring == nil {
			continue
		}
		name := ring.Name
		if ring.RealName != "" {
			name = ring.RealName
		}
		if strings.EqualFold(name, "see invisible") {
			return true
		}
	}
	return false
}

func inDisplayBounds(x, y, width, height int) bool {
	return x >= 0 && y >= 0 && x < width && y < height
}

func overlayDisplayEntity(cells []displayCell, width, height, x, y int, kind displayEntityKind, priority uint8) *displayCell {
	if !inDisplayBounds(x, y, width, height) {
		return nil
	}
	cell := &cells[y*width+x]
	if priority <= cell.Priority {
		return nil
	}
	cell.Entity = kind
	cell.Priority = priority
	return cell
}
