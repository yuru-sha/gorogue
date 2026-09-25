package screen

import (
	"fmt"

	"github.com/yuru-sha/gorogue/internal/core/command"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
)

// showInventory displays the player's inventory
func (s *GameScreen) showInventory() {
	s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdInventory}))
}

func (s *GameScreen) enterEquipModeFor(action command.Type) {
	s.equipAction = action
	if s.player.Inventory.IsEmpty() {
		s.AddMessage("You have nothing to equip.")
		return
	}

	s.equippableItems = make([]*gameitem.Item, 0)
	for _, candidate := range s.player.Inventory.Items {
		if s.canEquip(candidate) && fitsEquipAction(candidate, action) {
			s.equippableItems = append(s.equippableItems, candidate)
		}
	}
	if len(s.equippableItems) == 0 {
		s.AddMessage("You have no matching equipment.")
		return
	}

	s.inputMode = ModeEquip
	s.showEquipMenu()
}

func fitsEquipAction(candidate *gameitem.Item, action command.Type) bool {
	switch action {
	case command.CmdWield:
		return candidate.Type == gameitem.ItemWeapon
	case command.CmdWear:
		return candidate.Type == gameitem.ItemArmor
	case command.CmdRingOn:
		return candidate.Type == gameitem.ItemRing
	default:
		return true
	}
}

// showEquipMenu shows the equip item menu
func (s *GameScreen) showEquipMenu() {
	s.AddMessage("Equippable items:")
	for i, item := range s.equippableItems {
		letter := rune('a' + i)
		displayName := s.player.IdentifyMgr.GetDisplayName(item)
		s.AddMessage(fmt.Sprintf("%c) %s", letter, displayName))
	}
	s.AddMessage("Select an item (a-z, ESC to cancel)")
}

func (s *GameScreen) enterUnequipModeFor(action command.Type) {
	s.unequipAction = action
	if s.player.Equipment.Weapon == nil && s.player.Equipment.Armor == nil &&
		s.player.Equipment.RingLeft == nil && s.player.Equipment.RingRight == nil {
		s.AddMessage("You have nothing equipped to take off.")
		return
	}
	if action == command.CmdTakeOff && s.player.Equipment.Armor == nil {
		s.AddMessage("You are not wearing armor.")
		return
	}
	if action == command.CmdRingOff &&
		s.player.Equipment.RingLeft == nil && s.player.Equipment.RingRight == nil {
		s.AddMessage("You are not wearing a ring.")
		return
	}
	s.inputMode = ModeUnequip
	s.showUnequipMenu()
}

func (s *GameScreen) showUnequipMenu() {
	var prompt string
	switch s.unequipAction {
	case command.CmdTakeOff:
		prompt = "(a)rmor"
	case command.CmdRingOff:
		prompt = "(l)eft ring, (r)ight ring"
	default:
		prompt = "(w)eapon, (a)rmor, (l)eft ring, (r)ight ring"
	}
	s.AddMessage("Take off which item? " + prompt)
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

func (s *GameScreen) showIdentifyTargets(types []gameitem.ItemType) {
	s.AddMessage("Identify which item?")
	shown := 0
	for index, candidate := range s.player.Inventory.Items {
		for _, targetType := range types {
			if candidate.Type != targetType {
				continue
			}
			name := s.player.IdentifyMgr.GetDisplayName(candidate)
			s.AddMessage(fmt.Sprintf("%c) %s", rune('a'+index), name))
			shown++
			break
		}
	}
	if shown == 0 {
		s.AddMessage("You have no matching items. ESC to cancel.")
	}
}

// enterEatMode enters food selection mode.
func (s *GameScreen) enterEatMode() {
	for _, item := range s.player.Inventory.Items {
		if item.Type == gameitem.ItemFood {
			s.inputMode = ModeEat
			s.showFood()
			return
		}
	}
	s.AddMessage("You have no food to eat.")
}

// showFood displays available food.
func (s *GameScreen) showFood() {
	s.AddMessage("Available food:")
	for i, item := range s.player.Inventory.Items {
		if item.Type == gameitem.ItemFood {
			letter := rune('a' + i)
			s.AddMessage(fmt.Sprintf("%c) %s", letter, s.player.IdentifyMgr.GetDisplayName(item)))
		}
	}
	s.AddMessage("Eat which food? (a-z, ESC to cancel)")
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
