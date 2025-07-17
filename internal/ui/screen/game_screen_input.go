package screen

import (
	"fmt"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/state"
	gameitem "github.com/yuru-sha/gorogue/internal/game/item"
	"github.com/yuru-sha/gorogue/internal/game/magic"
	"github.com/yuru-sha/gorogue/internal/utils/logger"
)

// HandleInput handles input events
func (s *GameScreen) HandleInput(msg gruid.Msg) state.GameState {
	switch msg := msg.(type) {
	case gruid.MsgKeyDown:
		// モード別の処理
		switch s.inputMode {
		case ModeEquip:
			return s.handleEquipInput(msg.Key)
		case ModeUnequip:
			return s.handleUnequipInput(msg.Key)
		case ModeDrop:
			return s.handleDropInput(msg.Key)
		case ModeQuaff:
			return s.handleQuaffInput(msg.Key)
		case ModeRead:
			return s.handleReadInput(msg.Key)
		case ModeCLI:
			return s.handleCLIInput(msg.Key)
		default: // ModeNormal
			return s.handleNormalInput(msg.Key)
		}
	}
	return state.StateGame
}

// handleNormalInput handles input in normal mode
func (s *GameScreen) handleNormalInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		logger.Info("Returning to menu")
		return state.StateMenu
	case "Left", "h", gruid.KeyArrowLeft:
		s.tryMovePlayer(-1, 0)
	case "Right", "l", gruid.KeyArrowRight:
		s.tryMovePlayer(1, 0)
	case "Up", "k", gruid.KeyArrowUp:
		s.tryMovePlayer(0, -1)
	case "Down", "j", gruid.KeyArrowDown:
		s.tryMovePlayer(0, 1)
	case "y":
		s.tryMovePlayer(-1, -1)
	case "u":
		s.tryMovePlayer(1, -1)
	case "b":
		s.tryMovePlayer(-1, 1)
	case "n":
		s.tryMovePlayer(1, 1)
	case "Q":
		logger.Info("Quit requested")
		return state.StateMenu
	case "<", ",": // Go upstairs
		s.handleStairs(true)
	case ">", ".": // Go downstairs
		s.handleStairs(false)
	case "i": // Show inventory
		s.showInventory()
	case "d": // Drop item
		s.enterDropMode()
	case "w": // Wield/wear item
		s.enterEquipMode()
	case "t": // Take off item
		s.enterUnequipMode()
	case "q": // Quaff potion
		s.enterQuaffMode()
	case "r": // Read scroll
		s.enterReadMode()
	case ":": // Enter CLI mode
		s.enterCLIMode()
	case "^W", "W": // Ctrl+W or Shift+W to toggle wizard mode
		s.wizardMode.Toggle()
		status := "OFF"
		if s.wizardMode.IsActive {
			status = "ON"
		}
		s.AddMessage(fmt.Sprintf("ウィザードモード: %s", status))
	default:
		// Check if it's a wizard command
		if s.wizardMode.IsActive && len(string(key)) == 1 {
			result := s.wizardMode.ExecuteCommand(rune(string(key)[0]))
			if result != "" {
				s.AddMessage(result)
			}
		}
	}
	return state.StateGame
}

// handleStairs handles stair movement
func (s *GameScreen) handleStairs(goUp bool) {
	if s.dungeonManager == nil {
		s.AddMessage("Dungeon manager not available")
		return
	}

	if goUp {
		if s.dungeonManager.CanGoUpstairs() {
			if s.dungeonManager.GoUpstairs() {
				s.level = s.dungeonManager.GetCurrentLevel()
				s.wizardMode.SetLevel(s.level)
				s.AddMessage(fmt.Sprintf("階層 %d へ上がった", s.dungeonManager.GetCurrentFloor()))
			}
		} else {
			s.AddMessage("ここには上り階段がない")
		}
	} else {
		if s.dungeonManager.CanGoDownstairs() {
			if s.dungeonManager.GoDownstairs() {
				s.level = s.dungeonManager.GetCurrentLevel()
				s.wizardMode.SetLevel(s.level)
				s.AddMessage(fmt.Sprintf("階層 %d へ下りた", s.dungeonManager.GetCurrentFloor()))

				// 最終階層に到達した場合、イェンダーの魔除けを配置
				if s.dungeonManager.IsOnFinalFloor() {
					s.dungeonManager.PlaceAmuletOfYendor()
					s.AddMessage("この階層には強力な魔力を感じる...")
				}
			}
		} else {
			s.AddMessage("ここには下り階段がない")
		}
	}
}

