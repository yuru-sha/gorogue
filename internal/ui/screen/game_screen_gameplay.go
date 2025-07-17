package screen

import (
	"fmt"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// tryMovePlayer attempts to move the player in the given direction
func (s *GameScreen) tryMovePlayer(dx, dy int) {
	newX := s.player.Position.X + dx
	newY := s.player.Position.Y + dy

	// 境界チェック
	if newX < 0 || newX >= s.level.Width || newY < 0 || newY >= s.level.Height {
		logger.Debug("Player movement blocked by bounds",
			"current_x", s.player.Position.X,
			"current_y", s.player.Position.Y,
			"new_x", newX,
			"new_y", newY,
		)
		return
	}

	// 壁の衝突判定
	tile := s.level.GetTile(newX, newY)
	if !tile.Walkable() {
		logger.Debug("Player movement blocked by wall",
			"current_x", s.player.Position.X,
			"current_y", s.player.Position.Y,
			"new_x", newX,
			"new_y", newY,
			"tile_type", tile.Type,
		)
		return
	}

	// モンスターとの戦闘判定
	monster := s.level.GetMonsterAt(newX, newY)
	if monster != nil {
		s.playerAttackMonster(monster)
		return
	}

	// 移動実行
	s.player.Position.Move(dx, dy)
	logger.Debug("Player moved",
		"new_x", s.player.Position.X,
		"new_y", s.player.Position.Y,
	)

	// アイテムを拾う処理
	s.pickupItem(newX, newY)

	// モンスターのターンを実行
	s.level.UpdateMonsters(s.player)
}

// playerAttackMonster handles player attacking a monster
func (s *GameScreen) playerAttackMonster(monster *actor.Monster) {
	damage := s.player.CalculateDamage(monster.Defense)
	monster.TakeDamage(damage)

	message := fmt.Sprintf("%sに%dのダメージを与えた！", monster.Type.Name, damage)
	s.AddMessage(message)

	if !monster.IsAlive() {
		deathMessage := fmt.Sprintf("%sを倒した！", monster.Type.Name)
		s.AddMessage(deathMessage)

		// 経験値とゴールドを取得
		exp := monster.MaxHP + monster.Attack
		gold := monster.MaxHP / 2

		s.player.GainExp(exp)
		s.player.AddGold(gold)

		rewardMessage := fmt.Sprintf("%d経験値、%dゴールドを得た", exp, gold)
		s.AddMessage(rewardMessage)
	} else {
		// モンスターのターンを実行
		s.level.UpdateMonsters(s.player)
	}
}

// pickupItem handles picking up an item at the given position
func (s *GameScreen) pickupItem(x, y int) {
	item := s.level.GetItemAt(x, y)
	if item == nil {
		return
	}

	// インベントリに追加を試行
	if !s.player.Inventory.AddItem(item) {
		s.AddMessage("Your pack is full!")
		return
	}

	// アイテムタイプに応じたメッセージ（識別状態を考慮）
	displayName := s.player.IdentifyMgr.GetDisplayName(item)
	switch item.Type {
	case gameitem.ItemGold:
		s.AddMessage(fmt.Sprintf("You found %d gold pieces", item.Value))
	case gameitem.ItemAmulet:
		s.AddMessage(fmt.Sprintf("You picked up the %s!", displayName))
	default:
		s.AddMessage(fmt.Sprintf("You picked up %s", displayName))
	}

	// アイテムをレベルから削除
	s.level.RemoveItem(item)
}

// handleLook handles the look/examine command
func (s *GameScreen) handleLook() {
	s.AddMessage("Looking around... (not fully implemented yet)")
	// TODO: Implement look functionality - show what's in adjacent squares
}

// handlePickUp handles picking up items at current position
func (s *GameScreen) handlePickUp() {
	x, y := s.player.Position.X, s.player.Position.Y
	item := s.level.GetItemAt(x, y)

	if item == nil {
		s.AddMessage("There is nothing here to pick up.")
		return
	}

	// Try to add to inventory
	if !s.player.Inventory.AddItem(item) {
		s.AddMessage("Your pack is full!")
		return
	}

	// Show pickup message
	displayName := s.player.IdentifyMgr.GetDisplayName(item)
	switch item.Type {
	case gameitem.ItemGold:
		s.AddMessage(fmt.Sprintf("You found %d gold pieces", item.Value))
	case gameitem.ItemAmulet:
		s.AddMessage(fmt.Sprintf("You picked up the %s!", displayName))
	default:
		s.AddMessage(fmt.Sprintf("You picked up %s", displayName))
	}

	// Remove item from level
	s.level.RemoveItem(item)
}

// enterUseMode enters the use/apply mode
func (s *GameScreen) enterUseMode() {
	// For now, provide a message about what this will do
	s.AddMessage("Use what? (not fully implemented yet)")
	// TODO: Implement use mode for rings, wands, etc.
}

// handleWait handles the wait/rest command
func (s *GameScreen) handleWait() {
	s.AddMessage("You rest.")
	// Let monsters take their turn
	s.level.UpdateMonsters(s.player)
}

// handleSearch handles searching for hidden doors and traps
func (s *GameScreen) handleSearch() {
	s.AddMessage("You search the area.")
	// TODO: Implement search functionality for hidden doors/traps
	// For now, just let monsters take their turn
	s.level.UpdateMonsters(s.player)
}

// handleOpenDoor handles opening doors
func (s *GameScreen) handleOpenDoor() {
	s.AddMessage("Which direction? (not implemented yet)")
	// TODO: Implement door opening functionality
}

// handleCloseDoor handles closing doors
func (s *GameScreen) handleCloseDoor() {
	s.AddMessage("Which direction? (not implemented yet)")
	// TODO: Implement door closing functionality
}

// canGoDownstairs checks if the player can go down stairs
func (s *GameScreen) canGoDownstairs() bool {
	if s.dungeonManager == nil {
		return false
	}
	return s.dungeonManager.CanGoDownstairs()
}
