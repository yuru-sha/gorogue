package dungeon

import (
	"math/rand"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// DungeonBuilder is responsible for building dungeon levels
type DungeonBuilder struct {
	level         *Level
	roomConnector *RoomConnector
}

// NewDungeonBuilder creates a new dungeon builder
func NewDungeonBuilder(width, height, floorNum int) *DungeonBuilder {
	level := &Level{
		Width:       width,
		Height:      height,
		FloorNumber: floorNum,
		Rooms:       make([]*Room, 0),
		Monsters:    make([]*actor.Monster, 0),
		Items:       make([]*item.Item, 0),
	}

	// Initialize tiles with walls
	level.Tiles = make([][]*Tile, height)
	for y := range level.Tiles {
		level.Tiles[y] = make([]*Tile, width)
		for x := range level.Tiles[y] {
			level.Tiles[y][x] = NewTile(TileWall)
		}
	}

	return &DungeonBuilder{
		level:         level,
		roomConnector: NewRoomConnector(level),
	}
}

// Build builds the dungeon level using PyRogue-style dynamic system
func (b *DungeonBuilder) Build() *Level {
	// PyRogue風の階層に応じたダンジョンタイプの決定
	dungeonType := b.determineDungeonType()
	
	switch dungeonType {
	case "maze":
		b.generateMaze()
		logger.Info("Built maze dungeon", "floor", b.level.FloorNumber)
	case "bsp":
		b.generateRoomsWithBSP()
		logger.Info("Built BSP dungeon", "floor", b.level.FloorNumber, "rooms", len(b.level.Rooms))
	default:
		b.generateRoomsWithBSP()
		logger.Info("Built default BSP dungeon", "floor", b.level.FloorNumber, "rooms", len(b.level.Rooms))
	}

	// 特別な部屋の生成（迷路以外）
	if dungeonType != "maze" && b.shouldGenerateSpecialRoom() {
		b.generateSpecialRoom()
	}

	// 孤立部屋群の生成（特定階層）
	if b.shouldGenerateIsolatedRooms() {
		b.generateIsolatedRooms()
	}

	// 暗い部屋の生成（深い階層）
	if b.shouldGenerateDarkRooms() {
		b.generateDarkRooms()
	}

	// 階段の配置
	b.placeStairs()

	// モンスターの配置
	b.spawnMonsters()

	// アイテムの配置
	b.spawnItems()

	logger.Info("Built dungeon level",
		"floor", b.level.FloorNumber,
		"type", dungeonType,
		"rooms", len(b.level.Rooms),
		"monsters", len(b.level.Monsters),
		"items", len(b.level.Items),
	)

	return b.level
}

// determineDungeonType determines the dungeon type based on floor number (PyRogue style)
func (b *DungeonBuilder) determineDungeonType() string {
	floor := b.level.FloorNumber
	
	// PyRogue風の階層別ダンジョンタイプ
	switch {
	case floor == 7 || floor == 13 || floor == 19:
		return "maze"
	default:
		return "bsp"
	}
}

// shouldGenerateIsolatedRooms determines if isolated rooms should be generated
func (b *DungeonBuilder) shouldGenerateIsolatedRooms() bool {
	floor := b.level.FloorNumber
	
	// PyRogue風の孤立部屋群生成判定
	switch {
	case floor <= 3:
		return false // 浅い階層では生成しない
	case floor <= 10:
		return floor == 4 || floor == 8 // 4階、8階で確実に生成
	case floor <= 20:
		return floor == 11 || floor == 15 || floor == 18 // 11階、15階、18階で確実に生成
	default:
		return floor == 22 || floor == 25 // 22階、25階で確実に生成
	}
}

// shouldGenerateDarkRooms determines if dark rooms should be generated
func (b *DungeonBuilder) shouldGenerateDarkRooms() bool {
	floor := b.level.FloorNumber
	
	// PyRogue風の暗い部屋生成判定
	switch {
	case floor <= 5:
		return false // 浅い階層では生成しない
	case floor <= 12:
		return floor == 6 || floor == 10 // 6階、10階で確実に生成
	case floor <= 20:
		return floor == 14 || floor == 17 || floor == 20 // 14階、17階、20階で確実に生成
	default:
		return floor == 23 || floor == 24 // 深層では23階、24階で確実に生成
	}
}

// generateMaze generates a maze-type dungeon (PyRogue style)
func (b *DungeonBuilder) generateMaze() {
	mazeGenerator := NewMazeGenerator(b.level)
	mazeGenerator.GenerateMaze()
	
	// Create a single "room" representing the entire maze for stair placement
	mazeRoom := &Room{
		X:         1,
		Y:         1,
		Width:     b.level.Width - 2,
		Height:    b.level.Height - 2,
		IsSpecial: false,
		Connected: true,
	}
	b.level.Rooms = append(b.level.Rooms, mazeRoom)
}

// generateIsolatedRooms generates isolated room groups (PyRogue style)
func (b *DungeonBuilder) generateIsolatedRooms() {
	// Simple implementation: add 1-2 small isolated rooms
	for i := 0; i < 1+rand.Intn(2); i++ {
		for attempts := 0; attempts < 50; attempts++ {
			width := 4 + rand.Intn(4)  // 4-7 tiles wide
			height := 4 + rand.Intn(4) // 4-7 tiles high
			x := 2 + rand.Intn(b.level.Width-width-4)
			y := 2 + rand.Intn(b.level.Height-height-4)
			
			if b.canPlaceIsolatedRoom(x, y, width, height) {
				room := &Room{
					X:         x,
					Y:         y,
					Width:     width,
					Height:    height,
					IsSpecial: false,
					Connected: false, // Isolated rooms start as disconnected
				}
				b.level.Rooms = append(b.level.Rooms, room)
				
				// Create the room
				for dy := 0; dy < height; dy++ {
					for dx := 0; dx < width; dx++ {
						if b.level.IsInBounds(x+dx, y+dy) {
							b.level.SetTile(x+dx, y+dy, TileFloor)
						}
					}
				}
				
				// Connect to main dungeon with a secret passage
				b.createSecretPassage(room)
				break
			}
		}
	}
}

// generateDarkRooms applies darkness to some rooms (PyRogue style)
func (b *DungeonBuilder) generateDarkRooms() {
	// Apply darkness to 30-50% of rooms
	darkRoomCount := len(b.level.Rooms) * (30 + rand.Intn(21)) / 100
	
	// Shuffle rooms and make some of them dark
	shuffledRooms := make([]*Room, len(b.level.Rooms))
	copy(shuffledRooms, b.level.Rooms)
	rand.Shuffle(len(shuffledRooms), func(i, j int) {
		shuffledRooms[i], shuffledRooms[j] = shuffledRooms[j], shuffledRooms[i]
	})
	
	for i := 0; i < darkRoomCount && i < len(shuffledRooms); i++ {
		room := shuffledRooms[i]
		room.IsSpecial = true // Mark as special to indicate it's dark
		
		// Place a light source in the room (torch or similar)
		lightX := room.X + room.Width/2
		lightY := room.Y + room.Height/2
		if b.level.IsInBounds(lightX, lightY) {
			// Light source placement would go here
			// For now, just log it
			logger.Debug("Placed light source in dark room", "x", lightX, "y", lightY)
		}
	}
}

// canPlaceIsolatedRoom checks if an isolated room can be placed
func (b *DungeonBuilder) canPlaceIsolatedRoom(x, y, width, height int) bool {
	// Check bounds
	if x < 1 || y < 1 || x+width >= b.level.Width-1 || y+height >= b.level.Height-1 {
		return false
	}
	
	// Check for minimum distance from existing rooms
	minDistance := 3
	for _, room := range b.level.Rooms {
		if abs(x-room.X) < minDistance+width || abs(y-room.Y) < minDistance+height {
			return false
		}
	}
	
	// Check that the area is currently walls
	for dy := -1; dy <= height; dy++ {
		for dx := -1; dx <= width; dx++ {
			if b.level.IsInBounds(x+dx, y+dy) {
				if b.level.GetTile(x+dx, y+dy).Type != TileWall {
					return false
				}
			}
		}
	}
	
	return true
}

// createSecretPassage creates a secret passage to connect isolated room
func (b *DungeonBuilder) createSecretPassage(room *Room) {
	// Find the nearest main room
	var nearestRoom *Room
	minDistance := float64(b.level.Width + b.level.Height)
	
	for _, r := range b.level.Rooms {
		if r != room && r.Connected {
			distance := float64(abs(r.X-room.X) + abs(r.Y-room.Y))
			if distance < minDistance {
				minDistance = distance
				nearestRoom = r
			}
		}
	}
	
	if nearestRoom != nil {
		// Create a simple passage
		startX := room.X + room.Width/2
		startY := room.Y + room.Height/2
		endX := nearestRoom.X + nearestRoom.Width/2
		endY := nearestRoom.Y + nearestRoom.Height/2
		
		// Create L-shaped passage
		for x := min(startX, endX); x <= max(startX, endX); x++ {
			if b.level.IsInBounds(x, startY) {
				b.level.SetTile(x, startY, TileFloor)
			}
		}
		for y := min(startY, endY); y <= max(startY, endY); y++ {
			if b.level.IsInBounds(endX, y) {
				b.level.SetTile(endX, y, TileFloor)
			}
		}
		
		room.Connected = true
	}
}

// generateRooms generates rooms for the dungeon (PyRogue style)
func (b *DungeonBuilder) generateRooms() {
	numRooms := MinRooms + rand.Intn(MaxRooms-MinRooms+1)
	
	for i := 0; i < numRooms; i++ {
		for attempts := 0; attempts < 100; attempts++ {
			width := MinRoomSize + rand.Intn(MaxRoomSize-MinRoomSize+1)
			height := MinRoomSize + rand.Intn(MaxRoomSize-MinRoomSize+1)
			x := 1 + rand.Intn(b.level.Width-width-2)
			y := 1 + rand.Intn(b.level.Height-height-2)

			if b.canPlaceRoom(x, y, width, height) {
				// PyRogue風の「Gone Room」機能
				// 10-15%の確率で通路のみの空間を作成
				if rand.Float64() < 0.12 {
					b.createGoneRoom(x, y, width, height)
				} else {
					room := &Room{
						X:      x,
						Y:      y,
						Width:  width,
						Height: height,
					}
					b.addRoom(room)
				}
				break
			}
		}
	}

	logger.Debug("Generated rooms", "count", len(b.level.Rooms))
}

// generateRoomsWithGrid generates rooms using the original Rogue 3x3 grid system
func (b *DungeonBuilder) generateRoomsWithGrid() {
	gridGenerator := NewGridGenerator(b.level)
	gridGenerator.GenerateRooms()
	
	logger.Debug("Generated rooms with 3x3 grid system", "count", len(b.level.Rooms))
}

// generateRoomsWithBSP generates rooms using PyRogue-style BSP system
func (b *DungeonBuilder) generateRoomsWithBSP() {
	bspGenerator := NewBSPGenerator(b.level)
	bspGenerator.GenerateRooms()
	
	logger.Debug("Generated rooms with BSP system", "count", len(b.level.Rooms))
}

// canPlaceRoom checks if a room can be placed at the given position
func (b *DungeonBuilder) canPlaceRoom(x, y, width, height int) bool {
	// 部屋の周囲1マスも含めてチェック
	for dy := -1; dy <= height; dy++ {
		for dx := -1; dx <= width; dx++ {
			nx, ny := x+dx, y+dy
			if !b.level.IsInBounds(nx, ny) {
				return false
			}
			if b.level.GetTile(nx, ny).Type != TileWall {
				return false
			}
		}
	}
	return true
}

// addRoom adds a room to the level
func (b *DungeonBuilder) addRoom(room *Room) {
	// Fill room with floor tiles
	for y := room.Y; y < room.Y+room.Height; y++ {
		for x := room.X; x < room.X+room.Width; x++ {
			if b.level.IsInBounds(x, y) {
				b.level.SetTile(x, y, TileFloor)
			}
		}
	}

	// Add room to the list
	b.level.Rooms = append(b.level.Rooms, room)

	logger.Debug("Added room",
		"index", len(b.level.Rooms)-1,
		"x", room.X,
		"y", room.Y,
		"width", room.Width,
		"height", room.Height,
	)
}

// createGoneRoom creates a "gone room" - corridor-only space (PyRogue style)
func (b *DungeonBuilder) createGoneRoom(x, y, width, height int) {
	// Create a corridor-only space instead of a room
	// Fill with floor tiles but don't add to rooms list
	for dy := 0; dy < height; dy++ {
		for dx := 0; dx < width; dx++ {
			if b.level.IsInBounds(x+dx, y+dy) {
				b.level.SetTile(x+dx, y+dy, TileFloor)
			}
		}
	}
	
	// Add a few scattered floor tiles around the area for organic feel
	for attempt := 0; attempt < 5; attempt++ {
		extraX := x + rand.Intn(width)
		extraY := y + rand.Intn(height)
		
		// Extend randomly in one direction
		direction := rand.Intn(4)
		switch direction {
		case 0: // North
			if extraY > 0 {
				b.level.SetTile(extraX, extraY-1, TileFloor)
			}
		case 1: // East
			if extraX < b.level.Width-1 {
				b.level.SetTile(extraX+1, extraY, TileFloor)
			}
		case 2: // South
			if extraY < b.level.Height-1 {
				b.level.SetTile(extraX, extraY+1, TileFloor)
			}
		case 3: // West
			if extraX > 0 {
				b.level.SetTile(extraX-1, extraY, TileFloor)
			}
		}
	}
	
	logger.Debug("Created gone room (corridor space)",
		"x", x,
		"y", y,
		"width", width,
		"height", height,
	)
}

// connectRooms connects all rooms using Rogue-style algorithm
func (b *DungeonBuilder) connectRooms() {
	if len(b.level.Rooms) < 2 {
		return
	}

	b.roomConnector.Connect()
}

// placeDoors places doors at room entrances
func (b *DungeonBuilder) placeDoors() {
	doorPlacer := NewDoorPlacer(b.level)
	doorPlacer.PlaceDoors()
}

// shouldGenerateSpecialRoom determines if a special room should be generated
func (b *DungeonBuilder) shouldGenerateSpecialRoom() bool {
	// 1階では特別な部屋を生成しない
	if b.level.FloorNumber <= 1 {
		return false
	}

	// 5階ごとに10%の確率で生成
	if b.level.FloorNumber%5 == 0 && rand.Float64() < 0.1 {
		return true
	}

	return false
}

// generateSpecialRoom generates a special room
func (b *DungeonBuilder) generateSpecialRoom() {
	// 既に特別な部屋が存在する場合は生成しない
	for _, room := range b.level.Rooms {
		if room.IsSpecial {
			return
		}
	}

	// 5x5の特別な部屋を生成
	for attempts := 0; attempts < 100; attempts++ {
		x := 1 + rand.Intn(b.level.Width-7)
		y := 1 + rand.Intn(b.level.Height-7)

		if b.canPlaceRoom(x, y, 5, 5) {
			room := &Room{
				X:         x,
				Y:         y,
				Width:     5,
				Height:    5,
				IsSpecial: true,
			}
			b.addRoom(room)

			// 秘密のドアを配置
			b.placeSecretDoor(room)

			// 部屋の内容を生成
			b.populateSpecialRoom(room)

			logger.Info("Generated special room",
				"floor", b.level.FloorNumber,
				"x", x,
				"y", y,
			)
			return
		}
	}
}

// placeSecretDoor places a secret door for a special room
func (b *DungeonBuilder) placeSecretDoor(room *Room) {
	doorPlacer := NewDoorPlacer(b.level)
	doorPlacer.PlaceSecretDoor(room)
}

// populateSpecialRoom populates a special room with content
func (b *DungeonBuilder) populateSpecialRoom(room *Room) {
	// 部屋の種類をランダムに決定
	roomType := rand.Intn(6)

	switch roomType {
	case 0: // 宝物庫
		logger.Info("Generating treasure vault")
		b.populateTreasureVault(room)
	case 1: // 武器庫
		logger.Info("Generating armory")
		b.populateArmory(room)
	case 2: // 食料庫
		logger.Info("Generating food storage")
		b.populateFoodStorage(room)
	case 3: // 魔物のねぐら
		logger.Info("Generating monster lair")
		b.populateMonsterLair(room)
	case 4: // 実験室
		logger.Info("Generating laboratory")
		b.populateLaboratory(room)
	case 5: // 図書室
		logger.Info("Generating library")
		b.populateLibrary(room)
	}
}

// populateTreasureVault populates a treasure vault
func (b *DungeonBuilder) populateTreasureVault(room *Room) {
	// 部屋の中央にゴールドを配置
	cx, cy := room.X+room.Width/2, room.Y+room.Height/2
	goldItem := item.NewGold(cx, cy, true) // 特別な部屋のゴールド
	if goldItem != nil {
		goldItem.Value *= 3 // 3倍の価値
		b.level.Items = append(b.level.Items, goldItem)
	}

	// 周囲に追加の宝物を配置
	for i := 0; i < 2+rand.Intn(3); i++ {
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)
		if b.level.IsValidItemPosition(x, y) {
			// 高価なアイテムを配置
			itemTypes := []item.ItemType{item.ItemRing, item.ItemWeapon, item.ItemArmor}
			itemType := itemTypes[rand.Intn(len(itemTypes))]
			newItem := b.createHighValueItem(x, y, itemType)
			if newItem != nil {
				b.level.Items = append(b.level.Items, newItem)
			}
		}
	}
}