// handleEquipInput handles input in equip mode
func (s *GameScreen) handleEquipInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			index := int(string(key)[0] - 'a')
			if index < len(s.equippableItems) {
				item := s.equippableItems[index]
				if s.player.Equipment.EquipItem(item) {
					// インベントリからアイテムを削除
					for i, invItem := range s.player.Inventory.Items {
						if invItem == item {
							s.player.Inventory.RemoveItem(i)
							break
						}
					}
					displayName := s.player.IdentifyMgr.GetDisplayName(item)
					s.AddMessage(fmt.Sprintf("You equipped %s.", displayName))
				} else {
					s.AddMessage("You can't equip that item.")
				}
			} else {
				s.AddMessage("Invalid selection.")
			}
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleUnequipInput handles input in unequip mode
func (s *GameScreen) handleUnequipInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	case "w": // Unequip weapon
		s.unequipSlot("weapon", "weapon")
	case "a": // Unequip armor
		s.unequipSlot("armor", "armor")
	case "l": // Unequip left ring
		s.unequipSlot("ring_left", "left ring")
	case "r": // Unequip right ring
		s.unequipSlot("ring_right", "right ring")
	default:
		s.AddMessage("Invalid selection. Use (w)eapon, (a)rmor, (l)eft ring, (r)ight ring")
	}
	s.inputMode = ModeNormal
	return state.StateGame
}

// unequipSlot unequips an item from a specific slot
func (s *GameScreen) unequipSlot(slot, displaySlot string) {
	if item := s.player.Equipment.UnequipItem(slot); item != nil {
		if s.player.Inventory.AddItem(item) {
			displayName := s.player.IdentifyMgr.GetDisplayName(item)
			s.AddMessage(fmt.Sprintf("You took off %s.", displayName))
		} else {
			s.AddMessage("Your pack is full!")
			// 装備を戻す
			s.player.Equipment.EquipItem(item)
		}
	} else {
		s.AddMessage(fmt.Sprintf("You have no %s equipped.", displaySlot))
	}
}

// handleDropInput handles input in drop mode
func (s *GameScreen) handleDropInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			index := int(string(key)[0] - 'a')
			if item := s.player.Inventory.GetItem(index); item != nil {
				displayName := s.player.IdentifyMgr.GetDisplayName(item)
				s.AddMessage(fmt.Sprintf("You dropped %s.", displayName))
				// アイテムをプレイヤーの位置に配置
				s.level.AddItem(item, s.player.Position.X, s.player.Position.Y)
				s.player.Inventory.RemoveItem(index)
			} else {
				s.AddMessage("Invalid selection.")
			}
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleQuaffInput handles input in quaff mode
func (s *GameScreen) handleQuaffInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			index := int(string(key)[0] - 'a')
			if item := s.player.Inventory.GetItem(index); item != nil {
				if item.Type == gameitem.ItemPotion {
					result := magic.UsePotion(item.Name, s.player)
					s.AddMessage(result.Message)

					if result.Identified {
						s.player.IdentifyMgr.IdentifyByUse(item)
					}

					// ポーションを消費
					s.player.Inventory.RemoveItem(index)
				} else {
					s.AddMessage("You can't drink that!")
				}
			} else {
				s.AddMessage("Invalid selection.")
			}
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleReadInput handles input in read mode
func (s *GameScreen) handleReadInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			index := int(string(key)[0] - 'a')
			if item := s.player.Inventory.GetItem(index); item != nil {
				if item.Type == gameitem.ItemScroll {
					result := magic.UseScroll(item.Name, s.player, s.level)
					s.AddMessage(result.Message)

					if result.Identified {
						s.player.IdentifyMgr.IdentifyByUse(item)
					}

					// 巻物を消費
					s.player.Inventory.RemoveItem(index)
				} else {
					s.AddMessage("You can't read that!")
				}
			} else {
				s.AddMessage("Invalid selection.")
			}
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleCLIInput handles input in CLI mode
func (s *GameScreen) handleCLIInput(key gruid.Key) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.cliBuffer = ""
		s.AddMessage("CLI mode exited.")
		return state.StateGame
	case gruid.KeyEnter:
		if s.cliBuffer != "" {
			// Execute command
			result := s.cliMode.ExecuteCommand(s.cliBuffer)
			s.AddMessage(fmt.Sprintf("> %s", s.cliBuffer))
			s.AddMessage(result)

			// Add to history
			s.cliHistory = append(s.cliHistory, s.cliBuffer)
			if len(s.cliHistory) > 20 {
				s.cliHistory = s.cliHistory[1:]
			}

			s.cliBuffer = ""
		}
		s.inputMode = ModeNormal
		return state.StateGame
	case gruid.KeyBackspace:
		if s.cliBuffer != "" {
			s.cliBuffer = s.cliBuffer[:len(s.cliBuffer)-1]
		}
		return state.StateGame
	default:
		// Add character to buffer
		if len(string(key)) == 1 {
			char := string(key)[0]
			if char >= 32 && char <= 126 { // Printable ASCII
				s.cliBuffer += string(char)
			}
		}
		return state.StateGame
	}
}
