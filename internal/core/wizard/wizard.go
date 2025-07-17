package wizard

import (
	"fmt"

	"github.com/yuru-sha/gorogue/internal/game/actor"
	"github.com/yuru-sha/gorogue/internal/game/dungeon"
	"github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// WizardMode handles wizard mode functionality
type WizardMode struct {
	IsActive bool
	Level    *dungeon.Level
	Player   *actor.Player
}

// NewWizardMode creates a new wizard mode instance
func NewWizardMode(level *dungeon.Level, player *actor.Player) *WizardMode {
	return &WizardMode{
		IsActive: false,
		Level:    level,
		Player:   player,
	}
}

// Toggle toggles wizard mode on/off
func (w *WizardMode) Toggle() {
	w.IsActive = !w.IsActive
	status := "OFF"
	if w.IsActive {
		status = "ON"
	}
	logger.Info("Wizard mode toggled", "status", status)
}

// SetLevel updates the level reference for wizard mode
func (w *WizardMode) SetLevel(level *dungeon.Level) {
	w.Level = level
	logger.Debug("Wizard mode level updated")
}

// ExecuteCommand executes a wizard command
func (w *WizardMode) ExecuteCommand(command rune) string {
	if !w.IsActive {
		return ""
	}

	switch command {
	case 'h': // Help
		return w.showHelp()
	case 'g': // Grant gold
		return w.grantGold()
	case 'l': // Level up
		return w.levelUp()
	case 'r': // Full heal
		return w.fullHeal()
	case 'f': // Full food
		return w.fullFood()
	case 'k': // Kill all monsters
		return w.killAllMonsters()
	case 's': // Show stats
		return w.showStats()
	case 'i': // Create item
		return w.createItem()
	case 't': // Teleport
		return w.teleport()
	case 'v': // Toggle visibility
		return w.toggleVisibility()
	case 'w': // Walk through walls
		return w.toggleWalkThroughWalls()
	default:
		return "不明なウィザードコマンド"
	}
}

// showHelp displays wizard mode help
func (w *WizardMode) showHelp() string {
	return "ウィザードモード: h=ヘルプ g=ゴールド l=レベルアップ r=回復 f=満腹 k=全モンスター撃破 s=ステータス i=アイテム作成 t=テレポート v=視界切替 w=壁通り抜け"
}

// grantGold grants gold to the player
func (w *WizardMode) grantGold() string {
	w.Player.AddGold(1000)
	logger.Info("Wizard: Granted gold", "amount", 1000)
	return "1000ゴールドを取得しました"
}

// levelUp levels up the player
func (w *WizardMode) levelUp() string {
	w.Player.GainExp(w.Player.GetExpToNextLevel())
	logger.Info("Wizard: Player leveled up")
	return "レベルアップしました"
}

// fullHeal fully heals the player
func (w *WizardMode) fullHeal() string {
	w.Player.HP = w.Player.MaxHP
	logger.Info("Wizard: Player fully healed")
	return "完全回復しました"
}

// fullFood fills the player's hunger
func (w *WizardMode) fullFood() string {
	w.Player.Hunger = 100
	logger.Info("Wizard: Player hunger restored")
	return "満腹になりました"
}

// killAllMonsters kills all monsters in the level
func (w *WizardMode) killAllMonsters() string {
	count := 0
	for _, monster := range w.Level.Monsters {
		if monster.IsAlive() {
			monster.HP = 0
			count++
		}
	}
	logger.Info("Wizard: Killed all monsters", "count", count)
	return fmt.Sprintf("すべてのモンスター（%d体）を倒しました", count)
}

// showStats shows detailed player statistics
func (w *WizardMode) showStats() string {
	return fmt.Sprintf("Lv:%d HP:%d/%d Atk:%d Def:%d Exp:%d Gold:%d Hunger:%d",
		w.Player.Level, w.Player.HP, w.Player.MaxHP,
		w.Player.Attack, w.Player.Defense, w.Player.Exp,
		w.Player.Gold, w.Player.Hunger)
}

// createItem creates a random item at player's position
func (w *WizardMode) createItem() string {
	items := []item.ItemType{
		item.ItemWeapon, item.ItemArmor, item.ItemRing,
		item.ItemScroll, item.ItemPotion, item.ItemFood,
		item.ItemGold, item.ItemAmulet,
	}

	itemType := items[len(items)-1] // Create amulet for testing
	var newItem *item.Item

	switch itemType {
	case item.ItemGold:
		newItem = item.NewGold(w.Player.Position.X, w.Player.Position.Y, false)
	case item.ItemAmulet:
		newItem = item.NewAmulet(w.Player.Position.X, w.Player.Position.Y)
	default:
		newItem = item.NewItem(w.Player.Position.X, w.Player.Position.Y, itemType, "ウィザードアイテム", 100)
	}

	w.Level.Items = append(w.Level.Items, newItem)
	logger.Info("Wizard: Created item", "type", newItem.Name, "x", newItem.Position.X, "y", newItem.Position.Y)
	return fmt.Sprintf("%sを作成しました", newItem.Name)
}

// teleport teleports player to a random room
func (w *WizardMode) teleport() string {
	if len(w.Level.Rooms) == 0 {
		return "テレポート先がありません"
	}

	room := w.Level.Rooms[len(w.Level.Rooms)-1] // Last room
	newX := room.X + room.Width/2
	newY := room.Y + room.Height/2

	w.Player.Position.X = newX
	w.Player.Position.Y = newY

	logger.Info("Wizard: Player teleported", "x", newX, "y", newY)
	return "テレポートしました"
}

// toggleVisibility toggles all tiles visibility
func (w *WizardMode) toggleVisibility() string {
	// Toggle visibility of all tiles
	for y := 0; y < w.Level.Height; y++ {
		for x := 0; x < w.Level.Width; x++ {
			tile := w.Level.GetTile(x, y)
			if tile != nil {
				tile.Visible = !tile.Visible
			}
		}
	}
	logger.Info("Wizard: Toggled visibility")
	return "視界を切り替えました"
}

// toggleWalkThroughWalls toggles walk through walls ability
func (w *WizardMode) toggleWalkThroughWalls() string {
	// This would need to be implemented in the game logic
	// For now, just return a message
	logger.Info("Wizard: Wall walking toggle requested")
	return "壁通り抜けモード（未実装）"
}