// populateArmory populates an armory
func (b *DungeonBuilder) populateArmory(room *Room) {
	// 武器と防具を配置
	for i := 0; i < 3+rand.Intn(3); i++ {
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)
		if b.level.IsValidItemPosition(x, y) {
			var itemType item.ItemType
			if rand.Float64() < 0.5 {
				itemType = item.ItemWeapon
			} else {
				itemType = item.ItemArmor
			}
			newItem := b.createHighValueItem(x, y, itemType)
			if newItem != nil {
				b.level.Items = append(b.level.Items, newItem)
			}
		}
	}
}

// populateFoodStorage populates a food storage room
func (b *DungeonBuilder) populateFoodStorage(room *Room) {
	// 食料を大量に配置
	for i := 0; i < 4+rand.Intn(4); i++ {
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)
		if b.level.IsValidItemPosition(x, y) {
			newItem := item.NewFood(x, y)
			if newItem != nil {
				b.level.Items = append(b.level.Items, newItem)
			}
		}
	}
}

// populateMonsterLair populates a monster lair
func (b *DungeonBuilder) populateMonsterLair(room *Room) {
	// 強力なモンスターを配置
	cx, cy := room.X+room.Width/2, room.Y+room.Height/2
	
	// ボスモンスターを中央に配置
	if b.level.GetMonsterAt(cx, cy) == nil {
		bossType := b.selectBossMonsterType()
		boss := actor.NewMonster(cx, cy, bossType)
		b.scaleBossMonster(boss)
		b.level.Monsters = append(b.level.Monsters, boss)
	}

	// 周囲に雑魚モンスターを配置
	for i := 0; i < 2+rand.Intn(2); i++ {
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)
		if b.level.GetMonsterAt(x, y) == nil && b.level.IsWalkable(x, y) {
			monsterType := b.level.selectMonsterType()
			monster := actor.NewMonster(x, y, monsterType)
			b.level.scaleMonsterForFloor(monster)
			b.level.Monsters = append(b.level.Monsters, monster)
		}
	}
}

