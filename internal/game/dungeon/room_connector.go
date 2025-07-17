package dungeon

import (
	"math"
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// RoomConnector handles room connection using Rogue-style algorithm
type RoomConnector struct {
	level          *Level
	connectedRooms map[int]bool
	connections    map[int][]int
}

// NewRoomConnector creates a new room connector
func NewRoomConnector(level *Level) *RoomConnector {
	return &RoomConnector{
		level:          level,
		connectedRooms: make(map[int]bool),
		connections:    make(map[int][]int),
	}
}

// Connect connects all rooms using Rogue-style algorithm
func (c *RoomConnector) Connect() {
	if len(c.level.Rooms) < 2 {
		return
	}

	// 最初の部屋を接続済みとしてマーク
	c.connectedRooms[0] = true
	c.level.Rooms[0].Connected = true

	// 全ての部屋が接続されるまで繰り返す
	for len(c.connectedRooms) < len(c.level.Rooms) {
		bestPair := c.findBestRoomPair()
		if bestPair == nil {
			logger.Warn("Could not find room pair to connect")
			break
		}

		// 部屋を接続
		c.connectRoomPair(bestPair.fromIndex, bestPair.toIndex)
	}

	logger.Info("Connected rooms",
		"total_rooms", len(c.level.Rooms),
		"connected_rooms", len(c.connectedRooms),
	)
}

// roomPair represents a pair of rooms to connect
type roomPair struct {
	fromIndex int
	toIndex   int
	distance  float64
	isAdjacent bool
}

// findBestRoomPair finds the best pair of rooms to connect
func (c *RoomConnector) findBestRoomPair() *roomPair {
	var bestPair *roomPair
	minDistance := math.MaxFloat64

	// 未接続の部屋を探す
	for i := 0; i < len(c.level.Rooms); i++ {
		if c.connectedRooms[i] {
			continue
		}

		// この未接続の部屋に最も近い接続済みの部屋を探す
		for j := 0; j < len(c.level.Rooms); j++ {
			if !c.connectedRooms[j] || i == j {
				continue
			}

			// 隣接している部屋を優先
			if c.areRoomsAdjacent(i, j) {
				return &roomPair{
					fromIndex:  j,
					toIndex:    i,
					distance:   0,
					isAdjacent: true,
				}
			}

			// 距離を計算
			dist := c.calculateRoomDistance(i, j)
			if dist < minDistance {
				minDistance = dist
				bestPair = &roomPair{
					fromIndex:  j,
					toIndex:    i,
					distance:   dist,
					isAdjacent: false,
				}
			}
		}
	}

	return bestPair
}

// areRoomsAdjacent checks if two rooms are adjacent (share a wall)
func (c *RoomConnector) areRoomsAdjacent(i, j int) bool {
	r1 := c.level.Rooms[i]
	r2 := c.level.Rooms[j]

	// 水平方向の隣接チェック
	if r1.Y < r2.Y+r2.Height && r1.Y+r1.Height > r2.Y {
		// 右隣
		if r1.X+r1.Width+1 == r2.X {
			return true
		}
		// 左隣
		if r2.X+r2.Width+1 == r1.X {
			return true
		}
	}

	// 垂直方向の隣接チェック
	if r1.X < r2.X+r2.Width && r1.X+r1.Width > r2.X {
		// 下隣
		if r1.Y+r1.Height+1 == r2.Y {
			return true
		}
		// 上隣
		if r2.Y+r2.Height+1 == r1.Y {
			return true
		}
	}

	return false
}

// calculateRoomDistance calculates the distance between two rooms
func (c *RoomConnector) calculateRoomDistance(i, j int) float64 {
	r1 := c.level.Rooms[i]
	r2 := c.level.Rooms[j]

	// 部屋の中心点間の距離
	cx1 := float64(r1.X + r1.Width/2)
	cy1 := float64(r1.Y + r1.Height/2)
	cx2 := float64(r2.X + r2.Width/2)
	cy2 := float64(r2.Y + r2.Height/2)

	dx := cx2 - cx1
	dy := cy2 - cy1

	return math.Sqrt(dx*dx + dy*dy)
}

// connectRoomPair connects two rooms with a corridor
func (c *RoomConnector) connectRoomPair(fromIndex, toIndex int) {
	r1 := c.level.Rooms[fromIndex]
	r2 := c.level.Rooms[toIndex]

	// 隣接している場合は直接接続
	if c.areRoomsAdjacent(fromIndex, toIndex) {
		c.connectAdjacentRooms(r1, r2)
	} else {
		// L字型の通路で接続
		c.connectWithCorridor(r1, r2)
	}

	// 接続済みとしてマーク
	c.connectedRooms[toIndex] = true
	c.level.Rooms[toIndex].Connected = true

	// 接続情報を記録
	c.connections[fromIndex] = append(c.connections[fromIndex], toIndex)
	c.connections[toIndex] = append(c.connections[toIndex], fromIndex)

	logger.Debug("Connected rooms",
		"from", fromIndex,
		"to", toIndex,
		"adjacent", c.areRoomsAdjacent(fromIndex, toIndex),
	)
}

// connectAdjacentRooms connects two adjacent rooms with a doorway
func (c *RoomConnector) connectAdjacentRooms(r1, r2 *Room) {
	// 共有する壁の範囲を計算
	if r1.X+r1.Width+1 == r2.X || r2.X+r2.Width+1 == r1.X {
		// 水平方向に隣接
		minY := max(r1.Y, r2.Y) + 1
		maxY := min(r1.Y+r1.Height, r2.Y+r2.Height) - 1
		
		if minY <= maxY {
			// ランダムな位置に通路を作成
			y := minY + rand.Intn(maxY-minY+1)
			
			if r1.X+r1.Width+1 == r2.X {
				// r1が左、r2が右
				c.level.SetTile(r1.X+r1.Width, y, TileFloor)
				c.level.SetTile(r2.X-1, y, TileFloor)
			} else {
				// r2が左、r1が右
				c.level.SetTile(r2.X+r2.Width, y, TileFloor)
				c.level.SetTile(r1.X-1, y, TileFloor)
			}
		}
	} else {
		// 垂直方向に隣接
		minX := max(r1.X, r2.X) + 1
		maxX := min(r1.X+r1.Width, r2.X+r2.Width) - 1
		
		if minX <= maxX {
			// ランダムな位置に通路を作成
			x := minX + rand.Intn(maxX-minX+1)
			
			if r1.Y+r1.Height+1 == r2.Y {
				// r1が上、r2が下
				c.level.SetTile(x, r1.Y+r1.Height, TileFloor)
				c.level.SetTile(x, r2.Y-1, TileFloor)
			} else {
				// r2が上、r1が下
				c.level.SetTile(x, r2.Y+r2.Height, TileFloor)
				c.level.SetTile(x, r1.Y-1, TileFloor)
			}
		}
	}
}

// connectWithCorridor connects two rooms with an L-shaped corridor (PyRogue style)
func (c *RoomConnector) connectWithCorridor(r1, r2 *Room) {
	// 部屋の中心点を計算
	x1 := r1.X + r1.Width/2
	y1 := r1.Y + r1.Height/2
	x2 := r2.X + r2.Width/2
	y2 := r2.Y + r2.Height/2

	// PyRogue風の決定的なL字型通路生成
	// 距離が大きい方向を先に接続する
	dx := abs(x2 - x1)
	dy := abs(y2 - y1)
	
	if dx > dy {
		// 水平距離が大きい場合は水平優先
		c.level.CreateHorizontalCorridor(x1, x2, y1)
		c.level.CreateVerticalCorridor(y1, y2, x2)
	} else {
		// 垂直距離が大きい場合は垂直優先
		c.level.CreateVerticalCorridor(y1, y2, x1)
		c.level.CreateHorizontalCorridor(x1, x2, y2)
	}
	
	// 通路と部屋の接続点にドアを配置
	c.placeDoors(r1, r2, x1, y1, x2, y2)
}

// GetRoomConnections returns the connections for a room
func (c *RoomConnector) GetRoomConnections(roomIndex int) []int {
	return c.connections[roomIndex]
}

// IsRoomConnected checks if a room is connected
func (c *RoomConnector) IsRoomConnected(roomIndex int) bool {
	return c.connectedRooms[roomIndex]
}

// placeDoors places doors at corridor-room intersection points
func (c *RoomConnector) placeDoors(r1, r2 *Room, x1, y1, x2, y2 int) {
	// 通路と部屋の境界でドアを配置
	c.placeDoorAtRoomBoundary(r1, x1, y1, x2, y2)
	c.placeDoorAtRoomBoundary(r2, x1, y1, x2, y2)
}

// placeDoorAtRoomBoundary places a door where corridor meets room boundary
func (c *RoomConnector) placeDoorAtRoomBoundary(room *Room, x1, y1, x2, y2 int) {
	// 水平方向のドア配置
	if x1 < room.X && x2 >= room.X {
		// 左からの接続
		if y1 >= room.Y && y1 < room.Y+room.Height {
			c.level.SetTile(room.X-1, y1, TileDoor)
		}
	} else if x1 >= room.X+room.Width && x2 < room.X+room.Width {
		// 右からの接続
		if y1 >= room.Y && y1 < room.Y+room.Height {
			c.level.SetTile(room.X+room.Width, y1, TileDoor)
		}
	}
	
	// 垂直方向のドア配置
	if y1 < room.Y && y2 >= room.Y {
		// 上からの接続
		if x2 >= room.X && x2 < room.X+room.Width {
			c.level.SetTile(x2, room.Y-1, TileDoor)
		}
	} else if y1 >= room.Y+room.Height && y2 < room.Y+room.Height {
		// 下からの接続
		if x2 >= room.X && x2 < room.X+room.Width {
			c.level.SetTile(x2, room.Y+room.Height, TileDoor)
		}
	}
}

// abs returns the absolute value of an integer
func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}