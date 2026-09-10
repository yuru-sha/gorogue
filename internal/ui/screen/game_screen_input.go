package screen

import (
	"fmt"

	"github.com/anaseto/gruid"
	"github.com/yuru-sha/gorogue/internal/core/command"
	"github.com/yuru-sha/gorogue/internal/core/state"
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
		case ModeUse:
			return s.handleUseInput(msg.Key)
		case ModeQuaff:
			return s.handleQuaffInput(msg.Key)
		case ModeRead:
			return s.handleReadInput(msg.Key)
		case ModeEat:
			return s.handleEatInput(msg.Key)
		case ModeCLI:
			return s.handleCLIInput(msg.Key)
		case ModeDirection:
			return s.handleDirectionInput(msg.Key)
		default: // ModeNormal
			return s.handleNormalInput(msg.Key)
		}
	}
	return state.StateGame
}

// handleNormalInput handles input in normal mode
func (s *GameScreen) handleNormalInput(key gruid.Key) state.GameState {
	// Parse the key into a command
	cmd := s.cmdParser.Parse(key)

	switch cmd.Type {
	// Movement commands
	case command.CmdMoveWest, command.CmdMoveEast, command.CmdMoveNorth, command.CmdMoveSouth,
		command.CmdMoveNorthWest, command.CmdMoveNorthEast, command.CmdMoveSouthWest, command.CmdMoveSouthEast:
		s.addCommandResult(s.executeCommand(cmd))

	// Action commands
	case command.CmdLook:
		s.handleLook()
	case command.CmdInventory:
		s.showInventory()
	case command.CmdPickUp:
		s.handlePickUp()
	case command.CmdDrop:
		s.enterDropMode()
	case command.CmdUse:
		s.enterUseMode() // PyRogue unified use interface
	case command.CmdQuaff:
		s.enterQuaffMode()
	case command.CmdRead:
		s.enterReadMode()
	case command.CmdWield, command.CmdEquip:
		s.enterEquipMode()
	case command.CmdTakeOff, command.CmdUnequip:
		s.enterUnequipMode()
	case command.CmdEat:
		s.enterEatMode()
	case command.CmdWait:
		s.handleWait()
	case command.CmdSearch:
		s.handleSearch()
	case command.CmdOpen:
		s.handleOpenDoor()
	case command.CmdClose:
		s.handleCloseDoor()
	case command.CmdFight:
		s.handleFight()
	case command.CmdDisarm:
		s.handleDisarm()
	// Stair commands
	case command.CmdGoUpstairs:
		return s.handleStairs(true)
	case command.CmdGoDownstairs:
		return s.handleStairs(false)

	// System commands
	case command.CmdQuit:
		logger.Info("Quit requested")
		return state.StateMenu
	case command.CmdHelp:
		return state.StateHelp
	case command.CmdSymbol:
		return state.StateSymbol
	case command.CmdEscape:
		logger.Info("Returning to menu")
		return state.StateMenu
	case command.CmdWizard:
		s.wizardMode.Toggle()
		status := "OFF"
		if s.wizardMode.IsActive {
			status = "ON"
		}
		s.AddMessage(fmt.Sprintf("ウィザードモード: %s", status))
	case command.CmdCLI:
		s.enterCLIMode()

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
func (s *GameScreen) handleStairs(goUp bool) state.GameState {
	commandType := command.CmdGoDownstairs
	if goUp {
		commandType = command.CmdGoUpstairs
	}
	result := s.executeCommand(command.Command{Type: commandType})
	s.addCommandResult(result)
	if result.Victory {
		return state.StateVictory
	}
	return state.StateGame
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
			selection := int(string(key)[0] - 'a')
			if selection >= len(s.equippableItems) {
				s.AddMessage("Invalid selection.")
			} else {
				selected := s.equippableItems[selection]
				for index, item := range s.player.Inventory.Items {
					if item == selected {
						result := s.executeCommand(command.Command{Type: command.CmdEquip}, string(rune('a'+index)))
						s.addCommandResult(result)
						break
					}
				}
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
	result := s.executeCommand(command.Command{Type: command.CmdUnequip}, slot)
	if result.Message == "No item equipped in "+slot+" slot." && displaySlot != slot {
		result.Message = "You have no " + displaySlot + " equipped."
	}
	s.addCommandResult(result)
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
			s.addCommandResult(s.executeCommand(command.Command{Type: command.CmdDrop}, string(key)))
			s.inputMode = ModeNormal
		}
	}
	return state.StateGame
}

// handleQuaffInput handles input in quaff mode
func (s *GameScreen) handleQuaffInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdQuaff)
}

// handleReadInput handles input in read mode
func (s *GameScreen) handleReadInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdRead)
}

// handleEatInput handles food selection.
func (s *GameScreen) handleEatInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdEat)
}

func (s *GameScreen) handleUseInput(key gruid.Key) state.GameState {
	return s.handleItemSelection(key, command.CmdUse)
}

func (s *GameScreen) handleItemSelection(key gruid.Key, commandType command.Type) state.GameState {
	switch key {
	case gruid.KeyEscape:
		s.inputMode = ModeNormal
		s.AddMessage("Canceled.")
		return state.StateGame
	default:
		if len(string(key)) == 1 && string(key)[0] >= 'a' && string(key)[0] <= 'z' {
			s.addCommandResult(s.executeCommand(command.Command{Type: commandType}, string(key)))
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
			s.player = s.cliMode.Player
			s.dungeonManager = s.cliMode.Dungeon
			s.level = s.cliMode.Level
			if s.wizardMode != nil && s.level != nil {
				s.wizardMode.Player = s.player
				s.wizardMode.SetLevel(s.level)
			}

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

// handleDirectionInput handles input in direction mode
func (s *GameScreen) handleDirectionInput(key gruid.Key) state.GameState {
	// Parse the key into a command
	cmd := s.cmdParser.Parse(key)

	switch cmd.Type {
	case command.CmdMoveWest, command.CmdMoveEast, command.CmdMoveNorth, command.CmdMoveSouth,
		command.CmdMoveNorthWest, command.CmdMoveNorthEast, command.CmdMoveSouthWest, command.CmdMoveSouthEast:
		// Call the callback with the direction
		if s.directionCallback != nil {
			s.directionCallback(cmd.Direction.X, cmd.Direction.Y)
		}
		s.inputMode = ModeNormal
		s.directionCallback = nil
	case command.CmdEscape:
		s.AddMessage("Canceled.")
		s.inputMode = ModeNormal
		s.directionCallback = nil
	default:
		s.AddMessage("Invalid direction. Use hjklybnu or arrow keys.")
	}

	return state.StateGame
}
