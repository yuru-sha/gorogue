package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// StairsManager handles stair placement in the dungeon
type StairsManager struct {
	level *Level
}

// NewStairsManager creates a new stairs manager
func NewStairsManager(level *Level) *StairsManager {
	return &StairsManager{
		level: level,
	}
}

// PlaceStairs places stairs in the dungeon following Pyrogue's pattern
func (s *StairsManager) PlaceStairs() {
	if len(s.level.Rooms) == 0 {
		return
	}

	// 上り階段の配置（最初の階層を除く）
	if s.level.FloorNumber > 1 {
		s.placeUpStairs()
	}

	// 下り階段の配置（最終階層を除く）
	if s.level.FloorNumber < 26 {
		s.placeDownStairs()
	}

	logger.Debug("Placed stairs",
		"floor", s.level.FloorNumber,
		"up_stairs", s.level.FloorNumber > 1,
		"down_stairs", s.level.FloorNumber < 26,
	)
}

// placeUpStairs places up stairs in the first room
func (s *StairsManager) placeUpStairs() {
	// 上り階段は最初の部屋（接続済み）に配置
	firstRoom := s.level.Rooms[0]

	// 部屋の中央に配置を試みる
	centerX := firstRoom.X + firstRoom.Width/2
	centerY := firstRoom.Y + firstRoom.Height/2

	if s.isValidStairPosition(centerX, centerY) {
		s.level.SetTile(centerX, centerY, TileStairsUp)
		logger.Debug("Placed up stairs in center of first room",
			"x", centerX,
			"y", centerY,
		)
		return
	}

	// 中央に配置できない場合は、部屋内のランダムな位置に配置
	s.placeStairsInRoom(firstRoom, TileStairsUp)
}

// placeDownStairs places down stairs in the last connected room
func (s *StairsManager) placeDownStairs() {
	// 下り階段は最後の接続済み部屋に配置
	var lastConnectedRoom *Room
	for i := len(s.level.Rooms) - 1; i >= 0; i-- {
		if s.level.Rooms[i].Connected {
			lastConnectedRoom = s.level.Rooms[i]
			break
		}
	}

	if lastConnectedRoom == nil {
		// フォールバック: 最後の部屋を使用
		lastConnectedRoom = s.level.Rooms[len(s.level.Rooms)-1]
	}

	// 部屋の中央に配置を試みる
	centerX := lastConnectedRoom.X + lastConnectedRoom.Width/2
	centerY := lastConnectedRoom.Y + lastConnectedRoom.Height/2

	if s.isValidStairPosition(centerX, centerY) {
		s.level.SetTile(centerX, centerY, TileStairsDown)
		logger.Debug("Placed down stairs in center of last connected room",
			"x", centerX,
			"y", centerY,
		)
		return
	}

	// 中央に配置できない場合は、部屋内のランダムな位置に配置
	s.placeStairsInRoom(lastConnectedRoom, TileStairsDown)
}

// placeStairsInRoom places stairs in a specific room
func (s *StairsManager) placeStairsInRoom(room *Room, stairType TileType) {
	maxAttempts := 20

	for attempts := 0; attempts < maxAttempts; attempts++ {
		// 部屋の境界から1マス内側の範囲でランダムな位置を選択
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)

		if s.isValidStairPosition(x, y) {
			s.level.SetTile(x, y, stairType)
			logger.Debug("Placed stairs in room",
				"type", stairType,
				"x", x,
				"y", y,
				"attempts", attempts+1,
			)
			return
		}
	}

	// 全ての試行が失敗した場合のフォールバック
	// 部屋の左上角に配置
	x := room.X + 1
	y := room.Y + 1

	if s.level.IsInBounds(x, y) {
		s.level.SetTile(x, y, stairType)
		logger.Warn("Placed stairs at fallback position",
			"type", stairType,
			"x", x,
			"y", y,
		)
	}
}

// isValidStairPosition checks if a position is valid for stair placement
func (s *StairsManager) isValidStairPosition(x, y int) bool {
	// 境界チェック
	if !s.level.IsInBounds(x, y) {
		return false
	}

	// 床タイルかチェック
	tile := s.level.GetTile(x, y)
	if tile.Type != TileFloor {
		return false
	}

	// 既に階段がある位置かチェック
	if tile.Type == TileStairsUp || tile.Type == TileStairsDown {
		return false
	}

	// モンスターがいないかチェック
	if s.level.GetMonsterAt(x, y) != nil {
		return false
	}

	// アイテムがないかチェック
	if s.level.GetItemAt(x, y) != nil {
		return false
	}

	return true
}

// GetStairPositions returns the positions of stairs in the level
func (s *StairsManager) GetStairPositions() (upStairs, downStairs []Position) {
	for y := 0; y < s.level.Height; y++ {
		for x := 0; x < s.level.Width; x++ {
			tile := s.level.GetTile(x, y)
			if tile.Type == TileStairsUp {
				upStairs = append(upStairs, Position{X: x, Y: y})
			} else if tile.Type == TileStairsDown {
				downStairs = append(downStairs, Position{X: x, Y: y})
			}
		}
	}
	return
}

// IsStairPosition checks if a position contains stairs
func (s *StairsManager) IsStairPosition(x, y int) bool {
	if !s.level.IsInBounds(x, y) {
		return false
	}

	tile := s.level.GetTile(x, y)
	return tile.Type == TileStairsUp || tile.Type == TileStairsDown
}

// GetStairType returns the type of stairs at the given position
func (s *StairsManager) GetStairType(x, y int) TileType {
	if !s.level.IsInBounds(x, y) {
		return TileWall
	}

	tile := s.level.GetTile(x, y)
	if tile.Type == TileStairsUp || tile.Type == TileStairsDown {
		return tile.Type
	}

	return TileWall
}
