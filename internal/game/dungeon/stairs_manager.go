package dungeon

import (
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

	// 上り階段は1階では地上への出口になる。
	if s.level.FloorNumber >= 1 {
		s.placeUpStairs()
	}

	// 下り階段の配置（最終階層を除く）
	if s.level.FloorNumber < 26 {
		s.placeDownStairs()
	}

	logger.Debug("Placed stairs",
		"floor", s.level.FloorNumber,
		"up_stairs", s.level.FloorNumber >= 1,
		"down_stairs", s.level.FloorNumber < 26,
	)
}

func (s *StairsManager) placeUpStairs() {
	s.placeRogueStairs(TileStairsUp)
}

func (s *StairsManager) placeDownStairs() {
	s.placeRogueStairs(TileStairsDown)
}

func (s *StairsManager) placeRogueStairs(stairType TileType) {
	position, ok := s.level.findRogueFloor(nil, 0, false)
	if ok {
		s.level.SetTile(position.X, position.Y, stairType)
	}
}

// GetStairPositions returns the positions of stairs in the level
func (s *StairsManager) GetStairPositions() (upStairs, downStairs []Position) {
	for y := 0; y < s.level.Height; y++ {
		for x := 0; x < s.level.Width; x++ {
			tile := s.level.GetTile(x, y)
			switch tile.Type {
			case TileStairsUp:
				upStairs = append(upStairs, Position{X: x, Y: y})
			case TileStairsDown:
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
