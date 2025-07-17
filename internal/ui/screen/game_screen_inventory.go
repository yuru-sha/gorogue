package screen

import (
	"fmt"

	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
)

// showInventory displays the player's inventory
func (s *GameScreen) showInventory() {
	listing := s.player.Inventory.GetInventoryListing(s.player.IdentifyMgr)
	for _, line := range listing {
		s.AddMessage(line)
	}
}

// enterEquipMode enters equipment selection mode
func (s *GameScreen) enterEquipMode() {
	if s.player.Inventory.IsEmpty() {
		s.AddMessage("You have nothing to equip.")
		return
	}

	// 装備可能なアイテムをリストアップ
	s.equippableItems = make([]*gameitem.Item, 0)
	for _, item := range s.player.Inventory.Items {
		if s.canEquip(item) {
			s.equippableItems = append(s.equippableItems, item)
		}
	}

	if len(s.equippableItems) == 0 {
		s.AddMessage("You have no equippable items.")
		return
	}

	s.inputMode = ModeEquip
	s.showEquipMenu()
}

// showEquipMenu shows the equip item menu
func (s *GameScreen) showEquipMenu() {
	s.AddMessage("Equippable items:")
	for i, item := range s.equippableItems {
		letter := rune('a' + i)
		displayName := s.player.IdentifyMgr.GetDisplayName(item)
		s.AddMessage(fmt.Sprintf("%c) %s", letter, displayName))
	}
	s.AddMessage("Equip which item? (a-z, ESC to cancel)")
}

// enterUnequipMode enters unequip selection mode
func (s *GameScreen) enterUnequipMode() {
	// 現在装備しているアイテムがあるかチェック
	if s.player.Equipment.Weapon == nil && s.player.Equipment.Armor == nil &&
		s.player.Equipment.RingLeft == nil && s.player.Equipment.RingRight == nil {
		s.AddMessage("You have nothing equipped to take off.")
		return
	}

	s.inputMode = ModeUnequip
	s.showUnequipMenu()
}

// showUnequipMenu shows the unequip item menu
func (s *GameScreen) showUnequipMenu() {
	// 現在装備しているアイテムをチェック
	equippedItems := make([]string, 0)

	if s.player.Equipment.Weapon != nil {
		weaponName := s.player.IdentifyMgr.GetDisplayName(s.player.Equipment.Weapon)
		equippedItems = append(equippedItems, fmt.Sprintf("(w) %s", weaponName))
	}
	if s.player.Equipment.Armor != nil {
		armorName := s.player.IdentifyMgr.GetDisplayName(s.player.Equipment.Armor)
		equippedItems = append(equippedItems, fmt.Sprintf("(a) %s", armorName))
	}
	if s.player.Equipment.RingLeft != nil {
		ringName := s.player.IdentifyMgr.GetDisplayName(s.player.Equipment.RingLeft)
		equippedItems = append(equippedItems, fmt.Sprintf("(l) %s", ringName))
	}
	if s.player.Equipment.RingRight != nil {
		ringName := s.player.IdentifyMgr.GetDisplayName(s.player.Equipment.RingRight)
		equippedItems = append(equippedItems, fmt.Sprintf("(r) %s", ringName))
	}

	if len(equippedItems) == 0 {
		s.AddMessage("You have nothing equipped to take off.")
		return
	}

	s.AddMessage("Currently equipped:")
	for _, item := range equippedItems {
		s.AddMessage(item)
	}
	s.AddMessage("Take off which item? (w)eapon, (a)rmor, (l)eft ring, (r)ight ring")
}

// enterDropMode enters drop selection mode
func (s *GameScreen) enterDropMode() {
	if s.player.Inventory.IsEmpty() {
		s.AddMessage("You have nothing to drop.")
		return
	}

	s.inputMode = ModeDrop
	s.showDropMenu()
}

// showDropMenu shows the drop item menu
func (s *GameScreen) showDropMenu() {
	listing := s.player.Inventory.GetInventoryListing(s.player.IdentifyMgr)
	for _, line := range listing {
		s.AddMessage(line)
	}
	s.AddMessage("Drop which item? (a-z, ESC to cancel)")
}

// enterQuaffMode enters potion quaffing mode
func (s *GameScreen) enterQuaffMode() {
	if s.player.Inventory.IsEmpty() {
		s.AddMessage("You have no potions to drink.")
		return
	}

	// ポーションをリストアップ
	potions := make([]*gameitem.Item, 0)
	for _, item := range s.player.Inventory.Items {
		if item.Type == gameitem.ItemPotion {
			potions = append(potions, item)
		}
	}

	if len(potions) == 0 {
		s.AddMessage("You have no potions to drink.")
		return
	}

	s.inputMode = ModeQuaff
	s.showPotions()
}

// showPotions displays available potions
func (s *GameScreen) showPotions() {
	s.AddMessage("Available potions:")
	index := 0
	for i, item := range s.player.Inventory.Items {
		if item.Type == gameitem.ItemPotion {
			letter := rune('a' + i)
			displayName := s.player.IdentifyMgr.GetDisplayName(item)
			s.AddMessage(fmt.Sprintf("%c) %s", letter, displayName))
			index++
		}
	}
	s.AddMessage("Quaff which potion? (a-z, ESC to cancel)")
}

// enterReadMode enters scroll reading mode
func (s *GameScreen) enterReadMode() {
	if s.player.Inventory.IsEmpty() {
		s.AddMessage("You have no scrolls to read.")
		return
	}

	// 巻物をリストアップ
	scrolls := make([]*gameitem.Item, 0)
	for _, item := range s.player.Inventory.Items {
		if item.Type == gameitem.ItemScroll {
			scrolls = append(scrolls, item)
		}
	}

	if len(scrolls) == 0 {
		s.AddMessage("You have no scrolls to read.")
		return
	}

	s.inputMode = ModeRead
	s.showScrolls()
}

// showScrolls displays available scrolls
func (s *GameScreen) showScrolls() {
	s.AddMessage("Available scrolls:")
	index := 0
	for i, item := range s.player.Inventory.Items {
		if item.Type == gameitem.ItemScroll {
			letter := rune('a' + i)
			displayName := s.player.IdentifyMgr.GetDisplayName(item)
			s.AddMessage(fmt.Sprintf("%c) %s", letter, displayName))
			index++
		}
	}
	s.AddMessage("Read which scroll? (a-z, ESC to cancel)")
}

// enterCLIMode enters CLI debug mode
func (s *GameScreen) enterCLIMode() {
	if s.cliMode == nil {
		s.AddMessage("CLI mode not available.")
		return
	}

	s.cliMode.IsActive = true
	s.inputMode = ModeCLI
	s.cliBuffer = ""
	s.AddMessage("Entered CLI mode. Type 'help' for commands, ESC to exit.")
	s.AddMessage("CLI> ")
}

// canEquip checks if an item can be equipped
func (s *GameScreen) canEquip(item *gameitem.Item) bool {
	switch item.Type {
	case gameitem.ItemWeapon, gameitem.ItemArmor, gameitem.ItemRing:
		return true
	default:
		return false
	}
}
