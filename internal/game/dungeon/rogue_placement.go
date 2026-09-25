package dungeon

import (
	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
)

func (l *Level) populateRogueRoom(room *Room) {
	hasGold := false
	if l.random().Intn(2) == 0 {
		gold := item.NewGoldForFloorWithRand(0, 0, l.FloorNumber, l.random())
		if position, ok := l.findRogueRoomFloor(room, 0, false); ok {
			gold.Position.X, gold.Position.Y = position.X, position.Y
			l.Items = append(l.Items, gold)
			hasGold = true
		}
	}

	chance := 25
	if hasGold {
		chance = 80
	}
	if l.random().Intn(100) < chance {
		if position, ok := l.findRogueRoomFloor(room, 0, true); ok {
			l.addRogueMonster(position, l.FloorNumber, false)
		}
	}
}

func (l *Level) addRogueMonster(position Position, floor int, mean bool) {
	symbol := actor.GetRandomMonsterTypeForFloorWithRand(floor, l.random())
	monster := actor.NewMonsterWithRandAndFloor(position.X, position.Y, symbol, floor, l.random())
	if monster == nil {
		return
	}
	if mean {
		monster.Mean = true
	}
	l.Monsters = append(l.Monsters, monster)
}

func (l *Level) SpawnItems() {
	if len(l.Rooms) == 0 {
		return
	}
	if l.random().Intn(20) == 0 {
		l.placeRogueTreasureRoom()
	}

	for range 9 {
		if l.random().Intn(100) >= 36 {
			continue
		}
		gameItem, noFood := item.NewRandomThingWithStateWithRand(0, 0, l.FloorNumber, l.NoFood, l.random())
		l.NoFood = noFood
		if gameItem == nil {
			continue
		}
		position, ok := l.findRogueFloor(nil, 0, false)
		if !ok {
			continue
		}
		gameItem.Position.X, gameItem.Position.Y = position.X, position.Y
		l.Items = append(l.Items, gameItem)
	}

	if l.FloorNumber >= 26 && !l.hasAmulet() {
		if position, ok := l.findRogueFloor(nil, 0, false); ok {
			amulet := item.NewAmulet(position.X, position.Y)
			l.Items = append(l.Items, amulet)
		}
	}
}

func (l *Level) placeRogueTreasureRoom() {
	room := l.Rooms[l.random().Intn(len(l.Rooms))]
	spots := min((room.Height-2)*(room.Width-2)-2, 8)
	objectCount := l.random().Intn(spots) + 2
	for range objectCount {
		position, ok := l.findRogueRoomFloor(room, 20, false)
		if !ok {
			continue
		}
		gameItem, noFood := item.NewRandomThingWithStateWithRand(position.X, position.Y, l.FloorNumber, l.NoFood, l.random())
		l.NoFood = noFood
		if gameItem != nil {
			l.Items = append(l.Items, gameItem)
		}
	}

	monsterCount := max(l.random().Intn(spots)+2, objectCount+2)
	roomSpots := (room.Height - 2) * (room.Width - 2)
	monsterCount = min(monsterCount, roomSpots)
	for range monsterCount {
		position, ok := l.findRogueRoomFloor(room, 10, true)
		if ok {
			l.addRogueMonster(position, l.FloorNumber+1, true)
		}
	}
}

func (l *Level) placeRogueTraps() {
	if len(l.Rooms) == 0 || l.random().Intn(10) >= l.FloorNumber {
		return
	}
	trapCount := min(rogueRand(l.random(), l.FloorNumber/4)+1, 10)
	for range trapCount {
		for {
			position, ok := l.findRogueFloor(nil, 0, false)
			if !ok || l.GetTile(position.X, position.Y).Type != TileFloor || l.hasTrapAt(position) {
				continue
			}
			l.Traps = append(l.Traps, &Trap{
				Type:     TrapType(l.random().Intn(8)),
				Position: position,
			})
			break
		}
	}
}

func (l *Level) findRogueFloor(room *Room, limit int, monster bool) (Position, bool) {
	if room != nil {
		return l.findRogueRoomFloor(room, limit, monster)
	}
	for attempts := 0; limit == 0 || attempts < limit; attempts++ {
		candidate := l.Rooms[l.random().Intn(len(l.Rooms))]
		position, ok := l.findRogueRoomFloor(candidate, 1, monster)
		if !ok {
			continue
		}
		if l.GetItemAt(position.X, position.Y) != nil || l.hasTrapAt(position) {
			continue
		}
		return position, true
	}
	return Position{}, false
}

func (l *Level) findRogueRoomFloor(room *Room, limit int, monster bool) (Position, bool) {
	for attempts := 0; limit == 0 || attempts < limit; attempts++ {
		position := Position{
			X: room.X + l.random().Intn(room.Width-2) + 1,
			Y: room.Y + l.random().Intn(room.Height-2) + 1,
		}
		tile := l.GetTile(position.X, position.Y)
		if tile == nil || (room.IsMaze && tile.Type != TilePassage) || (!room.IsMaze && tile.Type != TileFloor) {
			continue
		}
		if l.GetItemAt(position.X, position.Y) != nil || l.hasTrapAt(position) {
			continue
		}
		if monster && l.GetMonsterAt(position.X, position.Y) != nil {
			continue
		}
		return position, true
	}
	return Position{}, false
}

func (l *Level) hasAmulet() bool {
	for _, gameItem := range l.Items {
		if gameItem.Type == item.ItemAmulet {
			return true
		}
	}
	return false
}

func (l *Level) hasTrapAt(position Position) bool {
	for _, trap := range l.Traps {
		if trap.Position == position {
			return true
		}
	}
	return false
}

// SearchSecrets checks adjacent hidden doors and passage segments using Rogue's
// blindness and hallucination penalties. It returns the number revealed.
func (l *Level) SearchSecrets(x, y int, blind, hallucinating bool) int {
	if !l.IsInBounds(x, y) {
		return 0
	}
	penalty := 0
	if hallucinating {
		penalty += 3
	}
	if blind {
		penalty += 2
	}
	found := 0
	for nextY := y - 1; nextY <= y+1; nextY++ {
		for nextX := x - 1; nextX <= x+1; nextX++ {
			if (nextX == x && nextY == y) || !l.IsInBounds(nextX, nextY) {
				continue
			}
			tile := l.GetTile(nextX, nextY)
			switch tile.Type {
			case TileSecretDoor:
				if l.random().Intn(5+penalty) == 0 {
					l.SetTile(nextX, nextY, TileDoor)
					found++
				}
			case TileSecretPassage:
				if l.random().Intn(3+penalty) == 0 {
					l.SetTile(nextX, nextY, TilePassage)
					found++
				}
			}
		}
	}
	if found > 0 {
		l.UpdateVisibility(x, y)
	}
	return found
}