// populateLaboratory populates a laboratory
func (b *DungeonBuilder) populateLaboratory(room *Room) {
	// 薬を配置
	for i := 0; i < 3+rand.Intn(3); i++ {
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)
		if b.level.IsValidItemPosition(x, y) {
			newItem := item.NewRandomPotion(x, y)
			if newItem != nil {
				b.level.Items = append(b.level.Items, newItem)
			}
		}
	}
}

// populateLibrary populates a library
func (b *DungeonBuilder) populateLibrary(room *Room) {
	// 巻物を配置
	for i := 0; i < 3+rand.Intn(3); i++ {
		x := room.X + 1 + rand.Intn(room.Width-2)
		y := room.Y + 1 + rand.Intn(room.Height-2)
		if b.level.IsValidItemPosition(x, y) {
			newItem := item.NewRandomScroll(x, y)
			if newItem != nil {
				b.level.Items = append(b.level.Items, newItem)
			}
		}
	}
}

// placeStairs places the stairs in the dungeon
func (b *DungeonBuilder) placeStairs() {
	stairsManager := NewStairsManager(b.level)
	stairsManager.PlaceStairs()
}

// spawnMonsters delegates to level's SpawnMonsters
func (b *DungeonBuilder) spawnMonsters() {
	b.level.SpawnMonsters()
}

