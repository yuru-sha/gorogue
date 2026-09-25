package dungeon

import "math/rand"

type rogueRoomSlot struct {
	room   *Room
	x, y   int
	width  int
	height int
	gone   bool
	maze   bool
}

var rogueRoomAdjacency = [9][9]bool{
	{false, true, false, true, false, false, false, false, false},
	{true, false, true, false, true, false, false, false, false},
	{false, true, false, false, false, true, false, false, false},
	{true, false, false, false, true, false, true, false, false},
	{false, true, false, true, false, true, false, true, false},
	{false, false, true, false, true, false, false, false, true},
	{false, false, false, true, false, false, false, true, false},
	{false, false, false, false, true, false, true, false, true},
	{false, false, false, false, false, true, false, true, false},
}

func (l *Level) generateRogueRooms() [9]rogueRoomSlot {
	var slots [9]rogueRoomSlot
	cellWidth, cellHeight := l.Width/3, l.Height/3
	if cellWidth < 5 || cellHeight < 5 {
		return slots
	}

	for range l.random().Intn(4) {
		slots[l.random().Intn(len(slots))].gone = true
	}

	for i := range slots {
		slot := &slots[i]
		topX := (i%3)*cellWidth + 1
		topY := (i / 3) * cellHeight
		if slot.gone {
			slot.x = topX + l.random().Intn(cellWidth-2) + 1
			slot.y = topY + l.random().Intn(cellHeight-2) + 1
			for slot.y <= 0 || slot.y >= l.Height-1 {
				slot.x = topX + l.random().Intn(cellWidth-2) + 1
				slot.y = topY + l.random().Intn(cellHeight-2) + 1
			}
			continue
		}

		isDark := false
		isMaze := false
		if l.random().Intn(10) < l.FloorNumber-1 {
			isDark = true
			if l.random().Intn(15) == 0 {
				isDark = false
				isMaze = true
			}
		}

		if isMaze {
			slot.width = cellWidth - 1
			slot.height = cellHeight - 1
			slot.x = topX
			slot.y = topY
			if slot.x == 1 {
				slot.x = 0
			}
			if slot.y == 0 {
				slot.y++
				slot.height--
			}
		} else {
			for {
				slot.width = l.random().Intn(cellWidth-4) + 4
				slot.height = l.random().Intn(cellHeight-4) + 4
				slot.x = topX + l.random().Intn(cellWidth-slot.width)
				slot.y = topY + l.random().Intn(cellHeight-slot.height)
				if slot.y != 0 {
					break
				}
			}
		}

		slot.maze = isMaze
		slot.room = &Room{
			X:         slot.x,
			Y:         slot.y,
			Width:     slot.width,
			Height:    slot.height,
			IsDark:    isDark,
			IsMaze:    isMaze,
			Connected: true,
		}
		l.Rooms = append(l.Rooms, slot.room)
		l.drawRogueRoom(slot.room)
		l.populateRogueRoom(slot.room)
	}
	return slots
}

func (l *Level) drawRogueRoom(room *Room) {
	if room.IsMaze {
		l.digRogueMaze(room)
		return
	}
	for y := room.Y + 1; y < room.Y+room.Height-1; y++ {
		for x := room.X + 1; x < room.X+room.Width-1; x++ {
			l.SetTile(x, y, TileFloor)
		}
	}
}

func (l *Level) digRogueMaze(room *Room) {
	visited := make([][]bool, room.Height+1)
	for y := range visited {
		visited[y] = make([]bool, room.Width+1)
	}
	startY := (l.random().Intn(room.Height) / 2) * 2
	startX := (l.random().Intn(room.Width) / 2) * 2
	l.putRoguePassage(room.X+startX, room.Y+startY)
	visited[startY][startX] = true
	directions := [...]Position{{X: 0, Y: 2}, {X: 0, Y: -2}, {X: 2, Y: 0}, {X: -2, Y: 0}}

	var dig func(int, int)
	dig = func(y, x int) {
		for {
			count, next := 0, Position{X: -1, Y: -1}
			for _, delta := range directions {
				newX, newY := x+delta.X, y+delta.Y
				if newX < 0 || newX > room.Width || newY < 0 || newY > room.Height || visited[newY][newX] {
					continue
				}
				count++
				if l.random().Intn(count) == 0 {
					next = Position{X: newX, Y: newY}
				}
			}
			if count == 0 {
				return
			}
			visited[next.Y][next.X] = true
			l.putRoguePassage(room.X+(x+next.X)/2, room.Y+(y+next.Y)/2)
			l.putRoguePassage(room.X+next.X, room.Y+next.Y)
			dig(next.Y, next.X)
		}
	}
	dig(startY, startX)
}

func (l *Level) connectRogueRooms(slots *[9]rogueRoomSlot) {
	var inGraph [9]bool
	var connected [9][9]bool
	roomCount := 1
	current := l.random().Intn(len(slots))
	inGraph[current] = true

	for roomCount < len(slots) {
		count, selected := l.chooseRogueRoomCandidate(current, inGraph)
		if count == 0 {
			var eligibleCount int
			current, eligibleCount = l.chooseRogueRoomRoot(inGraph)
			if eligibleCount == 0 {
				return
			}
			continue
		}
		inGraph[selected] = true
		connected[current][selected] = true
		connected[selected][current] = true
		l.connectRogueRoomSlots(slots, current, selected)
		roomCount++
	}

	for extra := l.random().Intn(5); extra > 0; extra-- {
		current = l.random().Intn(len(slots))
		count, selected := l.chooseRogueExtraConnection(current, &connected)
		if count == 0 {
			continue
		}
		connected[current][selected] = true
		connected[selected][current] = true
		l.connectRogueRoomSlots(slots, current, selected)
	}
}

