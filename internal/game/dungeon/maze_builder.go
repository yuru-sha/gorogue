package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// MazeBuilder generates maze-type levels for special floors (7, 13, 19)
type MazeBuilder struct {
	width      int
	height     int
	floorNum   int
	level      *Level
	visited    [][]bool
	complexity float64 // 迷路の複雑さ（0.0-1.0）
}

// NewMazeBuilder creates a new maze builder
func NewMazeBuilder(width, height, floorNum int) *MazeBuilder {
	// 階層に応じた迷路の複雑さを設定
	complexity := 0.3 // 基本複雑さ
	switch floorNum {
	case 7:
		complexity = 0.4 // 中程度の複雑さ
	case 13:
		complexity = 0.6 // 高い複雑さ
	case 19:
		complexity = 0.8 // 最高の複雑さ
	}

	return &MazeBuilder{
		width:      width,
		height:     height,
		floorNum:   floorNum,
		complexity: complexity,
		visited:    make([][]bool, height),
	}
}

// Build creates a maze-type level
func (mb *MazeBuilder) Build() *Level {
	// 初期化
	mb.level = &Level{
		Width:       mb.width,
		Height:      mb.height,
		Tiles:       make([][]*Tile, mb.height),
		Rooms:       make([]*Room, 0),
		FloorNumber: mb.floorNum,
		Monsters:    make([]*actor.Monster, 0),
		Items:       make([]*item.Item, 0),
	}

	// visited配列を初期化
	for i := range mb.visited {
		mb.visited[i] = make([]bool, mb.width)
	}

	// タイルを初期化（全て壁）
	for y := 0; y < mb.height; y++ {
		mb.level.Tiles[y] = make([]*Tile, mb.width)
		for x := 0; x < mb.width; x++ {
			mb.level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	// 迷路を生成
	mb.generateMaze()

	// 迷路から部屋を抽出（階段配置のため）
	mb.extractRoomsFromMaze()

	// 階段の配置
	mb.placeStairs()

	// モンスターの配置
	mb.level.SpawnMonsters()

	// アイテムの配置
	mb.level.SpawnItems()

	logger.Info("Generated maze level",
		"floor", mb.floorNum,
		"complexity", mb.complexity,
		"rooms", len(mb.level.Rooms),
		"monsters", len(mb.level.Monsters),
		"items", len(mb.level.Items),
	)

	return mb.level
}

// generateMaze generates the maze using recursive backtracking algorithm
func (mb *MazeBuilder) generateMaze() {
	// 迷路生成のための開始点を設定
	startX, startY := 1, 1
	mb.carvePath(startX, startY)

	// 複雑さに応じて追加の通路を生成
	mb.addComplexity()

	logger.Debug("Maze generation completed",
		"floor", mb.floorNum,
		"complexity", mb.complexity,
	)
}

// carvePath carves a path through the maze using recursive backtracking
func (mb *MazeBuilder) carvePath(x, y int) {
	// 現在の位置を床にする
	mb.level.SetTile(x, y, TileFloor)
	mb.visited[y][x] = true

	// 4方向のランダムな順序で試す
	directions := [][]int{{0, -2}, {2, 0}, {0, 2}, {-2, 0}} // 上、右、下、左（2マス間隔）

	// 方向をランダムにシャッフル
	for i := len(directions) - 1; i > 0; i-- {
		j := rand.Intn(i + 1)
		directions[i], directions[j] = directions[j], directions[i]
	}

	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]

		// 境界チェック
		if nx < 1 || nx >= mb.width-1 || ny < 1 || ny >= mb.height-1 {
			continue
		}

		// 未訪問の場合
		if !mb.visited[ny][nx] {
			// 間のタイルも床にする
			wallX, wallY := x+dir[0]/2, y+dir[1]/2
			mb.level.SetTile(wallX, wallY, TileFloor)

			// 再帰的に続行
			mb.carvePath(nx, ny)
		}
	}
}

// addComplexity adds additional complexity to the maze
func (mb *MazeBuilder) addComplexity() {
	// 複雑さに応じて追加の通路を生成
	numExtraPassages := int(float64(mb.width*mb.height) * mb.complexity * 0.01)

	for i := 0; i < numExtraPassages; i++ {
		// ランダムな壁を選択
		x := 1 + rand.Intn(mb.width-2)
		y := 1 + rand.Intn(mb.height-2)

		// 壁の場合、通路に変更する可能性がある
		if mb.level.GetTile(x, y).Type == TileWall {
			// 周囲に床があるかチェック
			if mb.hasAdjacentFloor(x, y) {
				mb.level.SetTile(x, y, TileFloor)
			}
		}
	}
}

// hasAdjacentFloor checks if there's a floor tile adjacent to the given position
func (mb *MazeBuilder) hasAdjacentFloor(x, y int) bool {
	directions := [][]int{{0, -1}, {1, 0}, {0, 1}, {-1, 0}} // 上、右、下、左

	for _, dir := range directions {
		nx, ny := x+dir[0], y+dir[1]
		if nx >= 0 && nx < mb.width && ny >= 0 && ny < mb.height {
			if mb.level.GetTile(nx, ny).Type == TileFloor {
				return true
			}
		}
	}
	return false
}

// extractRoomsFromMaze extracts room-like areas from the maze for stair placement
func (mb *MazeBuilder) extractRoomsFromMaze() {
	// 迷路から部屋のような領域を抽出
	// 階段配置のために必要

	// 開始地点と終了地点を部屋として登録
	startRoom := &Room{
		X:         1,
		Y:         1,
		Width:     3,
		Height:    3,
		Connected: true,
	}
	mb.level.Rooms = append(mb.level.Rooms, startRoom)

	// 終了地点を探す（迷路の反対側）
	endX, endY := mb.width-3, mb.height-3
	endRoom := &Room{
		X:         endX,
		Y:         endY,
		Width:     3,
		Height:    3,
		Connected: true,
	}
	mb.level.Rooms = append(mb.level.Rooms, endRoom)

	// 迷路の中央付近にも部屋を配置
	centerX, centerY := mb.width/2, mb.height/2
	centerRoom := &Room{
		X:         centerX - 1,
		Y:         centerY - 1,
		Width:     3,
		Height:    3,
		Connected: true,
	}
	mb.level.Rooms = append(mb.level.Rooms, centerRoom)

	logger.Debug("Extracted rooms from maze",
		"total_rooms", len(mb.level.Rooms),
		"floor", mb.floorNum,
	)
}

// placeStairs places stairs in the maze
func (mb *MazeBuilder) placeStairs() {
	if len(mb.level.Rooms) == 0 {
		return
	}

	// 上り階段の配置（最初の階層を除く）
	if mb.floorNum > 1 {
		firstRoom := mb.level.Rooms[0]
		mb.level.SetTile(
			firstRoom.X+firstRoom.Width/2,
			firstRoom.Y+firstRoom.Height/2,
			TileStairsUp,
		)
	}

	// 下り階段の配置（最終階層を除く）
	if mb.floorNum < 26 {
		lastRoom := mb.level.Rooms[len(mb.level.Rooms)-1]
		mb.level.SetTile(
			lastRoom.X+lastRoom.Width/2,
			lastRoom.Y+lastRoom.Height/2,
			TileStairsDown,
		)
	}

	logger.Debug("Placed stairs in maze",
		"floor", mb.floorNum,
		"up_stairs", mb.floorNum > 1,
		"down_stairs", mb.floorNum < 26,
	)
}
