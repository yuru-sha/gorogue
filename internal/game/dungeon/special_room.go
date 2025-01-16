package dungeon

import (
	"math/rand"
)

// SpecialRoomType represents different types of special rooms
type SpecialRoomType int

const (
	RoomTreasure SpecialRoomType = iota
	RoomArmory
	RoomFood
	RoomMonster
	RoomLaboratory
	RoomLibrary
)

// SpecialRoom represents a special room with unique properties
type SpecialRoom struct {
	Room
	Type SpecialRoomType
}

// GenerateSpecialRoom generates a special room if conditions are met
func (l *Level) GenerateSpecialRoom() {
	if !l.ShouldGenerateSpecialRoom() {
		return
	}

	// 特別な部屋の種類をランダムに選択
	roomType := SpecialRoomType(rand.Intn(6))

	// 5x5の部屋を生成
	for attempts := 0; attempts < 100; attempts++ {
		x := 1 + rand.Intn(l.Width-7)  // 周囲1マス + 5マス + 1マス
		y := 1 + rand.Intn(l.Height-7) // 周囲1マス + 5マス + 1マス

		if l.CanPlaceRoom(x, y, 5, 5) {
			room := &Room{
				X:         x,
				Y:         y,
				Width:     5,
				Height:    5,
				IsSpecial: true,
			}
			l.AddRoom(room)

			// 隠し扉を配置
			l.PlaceSecretDoor(room)

			// 部屋の内容を生成
			l.PopulateSpecialRoom(room, roomType)
			return
		}
	}
}

// PlaceSecretDoor places a secret door for the special room
func (l *Level) PlaceSecretDoor(room *Room) {
	// 部屋の各辺の中央に隠し扉を配置する位置を決定
	possibleDoors := []struct{ x, y int }{
		{room.X + room.Width/2, room.Y - 1},           // 上
		{room.X + room.Width/2, room.Y + room.Height}, // 下
		{room.X - 1, room.Y + room.Height/2},          // 左
		{room.X + room.Width, room.Y + room.Height/2}, // 右
	}

	// ランダムな位置に隠し扉を配置
	doorPos := possibleDoors[rand.Intn(len(possibleDoors))]
	if l.IsInBounds(doorPos.x, doorPos.y) {
		l.SetTile(doorPos.x, doorPos.y, TileSecretDoor)
	}
}

// PopulateSpecialRoom populates a special room with content
func (l *Level) PopulateSpecialRoom(room *Room, roomType SpecialRoomType) {
	// TODO: 各部屋タイプに応じたアイテムやモンスターの配置を実装
	// 現在は部屋の種類に応じたメッセージのみを出力
	switch roomType {
	case RoomTreasure:
		// 宝物庫: 大量のゴールド、珍しい武器や防具、貴重な指輪
	case RoomArmory:
		// 武器庫: 珍しい武器、高性能な防具
	case RoomFood:
		// 食料庫: 大量の食料
	case RoomMonster:
		// 魔物のねぐら: 強力なモンスターが複数出現
	case RoomLaboratory:
		// 実験室: ランダムな薬
	case RoomLibrary:
		// 図書室: ランダムな巻物
	}
}
