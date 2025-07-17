package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// DoorPlacer handles door placement in the dungeon
type DoorPlacer struct {
	level *Level
}

// NewDoorPlacer creates a new door placer
func NewDoorPlacer(level *Level) *DoorPlacer {
	return &DoorPlacer{
		level: level,
	}
}

// PlaceDoors places doors at appropriate locations
func (d *DoorPlacer) PlaceDoors() {
	for i, room := range d.level.Rooms {
		if room.IsSpecial {
			continue // 特別な部屋は秘密のドアを使用
		}

		// 通路と部屋の境界にドアを配置
		d.placeDoorForRoom(i, room)
	}
}

// placeDoorForRoom places doors for a specific room
func (d *DoorPlacer) placeDoorForRoom(roomIndex int, room *Room) {
	doorPositions := d.findDoorPositions(room)

	for _, pos := range doorPositions {
		// 15%の確率で秘密のドアを作成
		if rand.Float64() < 0.15 {
			d.level.SetTile(pos.X, pos.Y, TileSecretDoor)
			logger.Debug("Placed secret door",
				"room", roomIndex,
				"x", pos.X,
				"y", pos.Y,
			)
		} else {
			d.level.SetTile(pos.X, pos.Y, TileDoor)
			logger.Debug("Placed door",
				"room", roomIndex,
				"x", pos.X,
				"y", pos.Y,
			)
		}
	}
}

// findDoorPositions finds positions where doors should be placed
func (d *DoorPlacer) findDoorPositions(room *Room) []Position {
	var positions []Position

	// 部屋の境界をチェック
	for x := room.X; x < room.X+room.Width; x++ {
		// 上の境界
		if d.shouldPlaceDoor(x, room.Y-1, room) {
			positions = append(positions, Position{X: x, Y: room.Y - 1})
		}
		// 下の境界
		if d.shouldPlaceDoor(x, room.Y+room.Height, room) {
			positions = append(positions, Position{X: x, Y: room.Y + room.Height})
		}
	}

	for y := room.Y; y < room.Y+room.Height; y++ {
		// 左の境界
		if d.shouldPlaceDoor(room.X-1, y, room) {
			positions = append(positions, Position{X: room.X - 1, Y: y})
		}
		// 右の境界
		if d.shouldPlaceDoor(room.X+room.Width, y, room) {
			positions = append(positions, Position{X: room.X + room.Width, Y: y})
		}
	}

	return positions
}

// shouldPlaceDoor determines if a door should be placed at the given position
func (d *DoorPlacer) shouldPlaceDoor(x, y int, room *Room) bool {
	// 境界チェック
	if !d.level.IsInBounds(x, y) {
		return false
	}

	// 現在の位置が壁でない場合はドアを配置しない
	if d.level.GetTile(x, y).Type != TileWall {
		return false
	}

	// 隣接する位置が通路かチェック
	adjacentPositions := []Position{
		{x - 1, y}, {x + 1, y}, {x, y - 1}, {x, y + 1},
	}

	for _, pos := range adjacentPositions {
		if d.level.IsInBounds(pos.X, pos.Y) {
			tile := d.level.GetTile(pos.X, pos.Y)
			if tile.Type == TileFloor {
				// この位置が部屋の内部でない場合
				if !d.isInsideRoom(pos.X, pos.Y, room) {
					return true
				}
			}
		}
	}

	return false
}

// isInsideRoom checks if a position is inside the given room
func (d *DoorPlacer) isInsideRoom(x, y int, room *Room) bool {
	return x >= room.X && x < room.X+room.Width &&
		y >= room.Y && y < room.Y+room.Height
}

// PlaceSecretDoor places a secret door for a special room
func (d *DoorPlacer) PlaceSecretDoor(room *Room) {
	// 部屋の4辺のいずれかにランダムに秘密のドアを配置
	side := rand.Intn(4)
	var x, y int

	switch side {
	case 0: // 上辺
		x = room.X + rand.Intn(room.Width)
		y = room.Y - 1
	case 1: // 右辺
		x = room.X + room.Width
		y = room.Y + rand.Intn(room.Height)
	case 2: // 下辺
		x = room.X + rand.Intn(room.Width)
		y = room.Y + room.Height
	case 3: // 左辺
		x = room.X - 1
		y = room.Y + rand.Intn(room.Height)
	}

	if d.level.IsInBounds(x, y) {
		d.level.SetTile(x, y, TileSecretDoor)
		logger.Debug("Placed secret door for special room",
			"x", x,
			"y", y,
		)
	}
}

// OpenDoor opens a door at the given position
func (d *DoorPlacer) OpenDoor(x, y int) bool {
	if !d.level.IsInBounds(x, y) {
		return false
	}

	tile := d.level.GetTile(x, y)
	if tile.Type == TileDoor {
		d.level.SetTile(x, y, TileOpenDoor)
		logger.Debug("Opened door", "x", x, "y", y)
		return true
	}

	return false
}

// CloseDoor closes a door at the given position
func (d *DoorPlacer) CloseDoor(x, y int) bool {
	if !d.level.IsInBounds(x, y) {
		return false
	}

	tile := d.level.GetTile(x, y)
	if tile.Type == TileOpenDoor {
		d.level.SetTile(x, y, TileDoor)
		logger.Debug("Closed door", "x", x, "y", y)
		return true
	}

	return false
}

// RevealSecretDoor reveals a secret door
func (d *DoorPlacer) RevealSecretDoor(x, y int) bool {
	if !d.level.IsInBounds(x, y) {
		return false
	}

	tile := d.level.GetTile(x, y)
	if tile.Type == TileSecretDoor {
		d.level.SetTile(x, y, TileDoor)
		logger.Info("Revealed secret door", "x", x, "y", y)
		return true
	}

	return false
}