func (l *Level) chooseRogueRoomCandidate(current int, inGraph [9]bool) (count, selected int) {
	selected = -1
	for candidate := range inGraph {
		if rogueRoomAdjacency[current][candidate] && !inGraph[candidate] {
			count++
			if l.random().Intn(count) == 0 {
				selected = candidate
			}
		}
	}
	return count, selected
}

func (l *Level) chooseRogueRoomRoot(inGraph [9]bool) (current, eligibleCount int) {
	current = -1
	for candidate := range inGraph {
		if !inGraph[candidate] {
			continue
		}
		for neighbor := range inGraph {
			if rogueRoomAdjacency[candidate][neighbor] && !inGraph[neighbor] {
				eligibleCount++
				if l.random().Intn(eligibleCount) == 0 {
					current = candidate
				}
				break
			}
		}
	}
	return current, eligibleCount
}

func (l *Level) chooseRogueExtraConnection(current int, connected *[9][9]bool) (count, selected int) {
	selected = -1
	for candidate := range connected[current] {
		if rogueRoomAdjacency[current][candidate] && !connected[current][candidate] {
			count++
			if l.random().Intn(count) == 0 {
				selected = candidate
			}
		}
	}
	return count, selected
}

func (l *Level) connectRogueRoomSlots(slots *[9]rogueRoomSlot, from, to int) {
	if from > to {
		from, to = to, from
	}
	startRoom, endRoom := slots[from], slots[to]
	startX, startY := l.rogueRoomConnectionEndpoint(startRoom, from+1 != to, true)
	endX, endY := l.rogueRoomConnectionEndpoint(endRoom, from+1 != to, false)
	dx, dy := 1, 0
	if from+1 != to {
		dx, dy = 0, 1
	}

	distance := absInt(startY-endY) - 1
	turnX, turnY := 0, 0
	turnDistance := absInt(startX - endX)
	if dy == 1 {
		if startX < endX {
			turnX = 1
		} else {
			turnX = -1
		}
	} else {
		if startY < endY {
			turnY = 1
		} else {
			turnY = -1
		}
		distance = absInt(startX-endX) - 1
		turnDistance = absInt(startY - endY)
	}
	turnSpot := rogueRand(l.random(), distance-1) + 1

	if startRoom.gone {
		l.putRoguePassage(startX, startY)
	} else {
		l.putRogueDoor(startRoom.room, startX, startY)
	}
	if endRoom.gone {
		l.putRoguePassage(endX, endY)
	} else {
		l.putRogueDoor(endRoom.room, endX, endY)
	}

	x, y := startX, startY
	for distance > 0 {
		x += dx
		y += dy
		if distance == turnSpot {
			for turnDistance > 0 {
				l.putRoguePassage(x, y)
				x += turnX
				y += turnY
				turnDistance--
			}
		}
		l.putRoguePassage(x, y)
		distance--
	}
	l.putRoguePassage(x+dx, y+dy)
}

func (l *Level) rogueRoomConnectionEndpoint(room rogueRoomSlot, vertical, from bool) (x, y int) {
	x, y = room.x, room.y
	if room.gone {
		return x, y
	}
	if room.maze {
		return l.rogueMazeDoorPosition(room, vertical, from)
	}
	switch {
	case vertical && from:
		x += l.random().Intn(room.width-2) + 1
		y += room.height - 1
	case vertical:
		x += l.random().Intn(room.width-2) + 1
	case from:
		x += room.width - 1
		y += l.random().Intn(room.height-2) + 1
	default:
		y += l.random().Intn(room.height-2) + 1
	}
	return x, y
}

func (l *Level) rogueMazeDoorPosition(room rogueRoomSlot, vertical, from bool) (x, y int) {
	x, y = room.x, room.y
	if vertical {
		x += l.random().Intn(room.width-2) + 1
		if from {
			y += room.height - 1
		}
		insideY := y + 1
		if from {
			insideY = y - 1
		}
		if tile := l.GetTile(x, y); tile == nil || tile.Type != TilePassage {
			l.SetTile(x, y, TilePassage)
		}
		if tile := l.GetTile(x, insideY); tile == nil || tile.Type != TilePassage {
			l.SetTile(x, insideY, TilePassage)
		}
		return x, y
	}

	y += l.random().Intn(room.height-2) + 1
	if from {
		x += room.width - 1
	}
	insideX := x + 1
	if from {
		insideX = x - 1
	}
	if tile := l.GetTile(x, y); tile == nil || tile.Type != TilePassage {
		l.SetTile(x, y, TilePassage)
	}
	if tile := l.GetTile(insideX, y); tile == nil || tile.Type != TilePassage {
		l.SetTile(insideX, y, TilePassage)
	}
	return x, y
}

func (l *Level) putRoguePassage(x, y int) {
	tileType := TilePassage
	if l.random().Intn(10)+1 < l.FloorNumber && l.random().Intn(40) == 0 {
		tileType = TileSecretPassage
	}
	l.SetTile(x, y, tileType)
}

func (l *Level) putRogueDoor(room *Room, x, y int) {
	if room.IsMaze {
		return
	}
	tileType := TileDoor
	if l.random().Intn(10)+1 < l.FloorNumber && l.random().Intn(5) == 0 {
		tileType = TileSecretDoor
	}
	l.SetTile(x, y, tileType)
}

func rogueRand(rng *rand.Rand, n int) int {
	if n <= 0 {
		return 0
	}
	return rng.Intn(n)
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