// spawnItems delegates to level's SpawnItems
func (b *DungeonBuilder) spawnItems() {
	b.level.SpawnItems()
}

// createHighValueItem creates a high-value item of the specified type
func (b *DungeonBuilder) createHighValueItem(x, y int, itemType item.ItemType) *item.Item {
	baseItem := b.level.createRandomItem(x, y, itemType)
	if baseItem != nil {
		// 価値を2-3倍にする
		multiplier := 2 + rand.Float64()
		baseItem.Value = int(float64(baseItem.Value) * multiplier)
	}
	return baseItem
}

// selectBossMonsterType selects a boss monster type for special rooms
func (b *DungeonBuilder) selectBossMonsterType() rune {
	// 階層に応じたボスモンスター
	switch {
	case b.level.FloorNumber <= 10:
		bosses := []rune{'O', 'T'} // オーガ、トロール
		return bosses[rand.Intn(len(bosses))]
	case b.level.FloorNumber <= 20:
		bosses := []rune{'T', 'D'} // トロール、ドラゴン
		return bosses[rand.Intn(len(bosses))]
	default:
		return 'D' // ドラゴン
	}
}

// scaleBossMonster scales a boss monster's stats
func (b *DungeonBuilder) scaleBossMonster(monster *actor.Monster) {
	// 通常のスケーリングを適用
	b.level.scaleMonsterForFloor(monster)
	
	// ボス用の追加スケーリング（2倍）
	monster.MaxHP *= 2
	monster.HP = monster.MaxHP
	monster.Attack = int(float64(monster.Attack) * 1.5)
	monster.Defense = int(float64(monster.Defense) * 1.5)
	
	logger.Debug("Scaled boss monster",
		"type", monster.Type.Name,
		"hp", monster.HP,
		"attack", monster.Attack,
		"defense", monster.Defense,
	)
}